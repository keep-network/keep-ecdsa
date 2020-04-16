package firewall

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ipfs/go-log"

	coreChain "github.com/keep-network/keep-core/pkg/chain"
	coreFirewall "github.com/keep-network/keep-core/pkg/firewall"
	coreNet "github.com/keep-network/keep-core/pkg/net"
	coreKey "github.com/keep-network/keep-core/pkg/net/key"

	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
)

var logger = log.Logger("keep-ecdsa-firewall")

// KeepCacheLifetime is the time the cache maintains the list of active keep
// members. We use the cache to minimize calls to Ethereum client.
const KeepCacheLifetime = 24 * time.Hour

var errNoMinStakeNoActiveKeep = fmt.Errorf("remote peer has no minimum " +
	"stake and is not a member in any of active keeps")

// NewStakeOrActiveKeepPolicy if a firewall policy checking if the remote peer
// has a minimum stake and if it has no minimum stake if it is a member in at
// least one active keep.
func NewStakeOrActiveKeepPolicy(
	chain eth.Handle,
	stakeMonitor coreChain.StakeMonitor,
) coreNet.Firewall {
	return &stakeOrActiveKeepPolicy{
		chain:                  chain,
		minimumStakePolicy:     coreFirewall.MinimumStakePolicy(stakeMonitor),
		activeKeepMembersCache: newTimeCache(KeepCacheLifetime),
	}
}

type stakeOrActiveKeepPolicy struct {
	chain                  eth.Handle
	minimumStakePolicy     coreNet.Firewall
	activeKeepMembersCache *timeCache
}

func (soakp *stakeOrActiveKeepPolicy) Validate(
	remotePeerPublicKey *ecdsa.PublicKey,
) error {
	// Validate minimum stake policy. If the remote peer has the minimum stake,
	// we are fine and we should let to connect.
	if err := soakp.minimumStakePolicy.Validate(remotePeerPublicKey); err == nil {
		return nil
	}

	// In case the remote peer has no minimum stake, we need to check if it is
	// a member in at least one active keep. If so, we let to connect.
	// Otherwise, we do not let to connect.
	return soakp.validateActiveKeepMembership(remotePeerPublicKey)
}

func (soakp *stakeOrActiveKeepPolicy) validateActiveKeepMembership(
	remotePeerPublicKey *ecdsa.PublicKey,
) error {
	remotePeerNetworkPublicKey := coreKey.NetworkPublic(*remotePeerPublicKey)
	remotePeerAddress := coreKey.NetworkPubKeyToEthAddress(&remotePeerNetworkPublicKey)

	// First, check in the in-memory time cache to minimize hits to ETH client.
	// If the Keep client with the given chain address is in the cache it means
	// it's been a member in at least one active keep the last time
	// validateActiveKeepMembership was executed and caching period has not
	// elapsed yet.
	//
	// If the caching period elapsed, this check will return false and we
	// have to ask the chain about the current status.
	//
	// Similarly, if the client was not a member of an active keep the last time
	// validateActiveKeepMembership was executed, we have to ask the chain
	// about the current status.
	if soakp.activeKeepMembersCache.has(remotePeerAddress) {
		return nil
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
		keep, err := soakp.chain.GetKeepAtIndex(keepIndex)
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
		isActive, err := soakp.chain.IsActive(keep)
		if err != nil {
			logger.Errorf(
				"could not check if keep [%x] is active: [%v]",
				keepIndex,
				err,
			)
			continue
		}
		if !isActive {
			continue
		}

		// Get all the members of the active keep and store them in the active
		// keep members cache.
		members, err := soakp.chain.GetMembers(keep)
		if err != nil {
			logger.Errorf(
				"could not get members of keep [%x]: [%v]",
				keepIndex,
				err,
			)
			continue
		}
		for _, member := range members {
			soakp.activeKeepMembersCache.add(member.String())
		}

		// If the remote peer address has been added to the cache we can
		// connect with this client as it's a member of an active keep.
		if soakp.activeKeepMembersCache.has(remotePeerAddress) {
			return nil
		}
	}

	// If we are here, it means that the client is not a member in any of
	// active keeps and it's minimum stake check failed as well. We are not
	// allowing to connect with that peer.
	return errNoMinStakeNoActiveKeep
}
