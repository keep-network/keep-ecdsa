// Package ethereum contains implementation of ethereum chain interface.
package ethereum

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-core/pkg/chain/ethereum"
)

var logger = log.Logger("keep-chain-eth-ethereum")

// Signing returns signing interface for creating and verifying signatures.
func (ec *ethereumChain) Signing() chain.Signing {
	return ethutil.NewSigner(ec.accountKey.PrivateKey)
}

// BlockCounter returns a block counter.
func (ec *ethereumChain) BlockCounter() (chain.BlockCounter, error) {
	return ec.blockCounter, nil
}

func (ec *ethereumChain) withRetry(fn func() error) error {
	const numberOfRetries = 10
	const delay = 12 * time.Second

	for i := 1; ; i++ {
		err := fn()
		if err != nil {
			logger.Errorf("Error occurred [%v]; on [%v] retry", err, i)
			if i == numberOfRetries {
				return err
			}
			time.Sleep(delay)
		} else {
			return nil
		}
	}
}

// HasMinimumStake returns true if the specified address is staked.  False will
// be returned if not staked.  If err != nil then it was not possible to determine
// if the address is staked or not.
func (ec *ethereumChain) HasMinimumStake(address common.Address) (bool, error) {
	return ec.bondedECDSAKeepFactoryContract.HasMinimumStake(address)
}

// BalanceOf returns the stake balance of the specified address.
// FIXME make private
func (ec *ethereumChain) BalanceOf(address common.Address) (*big.Int, error) {
	return ec.bondedECDSAKeepFactoryContract.BalanceOf(address)
}

// BlockTimestamp returns given block's timestamp.
func (ec *ethereumChain) BlockTimestamp(blockNumber *big.Int) (uint64, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelCtx()

	header, err := ec.client.HeaderByNumber(ctx, blockNumber)
	if err != nil {
		return 0, err
	}

	return header.Time, nil
}

// weiBalanceOf returns the wei balance of the given address from the latest known block.
func (ec *ethereumChain) weiBalanceOf(address common.Address) (*big.Int, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelCtx()

	return ec.client.BalanceAt(ctx, address, nil)
}

// BalanceMonitor returns a balance monitor.
func (ec *ethereumChain) BalanceMonitor() (chain.BalanceMonitor, error) {
	return ethereum.NewBalanceMonitor(ec.weiBalanceOf), nil
}
