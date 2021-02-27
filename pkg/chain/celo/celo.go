//+build celo

package celo

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/keep-network/keep-common/pkg/chain/celo"

	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-common/pkg/chain/celo/celoutil"
	"github.com/keep-network/keep-common/pkg/subscription"
	corechain "github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

var logger = log.Logger("keep-chain-celo")

// Address returns client's Celo address.
func (cc *celoChain) Address() ExternalAddress {
	return toExternalAddress(cc.accountKey.Address)
}

// Signing returns signing interface for creating and verifying signatures.
func (cc *celoChain) Signing() corechain.Signing {
	return celoutil.NewSigner(cc.accountKey.PrivateKey)
}

// BlockCounter returns a block counter.
func (cc *celoChain) BlockCounter() corechain.BlockCounter {
	return cc.blockCounter
}

// RegisterAsMemberCandidate registers client as a candidate to be selected
// to a keep.
func (cc *celoChain) RegisterAsMemberCandidate(
	application ExternalAddress,
) error {
	gasEstimate, err := cc.bondedECDSAKeepFactoryContract.
		RegisterMemberCandidateGasEstimate(fromExternalAddress(application))
	if err != nil {
		return fmt.Errorf("failed to estimate gas [%v]", err)
	}

	// If we have multiple sortition pool join transactions queued - and that
	// happens when multiple operators become eligible to join at the same time,
	// e.g. after lowering the minimum bond requirement, transactions mined at
	// the end may no longer have valid gas limits as they were estimated based
	// on a different state of the pool. We add 20% safety margin to the
	// original gas estimation to account for that.
	gasEstimateWithMargin := float64(gasEstimate) * float64(1.2)
	transaction, err := cc.bondedECDSAKeepFactoryContract.
		RegisterMemberCandidate(
			fromExternalAddress(application),
			celoutil.TransactionOptions{
				GasLimit: uint64(gasEstimateWithMargin),
			},
		)
	if err != nil {
		return err
	}

	logger.Debugf(
		"submitted RegisterMemberCandidate transaction with hash: [%x]",
		transaction.Hash(),
	)

	return nil
}

// OnBondedECDSAKeepCreated installs a callback that is invoked when an on-chain
// notification of a new ECDSA keep creation is seen.
func (cc *celoChain) OnBondedECDSAKeepCreated(
	handler func(event *chain.BondedECDSAKeepCreatedEvent),
) subscription.EventSubscription {
	onEvent := func(
		KeepAddress InternalAddress,
		Members []InternalAddress,
		Owner InternalAddress,
		Application InternalAddress,
		HonestThreshold *big.Int,
		blockNumber uint64,
	) {
		handler(&chain.BondedECDSAKeepCreatedEvent{
			KeepAddress:     toExternalAddress(KeepAddress),
			Members:         toExternalAddresses(Members),
			HonestThreshold: HonestThreshold.Uint64(),
			BlockNumber:     blockNumber,
		})
	}

	return cc.bondedECDSAKeepFactoryContract.BondedECDSAKeepCreated(
		nil,
		nil,
		nil,
		nil,
	).OnEvent(onEvent)
}

// HasMinimumStake returns true if the specified address is staked.  False will
// be returned if not staked.  If err != nil then it was not possible to determine
// if the address is staked or not.
func (cc *celoChain) HasMinimumStake(address ExternalAddress) (bool, error) {
	return cc.bondedECDSAKeepFactoryContract.HasMinimumStake(
		fromExternalAddress(address),
	)
}

// BalanceOf returns the stake balance of the specified address.
func (cc *celoChain) BalanceOf(address ExternalAddress) (*big.Int, error) {
	return cc.bondedECDSAKeepFactoryContract.BalanceOf(
		fromExternalAddress(address),
	)
}

// IsRegisteredForApplication checks if the operator is registered
// as a signer candidate in the factory for the given application.
func (cc *celoChain) IsRegisteredForApplication(
	application ExternalAddress,
) (bool, error) {
	return cc.bondedECDSAKeepFactoryContract.IsOperatorRegistered(
		fromExternalAddress(cc.Address()),
		fromExternalAddress(application),
	)
}

// IsEligibleForApplication checks if the operator is eligible to register
// as a signer candidate for the given application.
func (cc *celoChain) IsEligibleForApplication(
	application ExternalAddress,
) (bool, error) {
	return cc.bondedECDSAKeepFactoryContract.IsOperatorEligible(
		fromExternalAddress(cc.Address()),
		fromExternalAddress(application),
	)
}

// IsStatusUpToDateForApplication checks if the operator's status
// is up to date in the signers' pool of the given application.
func (cc *celoChain) IsStatusUpToDateForApplication(
	application ExternalAddress,
) (bool, error) {
	return cc.bondedECDSAKeepFactoryContract.IsOperatorUpToDate(
		fromExternalAddress(cc.Address()),
		fromExternalAddress(application),
	)
}

// UpdateStatusForApplication updates the operator's status in the signers'
// pool for the given application.
func (cc *celoChain) UpdateStatusForApplication(
	application ExternalAddress,
) error {
	transaction, err := cc.bondedECDSAKeepFactoryContract.UpdateOperatorStatus(
		fromExternalAddress(cc.Address()),
		fromExternalAddress(application),
	)
	if err != nil {
		return err
	}

	logger.Debugf(
		"submitted UpdateOperatorStatus transaction with hash: [%x]",
		transaction.Hash(),
	)

	return nil
}

// IsOperatorAuthorized checks if the factory has the authorization to
// operate on stake represented by the provided operator.
func (cc *celoChain) IsOperatorAuthorized(
	operator ExternalAddress,
) (bool, error) {
	return cc.bondedECDSAKeepFactoryContract.IsOperatorAuthorized(
		fromExternalAddress(operator),
	)
}

// GetKeepCount returns number of keeps.
func (cc *celoChain) GetKeepCount() (*big.Int, error) {
	return cc.bondedECDSAKeepFactoryContract.GetKeepCount()
}

// BlockTimestamp returns given block's timestamp.
func (cc *celoChain) BlockTimestamp(blockNumber *big.Int) (uint64, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelCtx()

	header, err := cc.client.HeaderByNumber(ctx, blockNumber)
	if err != nil {
		return 0, err
	}

	return header.Time, nil
}

// WeiBalanceOf returns the wei balance of the given address from the latest
// known block.
func (cc *celoChain) WeiBalanceOf(address ExternalAddress) (*celo.Wei, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelCtx()

	balance, err := cc.client.BalanceAt(ctx, fromExternalAddress(address), nil)
	if err != nil {
		return nil, err
	}

	return celo.WrapWei(balance), err
}
