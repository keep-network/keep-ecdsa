package firewall

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/cache"

	coreChain "github.com/keep-network/keep-core/pkg/chain"
	coreFirewall "github.com/keep-network/keep-core/pkg/firewall"
	coreNet "github.com/keep-network/keep-core/pkg/net"
	coreKey "github.com/keep-network/keep-core/pkg/net/key"

	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
)

var logger = log.Logger("keep-firewall")

const (
	// authorizationCachePeriod it the time the cache maintains information
	// about operators authorized and not authorized for the keep factory.
	// We use the cache to minimize calls to Ethereum client.
	authorizationCachePeriod = 24 * time.Hour

	// activeKeepCachePeriod is the time the cache maintains information
	// about active and no active keep members. We use the cache to minimize
	// calls to Ethereum client.
	activeKeepCachePeriod = 168 * time.Hour // one week
)

var errNoAuthorization = fmt.Errorf("remote peer has no authorization on the factory")

var errNoMinStakeNoActiveKeep = fmt.Errorf("remote peer has no minimum " +
	"stake and is not a member in any of active keeps")

// NewStakeOrActiveKeepPolicy is a firewall policy checking if the remote peer
// has a minimum stake and in case it has no minimum stake if it is a member of
// at least one active keep.
func NewStakeOrActiveKeepPolicy(
	chain eth.Handle,
	stakeMonitor coreChain.StakeMonitor,
) coreNet.Firewall {
	return &stakeOrActiveKeepPolicy{
		chain:                       chain,
		minimumStakePolicy:          coreFirewall.MinimumStakePolicy(stakeMonitor),
		authorizedOperatorsCache:    cache.NewTimeCache(authorizationCachePeriod),
		nonAuthorizedOperatorsCache: cache.NewTimeCache(authorizationCachePeriod),
		activeKeepMembersCache:      cache.NewTimeCache(activeKeepCachePeriod),
		noActiveKeepMembersCache:    cache.NewTimeCache(activeKeepCachePeriod),
		keepInfoCache:               newKeepInfoCache(),
	}
}

type stakeOrActiveKeepPolicy struct {
	chain                       eth.Handle
	minimumStakePolicy          coreNet.Firewall
	authorizedOperatorsCache    *cache.TimeCache
	nonAuthorizedOperatorsCache *cache.TimeCache
	activeKeepMembersCache      *cache.TimeCache
	noActiveKeepMembersCache    *cache.TimeCache
	keepInfoCache               *keepInfoCache
}

func (soakp *stakeOrActiveKeepPolicy) Validate(
	remotePeerPublicKey *ecdsa.PublicKey,
) error {
	remotePeerNetworkPublicKey := coreKey.NetworkPublic(*remotePeerPublicKey)
	remotePeerAddress := coreKey.NetworkPubKeyToEthAddress(&remotePeerNetworkPublicKey)

	logger.Infof("validating remote peer: %s", remotePeerAddress)

	// Validate minimum stake policy. If the remote peer has the minimum stake,
	// we are fine and we should let to connect.
	if err := soakp.minimumStakePolicy.Validate(remotePeerPublicKey); err == nil {
		return nil
	}

	// Check if the remote peer has authorization on the factory.
	// The authorization cannot be revoked.
	// If peer has no authorization on the factory it means it has never
	// participated in any group selection so there is no chance it can be
	// a member of any keep.
	err := soakp.validateAuthorization(remotePeerAddress)
	if err != nil {
		return err
	}

	// In case the remote peer has no minimum stake, we need to check if it is
	// a member in at least one active keep. If so, we let to connect.
	// Otherwise, we do not let to connect.
	return soakp.validateActiveKeepMembership(remotePeerAddress)
}

func (soakp *stakeOrActiveKeepPolicy) validateAuthorization(
	remotePeerAddress string,
) error {
	// Before hitting ETH client, consult the in-memory time cache.
	// If the caching time for the given entry elapsed or if that entry is
	// not in the cache, we'll have to consult the chain and execute a call
	// to ETH client.
	soakp.authorizedOperatorsCache.Sweep()
	soakp.nonAuthorizedOperatorsCache.Sweep()

	if soakp.authorizedOperatorsCache.Has(remotePeerAddress) {
		return nil
	}

	if soakp.nonAuthorizedOperatorsCache.Has(remotePeerAddress) {
		return errNoAuthorization
	}

	// We do not know if the remote peer has or has not the authorization so
	// we need to ask ETH client about it.
	isAuthorized, err := soakp.chain.IsOperatorAuthorized(
		common.HexToAddress(remotePeerAddress),
	)
	if err != nil {
		return fmt.Errorf(
			"could not check authorization for address [%v]: [%v]",
			remotePeerAddress,
			err,
		)
	}

	if !isAuthorized {
		soakp.nonAuthorizedOperatorsCache.Add(remotePeerAddress)
		return errNoAuthorization
	}

	soakp.authorizedOperatorsCache.Add(remotePeerAddress)
	return nil
}

func (soakp *stakeOrActiveKeepPolicy) validateActiveKeepMembership(
	remotePeerAddress string,
) error {

	// First, check in the in-memory time cache to minimize hits to ETH client.
	// If the Keep client with the given chain address is in the active members
	// cache it means it's been a member in at least one active keep the last time
	// validateActiveKeepMembership was executed and caching period has not
	// elapsed yet. Similarly, if the client is in the no active keep members
	// cache it means it hasn't been a member of any active keep during the last check.
	//
	// If the caching period elapsed, this check will return false and we
	// have to ask the chain about the current status.
	soakp.activeKeepMembersCache.Sweep()
	soakp.noActiveKeepMembersCache.Sweep()

	if soakp.activeKeepMembersCache.Has(remotePeerAddress) {
		return nil
	}

	if soakp.noActiveKeepMembersCache.Has(remotePeerAddress) {
		return errNoMinStakeNoActiveKeep
	}

	zero := big.NewInt(0)
	one := big.NewInt(1)

	// Start iterating through all keeps known to the factory starting from the
	// ones most recently created as there is a higher chance they are active.
	keepCount, err := soakp.chain.GetKeepCount()
	if err != nil {
		return fmt.Errorf("could not get keep count: [%v]", err)
	}

	lastIndex := new(big.Int).Sub(keepCount, one)

	for keepIndex := new(big.Int).Set(lastIndex); keepIndex.Cmp(zero) != -1; keepIndex.Sub(keepIndex, one) {
		keepAddress, err := soakp.chain.GetKeepAtIndex(keepIndex)
		if err != nil {
			logger.Errorf(
				"could not get keep at index [%v]: [%v]",
				keepIndex,
				err,
			)
			continue
		}

		// We are interested only in active keeps. If the current keep is not
		// active, we skip it. We still need to process the rest of the keeps
		// because it's possible that although this keep is not active some
		// peers created before this one are still active.
		isActive, err := soakp.isKeepActive(keepAddress)
		if err != nil {
			logger.Errorf(
				"could not check if keep [%x] is active: [%v]",
				keepAddress,
				err,
			)
			continue
		}
		if !isActive {
			continue
		}

		// Get all the members of the active keep and store them in the active
		// keep members cache.
		members, err := soakp.getKeepMembers(keepAddress)
		if err != nil {
			logger.Errorf(
				"could not get members of keep [%x]: [%v]",
				keepAddress,
				err,
			)
			continue
		}
		for _, member := range members {
			soakp.activeKeepMembersCache.Add(member)
		}

		// If the remote peer address has been added to the cache we can
		// connect with this client as it's a member of an active keep.
		if soakp.activeKeepMembersCache.Has(remotePeerAddress) {
			return nil
		}
	}

	soakp.noActiveKeepMembersCache.Add(remotePeerAddress)

	// If we are here, it means that the client is not a member in any of
	// active keeps and it's minimum stake check failed as well. We are not
	// allowing to connect with that peer.
	return errNoMinStakeNoActiveKeep
}

// isKeepActive performs on-chain check whether the keep with the given address
// is active if the keep has not been previously marked as inactive in the cache.
// If the keep has been marked as inactive in the cache, function returns false
// without hitting the chain.
func (soakp *stakeOrActiveKeepPolicy) isKeepActive(
	keepAddress common.Address,
) (bool, error) {
	cache := soakp.keepInfoCache

	cache.mutex.RLock()
	isInactive, isCached := cache.isInactive[keepAddress.String()]
	cache.mutex.RUnlock()

	if isCached && isInactive {
		return false, nil
	}

	isActive, err := soakp.chain.IsActive(keepAddress)
	if err != nil {
		return false, err
	}

	if !isActive {
		cache.mutex.Lock()
		cache.isInactive[keepAddress.String()] = true
		cache.mutex.Unlock()
	}

	return isActive, nil
}

// getKeepMembers fetches members of the keep with the given address from the
// chain or reads them from a cache if this information is available there.
func (soakp *stakeOrActiveKeepPolicy) getKeepMembers(
	keepAddress common.Address,
) ([]string, error) {
	cache := soakp.keepInfoCache

	cache.mutex.RLock()
	members, areCached := cache.members[keepAddress.String()]
	cache.mutex.RUnlock()

	if areCached {
		return members, nil
	}

	memberAddresses, err := soakp.chain.GetMembers(keepAddress)
	if err != nil {
		return nil, nil
	}

	members = make([]string, len(memberAddresses))
	for i, member := range memberAddresses {
		members[i] = member.String()
	}

	cache.mutex.Lock()
	cache.members[keepAddress.String()] = members
	cache.mutex.Unlock()

	return members, nil
}

// keepInfoCache caches invariant information obtained from the chain.
// This cache never expires.
//
// There are two invariants that can be cached:
// 1. Information whether the keep is inactive. Inactive keep can never become
//    active again.
// 2. Information about keep members. Keep members never change.
type keepInfoCache struct {
	isInactive map[string]bool
	members    map[string][]string
	mutex      sync.RWMutex
}

func newKeepInfoCache() *keepInfoCache {
	return &keepInfoCache{
		isInactive: make(map[string]bool),
		members:    make(map[string][]string),
	}
}
