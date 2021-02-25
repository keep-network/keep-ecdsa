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
func (c *Chain) Address() ExternalAddress {
	return toExternalAddress(c.accountKey.Address)
}

// Signing returns signing interface for creating and verifying signatures.
func (c *Chain) Signing() corechain.Signing {
	return celoutil.NewSigner(c.accountKey.PrivateKey)
}

// BlockCounter returns a block counter.
func (c *Chain) BlockCounter() corechain.BlockCounter {
	return c.blockCounter
}

// RegisterAsMemberCandidate registers client as a candidate to be selected
// to a keep.
func (c *Chain) RegisterAsMemberCandidate(
	application ExternalAddress,
) error {
	gasEstimate, err := c.bondedECDSAKeepFactoryContract.
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
	transaction, err := c.bondedECDSAKeepFactoryContract.
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
func (c *Chain) OnBondedECDSAKeepCreated(
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

	return c.bondedECDSAKeepFactoryContract.BondedECDSAKeepCreated(
		nil,
		nil,
		nil,
		nil,
	).OnEvent(onEvent)
}

// HasMinimumStake returns true if the specified address is staked.  False will
// be returned if not staked.  If err != nil then it was not possible to determine
// if the address is staked or not.
func (c *Chain) HasMinimumStake(address ExternalAddress) (bool, error) {
	return c.bondedECDSAKeepFactoryContract.HasMinimumStake(
		fromExternalAddress(address),
	)
}

// BalanceOf returns the stake balance of the specified address.
func (c *Chain) BalanceOf(address ExternalAddress) (*big.Int, error) {
	return c.bondedECDSAKeepFactoryContract.BalanceOf(
		fromExternalAddress(address),
	)
}

// IsRegisteredForApplication checks if the operator is registered
// as a signer candidate in the factory for the given application.
func (c *Chain) IsRegisteredForApplication(
	application ExternalAddress,
) (bool, error) {
	return c.bondedECDSAKeepFactoryContract.IsOperatorRegistered(
		fromExternalAddress(c.Address()),
		fromExternalAddress(application),
	)
}

// IsEligibleForApplication checks if the operator is eligible to register
// as a signer candidate for the given application.
func (c *Chain) IsEligibleForApplication(
	application ExternalAddress,
) (bool, error) {
	return c.bondedECDSAKeepFactoryContract.IsOperatorEligible(
		fromExternalAddress(c.Address()),
		fromExternalAddress(application),
	)
}

// IsStatusUpToDateForApplication checks if the operator's status
// is up to date in the signers' pool of the given application.
func (c *Chain) IsStatusUpToDateForApplication(
	application ExternalAddress,
) (bool, error) {
	return c.bondedECDSAKeepFactoryContract.IsOperatorUpToDate(
		fromExternalAddress(c.Address()),
		fromExternalAddress(application),
	)
}

// UpdateStatusForApplication updates the operator's status in the signers'
// pool for the given application.
func (c *Chain) UpdateStatusForApplication(
	application ExternalAddress,
) error {
	transaction, err := c.bondedECDSAKeepFactoryContract.UpdateOperatorStatus(
		fromExternalAddress(c.Address()),
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
func (c *Chain) IsOperatorAuthorized(
	operator ExternalAddress,
) (bool, error) {
	return c.bondedECDSAKeepFactoryContract.IsOperatorAuthorized(
		fromExternalAddress(operator),
	)
}

// GetKeepCount returns number of keeps.
func (c *Chain) GetKeepCount() (*big.Int, error) {
	return c.bondedECDSAKeepFactoryContract.GetKeepCount()
}

// GetKeepAtIndex returns the address of the keep at the given index.
func (c *Chain) GetKeepAtIndex(
	keepIndex *big.Int,
) (ExternalAddress, error) {
	address, err := c.bondedECDSAKeepFactoryContract.GetKeepAtIndex(keepIndex)
	if err != nil {
		return ExternalAddress{}, nil
	}

	return toExternalAddress(address), err
}

// BlockTimestamp returns given block's timestamp.
func (c *Chain) BlockTimestamp(blockNumber *big.Int) (uint64, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelCtx()

	header, err := c.client.HeaderByNumber(ctx, blockNumber)
	if err != nil {
		return 0, err
	}

	return header.Time, nil
}

// WeiBalanceOf returns the wei balance of the given address from the latest
// known block.
func (c *Chain) WeiBalanceOf(address ExternalAddress) (*celo.Wei, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelCtx()

	balance, err := c.client.BalanceAt(ctx, fromExternalAddress(address), nil)
	if err != nil {
		return nil, err
	}

	return celo.WrapWei(balance), err
}

// BalanceMonitor returns a balance monitor.
func (c *Chain) BalanceMonitor() (*celoutil.BalanceMonitor, error) {
	weiBalanceOf := func(address InternalAddress) (*celo.Wei, error) {
		return c.WeiBalanceOf(toExternalAddress(address))
	}

	return celoutil.NewBalanceMonitor(weiBalanceOf), nil
}
