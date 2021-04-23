//+build !celo

// Package ethereum contains implementation of ethereum chain interface.
package ethereum

import (
	"context"
	cecdsa "crypto/ecdsa"
	"math/big"
	"sync"
	"time"

	"github.com/keep-network/keep-common/pkg/chain/ethereum"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/subscription"
	corechain "github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

var logger = log.Logger("keep-chain-eth-ethereum")

// Offline returns a chain.Handle for an offline Ethereum client. Use Connect to
// get a chain handle that can perform online actions.
func Offline(
	accountKey *keystore.Key,
	config *ethereum.Config,
) chain.OfflineHandle {
	ethereum := &ethereumChain{
		config:     config,
		accountKey: accountKey,

		transactionMutex: &sync.Mutex{},
	}

	return ethereum
}

func (ec *ethereumChain) Name() string {
	return "ethereum"
}

// operatorAddress returns client operator's Ethereum address.
func (ec *ethereumChain) operatorAddress() common.Address {
	return ec.accountKey.Address
}

func (ec *ethereumChain) OperatorID() chain.ID {
	return ethereumChainID(ec.accountKey.Address)
}

func (ec *ethereumChain) PublicKeyToOperatorID(publicKey *cecdsa.PublicKey) chain.ID {
	return ethereumChainID(crypto.PubkeyToAddress(*publicKey))
}

// Signing returns signing interface for creating and verifying signatures.
func (ec *ethereumChain) Signing() corechain.Signing {
	return ethutil.NewSigner(ec.accountKey.PrivateKey)
}

// BlockCounter returns a block counter.
func (ec *ethereumChain) BlockCounter() corechain.BlockCounter {
	return ec.blockCounter
}

// OnBondedECDSAKeepCreated installs a callback that is invoked when an on-chain
// notification of a new ECDSA keep creation is seen.
func (ec *ethereumChain) OnBondedECDSAKeepCreated(
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
		keep, err := ec.GetKeepWithID(ethereumChainID(KeepAddress))
		if err != nil {
			logger.Errorf(
				"Failed to look up keep with address [%v] for "+
					"BondedECDSAKeepCreated event at block [%v]: [%v].",
				KeepAddress,
				blockNumber,
				err,
			)
			return
		}

		thisOperatorIsMember := false
		memberIDs := []chain.ID{}
		for _, memberAddress := range Members {
			if memberAddress == ec.operatorAddress() {
				thisOperatorIsMember = true
			}

			memberIDs = append(memberIDs, ethereumChainID(memberAddress))
		}

		handler(&chain.BondedECDSAKeepCreatedEvent{
			Keep:                 keep,
			MemberIDs:            memberIDs,
			HonestThreshold:      HonestThreshold.Uint64(),
			BlockNumber:          blockNumber,
			ThisOperatorIsMember: thisOperatorIsMember,
		})
	}

	return ec.bondedECDSAKeepFactoryContract.BondedECDSAKeepCreated(
		nil,
		nil,
		nil,
		nil,
	).OnEvent(onEvent)
}

// HasMinimumStake returns true if the specified address is staked.  False will
// be returned if not staked.  If err != nil then it was not possible to determine
// if the address is staked or not.
func (ec *ethereumChain) hasMinimumStake(address common.Address) (bool, error) {
	return ec.bondedECDSAKeepFactoryContract.HasMinimumStake(address)
}

// BalanceOf returns the stake balance of the specified address.
func (ec *ethereumChain) balanceOf(address common.Address) (*big.Int, error) {
	return ec.bondedECDSAKeepFactoryContract.BalanceOf(address)
}

// IsOperatorAuthorized checks if the factory has the authorization to
// operate on stake represented by the provided operator.
func (ec *ethereumChain) IsOperatorAuthorized(operatorID chain.ID) (bool, error) {
	operatorAddress, err := fromChainID(operatorID)
	if err != nil {
		return false, err
	}

	return ec.bondedECDSAKeepFactoryContract.IsOperatorAuthorized(
		operatorAddress,
	)
}

// GetKeepCount returns number of keeps.
func (ec *ethereumChain) GetKeepCount() (*big.Int, error) {
	return ec.bondedECDSAKeepFactoryContract.GetKeepCount()
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

// WeiBalanceOf returns the wei balance of the given address from the latest
// known block.
func (ec *ethereumChain) WeiBalanceOf(
	address common.Address,
) (*ethereum.Wei, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelCtx()

	balance, err := ec.client.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, err
	}

	return ethereum.WrapWei(balance), err
}

// BalanceMonitor returns a balance monitor.
func (ec *ethereumChain) BalanceMonitor() (*ethutil.BalanceMonitor, error) {
	return ethutil.NewBalanceMonitor(ec.WeiBalanceOf), nil
}
