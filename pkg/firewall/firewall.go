package firewall

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/cache"

	coreChain "github.com/keep-network/keep-core/pkg/chain"
	coreFirewall "github.com/keep-network/keep-core/pkg/firewall"
	coreNet "github.com/keep-network/keep-core/pkg/net"

	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

var logger = log.Logger("keep-firewall")

const (
	// authorizationCachePeriod it the time the cache maintains information
	// about operators authorized and not authorized for the keep factory.
	// We use the cache to minimize calls to Ethereum client.
	authorizationCachePeriod = 24 * time.Hour

	// activeKeepMemberCachePeriod is the time the cache maintains information
	// about active and no active keep members. We use the cache to minimize
	// calls to Ethereum client.
	activeKeepMemberCachePeriod = 168 * time.Hour // one week
)

var errNoAuthorization = fmt.Errorf("remote peer has no authorization on the factory")

var errNoMinStakeNoActiveKeep = fmt.Errorf("remote peer has no minimum " +
	"stake and is not a member in any of active keeps")

// NewStakeOrActiveKeepPolicy is a firewall policy checking if the remote peer
// has a minimum stake and in case it has no minimum stake if it is a member of
// at least one active keep.
func NewStakeOrActiveKeepPolicy(
	chainHandle chain.Handle,
	stakeMonitor coreChain.StakeMonitor,
) coreNet.Firewall {
	return &stakeOrActiveKeepPolicy{
		chain:                       chainHandle,
		minimumStakePolicy:          coreFirewall.MinimumStakePolicy(stakeMonitor),
		authorizedOperatorsCache:    cache.NewTimeCache(authorizationCachePeriod),
		nonAuthorizedOperatorsCache: cache.NewTimeCache(authorizationCachePeriod),
		activeKeepMembersCache:      cache.NewTimeCache(activeKeepMemberCachePeriod),
		noActiveKeepMembersCache:    cache.NewTimeCache(activeKeepMemberCachePeriod),
		keepInfoCache:               newKeepInfoCache(activeKeepMemberCachePeriod),
	}
}

type stakeOrActiveKeepPolicy struct {
	chain                       chain.Handle
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
	remotePeerOperatorID := soakp.chain.PublicKeyToOperatorID(remotePeerPublicKey)

	logger.Debugf(
		"validating firewall rules for operator ID [%v]",
		remotePeerOperatorID,
	)

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
	err := soakp.validateAuthorization(remotePeerOperatorID)
	if err != nil {
		return err
	}

	// In case the remote peer has no minimum stake, we need to check if it is
	// a member in at least one active keep. If so, we let to connect.
	// Otherwise, we do not let to connect.
	return soakp.validateActiveKeepMembership(remotePeerOperatorID)
}

func (soakp *stakeOrActiveKeepPolicy) validateAuthorization(
	remotePeerOperatorID chain.ID,
) error {
	logger.Debugf(
		"validating authorization for [%v]",
		remotePeerOperatorID,
	)

	// Before hitting ETH client, consult the in-memory time cache.
	// If the caching time for the given entry elapsed or if that entry is
	// not in the cache, we'll have to consult the chain and execute a call
	// to ETH client.
	soakp.authorizedOperatorsCache.Sweep()
	soakp.nonAuthorizedOperatorsCache.Sweep()

	if soakp.authorizedOperatorsCache.Has(remotePeerOperatorID.String()) {
		return nil
	}

	if soakp.nonAuthorizedOperatorsCache.Has(remotePeerOperatorID.String()) {
		return errNoAuthorization
	}

	// We do not know if the remote peer has or has not the authorization so
	// we need to ask ETH client about it.
	isAuthorized, err := soakp.chain.IsOperatorAuthorized(remotePeerOperatorID)
	if err != nil {
		return fmt.Errorf(
			"could not validate authorization for operator ID [%v]: [%v]",
			remotePeerOperatorID,
			err,
		)
	}

	if !isAuthorized {
		soakp.nonAuthorizedOperatorsCache.Add(remotePeerOperatorID.String())
		return errNoAuthorization
	}

	soakp.authorizedOperatorsCache.Add(remotePeerOperatorID.String())
	return nil
}

func (soakp *stakeOrActiveKeepPolicy) validateActiveKeepMembership(
	remotePeerOperatorID chain.ID,
) error {
	logger.Debugf(
		"validating active keep membership for [%v]",
		remotePeerOperatorID,
	)

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

	if soakp.activeKeepMembersCache.Has(remotePeerOperatorID.String()) {
		return nil
	}

	if soakp.noActiveKeepMembersCache.Has(remotePeerOperatorID.String()) {
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
		keep, err := soakp.getKeepAtIndex(keepIndex)
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
		isActive, err := soakp.isKeepActive(keep)
		if err != nil {
			logger.Errorf(
				"could not check if keep [%s] is active: [%v]",
				keep.ID(),
				err,
			)
			continue
		}
		if !isActive {
			continue
		}

		// Get all the members of the active keep and store them in the active
		// keep members cache.
		members, err := soakp.getKeepMembers(keep)
		if err != nil {
			logger.Errorf(
				"could not get members of keep [%s]: [%v]",
				keep.ID(),
				err,
			)
			continue
		}
		for _, member := range members {
			soakp.activeKeepMembersCache.Add(member)
		}

		// If the remote peer address has been added to the cache we can
		// connect with this client as it's a member of an active keep.
		if soakp.activeKeepMembersCache.Has(remotePeerOperatorID.String()) {
			return nil
		}
	}

	soakp.noActiveKeepMembersCache.Add(remotePeerOperatorID.String())

	// If we are here, it means that the client is not a member in any of
	// active keeps and it's minimum stake check failed as well. We are not
	// allowing to connect with that peer.
	return errNoMinStakeNoActiveKeep
}

// getKeepAtIndex returns keep handle for a keep registered in the factory
// under the given index. Keep index to keep ID mapping is cached so this
// function allows to limit the number of on-chain calls.
func (soakp *stakeOrActiveKeepPolicy) getKeepAtIndex(
	index *big.Int,
) (chain.BondedECDSAKeepHandle, error) {
	cache := soakp.keepInfoCache

	cache.mutex.RLock()
	cachedID, isCached := cache.indexToID[index.String()]
	cache.mutex.RUnlock()

	if isCached {
		return soakp.chain.GetKeepWithID(cachedID)
	}

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cachedID, isCached = cache.indexToID[index.String()]
	if isCached {
		return soakp.chain.GetKeepWithID(cachedID)
	}

	logger.Debugf("fetching keep at index [%v] from the chain", index)
	keep, err := soakp.chain.GetKeepAtIndex(index)
	if err != nil {
		return nil, err
	}

	cache.indexToID[index.String()] = keep.ID()

	return keep, nil
}

// isKeepActive checks whether the keep with the given address is active. If the
// keep has been previously identified as inactive, function returns false
// without querying the chain. If the keep has been identified as active some
// time ago, depending on the time-caching policy, this function may return
// true without querying the chain.
func (soakp *stakeOrActiveKeepPolicy) isKeepActive(
	keep chain.BondedECDSAKeepHandle,
) (bool, error) {
	cache := soakp.keepInfoCache

	cache.mutex.RLock()
	isInactive, isCached := cache.isInactive[keep.ID().String()]
	cache.mutex.RUnlock()

	if isCached && isInactive {
		return false, nil
	}

	cache.isActive.Sweep()
	if cache.isActive.Has(keep.ID().String()) {
		return true, nil
	}

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	isInactive, isCached = cache.isInactive[keep.ID().String()]
	if isCached && isInactive {
		return false, nil
	}
	if cache.isActive.Has(keep.ID().String()) {
		return true, nil
	}

	logger.Debugf(
		"checking if keep with ID [%v] is active on the chain",
		keep.ID(),
	)
	isActive, err := keep.IsActive()
	if err != nil {
		return false, err
	}

	if isActive {
		cache.isActive.Add(keep.ID().String())
	} else {
		cache.isInactive[keep.ID().String()] = true
	}

	return isActive, nil
}

// getKeepMembers fetches members of the keep with the given address from the
// chain or reads them from a cache if this information is available there.
func (soakp *stakeOrActiveKeepPolicy) getKeepMembers(
	keep chain.BondedECDSAKeepHandle,
) ([]string, error) {
	cache := soakp.keepInfoCache

	cache.mutex.RLock()
	members, areCached := cache.members[keep.ID().String()]
	cache.mutex.RUnlock()

	if areCached {
		return members, nil
	}

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	members, areCached = cache.members[keep.ID().String()]
	if areCached {
		return members, nil
	}

	logger.Debugf(
		"getting members of the keep with ID [%v] from the chain",
		keep.ID(),
	)
	memberIDs, err := keep.GetMembers()
	if err != nil {
		return nil, nil
	}

	members = make([]string, len(memberIDs))
	for i, member := range memberIDs {
		members[i] = member.String()
	}

	cache.members[keep.ID().String()] = members

	return members, nil
}

// keepInfoCache caches information about keeps obtained from the chain.
//
// There are four pieces of information cached:
// 1. Keep index - keep on-chain ID mapping. Keep never changes its ID and keep
//    factory never changes index of a keep so this information, once cached,
//    never expires.
// 2. Information whether the keep is active. Keep may become inactive at any
//    moment so this information, if cached, does expire after some time, and
//    there is a risk information in the cache is out of date.
// 3. Information whether the keep is inactive. Inactive keep can never become
//    active again. This information, once cached, never expires.
// 4. Information about keep members. Keep members never change so this
//    information, once cached, never expires.
type keepInfoCache struct {
	indexToID  map[string]chain.ID // keep index -> keep on-chain ID
	isActive   *cache.TimeCache    // keep on-chain ID -> true (if active)
	isInactive map[string]bool     // keep on-chain ID -> true (if inactive)
	members    map[string][]string // keep on-chain ID -> member on-chain IDs
	mutex      sync.RWMutex
}

func newKeepInfoCache(isActiveKeepCachePeriod time.Duration) *keepInfoCache {
	return &keepInfoCache{
		indexToID:  make(map[string]chain.ID),
		isActive:   cache.NewTimeCache(isActiveKeepCachePeriod),
		isInactive: make(map[string]bool),
		members:    make(map[string][]string),
	}
}
