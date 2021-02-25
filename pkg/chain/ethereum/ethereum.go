//+build !celo

// Package ethereum contains implementation of ethereum chain interface.
package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/keep-network/keep-common/pkg/chain/ethereum"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/subscription"
	corechain "github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

var logger = log.Logger("keep-chain-eth-ethereum")

// Address returns client's ethereum address.
func (c *Chain) Address() common.Address {
	return c.accountKey.Address
}

// Signing returns signing interface for creating and verifying signatures.
func (c *Chain) Signing() corechain.Signing {
	return ethutil.NewSigner(c.accountKey.PrivateKey)
}

// BlockCounter returns a block counter.
func (c *Chain) BlockCounter() corechain.BlockCounter {
	return c.blockCounter
}

// RegisterAsMemberCandidate registers client as a candidate to be selected
// to a keep.
func (c *Chain) RegisterAsMemberCandidate(application common.Address) error {
	gasEstimate, err := c.bondedECDSAKeepFactoryContract.RegisterMemberCandidateGasEstimate(application)
	if err != nil {
		return fmt.Errorf("failed to estimate gas [%v]", err)
	}

	// If we have multiple sortition pool join transactions queued - and that
	// happens when multiple operators become eligible to join at the same time,
	// e.g. after lowering the minimum bond requirement, transactions mined at
	// the end may no longer have valid gas limits as they were estimated based
	// on a different state of the pool. We add 20% safety margin to the original
	// gas estimation to account for that.
	gasEstimateWithMargin := float64(gasEstimate) * float64(1.2)
	transaction, err := c.bondedECDSAKeepFactoryContract.RegisterMemberCandidate(
		application,
		ethutil.TransactionOptions{
			GasLimit: uint64(gasEstimateWithMargin),
		},
	)
	if err != nil {
		return err
	}

	logger.Debugf("submitted RegisterMemberCandidate transaction with hash: [%x]", transaction.Hash())

	return nil
}

// OnBondedECDSAKeepCreated installs a callback that is invoked when an on-chain
// notification of a new ECDSA keep creation is seen.
func (c *Chain) OnBondedECDSAKeepCreated(
	handler func(event *chain.BondedECDSAKeepCreatedEvent),
) subscription.EventSubscription {
	onEvent := func(
		KeepAddress common.Address,
		Members []common.Address,
		Owner common.Address,
		Application common.Address,
		HonestThreshold *big.Int,
		blockNumber uint64,
	) {
		handler(&chain.BondedECDSAKeepCreatedEvent{
			KeepAddress:     KeepAddress,
			Members:         Members,
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
func (c *Chain) HasMinimumStake(address common.Address) (bool, error) {
	return c.bondedECDSAKeepFactoryContract.HasMinimumStake(address)
}

// BalanceOf returns the stake balance of the specified address.
func (c *Chain) BalanceOf(address common.Address) (*big.Int, error) {
	return c.bondedECDSAKeepFactoryContract.BalanceOf(address)
}

// IsRegisteredForApplication checks if the operator is registered
// as a signer candidate in the factory for the given application.
func (c *Chain) IsRegisteredForApplication(application common.Address) (bool, error) {
	return c.bondedECDSAKeepFactoryContract.IsOperatorRegistered(
		c.Address(),
		application,
	)
}

// IsEligibleForApplication checks if the operator is eligible to register
// as a signer candidate for the given application.
func (c *Chain) IsEligibleForApplication(application common.Address) (bool, error) {
	return c.bondedECDSAKeepFactoryContract.IsOperatorEligible(
		c.Address(),
		application,
	)
}

// IsStatusUpToDateForApplication checks if the operator's status
// is up to date in the signers' pool of the given application.
func (c *Chain) IsStatusUpToDateForApplication(application common.Address) (bool, error) {
	return c.bondedECDSAKeepFactoryContract.IsOperatorUpToDate(
		c.Address(),
		application,
	)
}

// UpdateStatusForApplication updates the operator's status in the signers'
// pool for the given application.
func (c *Chain) UpdateStatusForApplication(application common.Address) error {
	transaction, err := c.bondedECDSAKeepFactoryContract.UpdateOperatorStatus(
		c.Address(),
		application,
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
func (c *Chain) IsOperatorAuthorized(operator common.Address) (bool, error) {
	return c.bondedECDSAKeepFactoryContract.IsOperatorAuthorized(operator)
}

// GetKeepCount returns number of keeps.
func (c *Chain) GetKeepCount() (*big.Int, error) {
	return c.bondedECDSAKeepFactoryContract.GetKeepCount()
}

// GetKeepAtIndex returns the address of the keep at the given index.
func (c *Chain) GetKeepAtIndex(
	keepIndex *big.Int,
) (common.Address, error) {
	return c.bondedECDSAKeepFactoryContract.GetKeepAtIndex(keepIndex)
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
func (c *Chain) WeiBalanceOf(
	address common.Address,
) (*ethereum.Wei, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelCtx()

	balance, err := c.client.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, err
	}

	return ethereum.WrapWei(balance), err
}

// BalanceMonitor returns a balance monitor.
func (c *Chain) BalanceMonitor() (*ethutil.BalanceMonitor, error) {
	return ethutil.NewBalanceMonitor(c.WeiBalanceOf), nil
}
