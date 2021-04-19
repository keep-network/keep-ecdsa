//+build celo

package celo

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/celo-org/celo-blockchain/accounts/keystore"
	"github.com/celo-org/celo-blockchain/common"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/chain/celo"
	"github.com/keep-network/keep-common/pkg/chain/celo/celoutil"
	"github.com/keep-network/keep-common/pkg/subscription"

	corechain "github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

var logger = log.Logger("keep-chain-celo")

// Offline returns a chain.Handle for an offline Celo client. Use Connect to
// get a chain handle that can perform online actions.
func Offline(
	accountKey *keystore.Key,
	config *celo.Config,
) chain.OfflineHandle {
	celo := &celoChain{
		config:     config,
		accountKey: accountKey,

		transactionMutex: &sync.Mutex{},
	}

	return celo
}

func (cc *celoChain) Name() string {
	return "celo"
}

// operatorAddress returns client operator's Celo address.
func (cc *celoChain) operatorAddress() common.Address {
	return cc.accountKey.Address
}

func (cc *celoChain) OperatorID() chain.ID {
	return celoChainID(cc.accountKey.Address)
}

// Signing returns signing interface for creating and verifying signatures.
func (cc *celoChain) Signing() corechain.Signing {
	return celoutil.NewSigner(cc.accountKey.PrivateKey)
}

// BlockCounter returns a block counter.
func (cc *celoChain) BlockCounter() corechain.BlockCounter {
	return cc.blockCounter
}

// OnBondedECDSAKeepCreated installs a callback that is invoked when an on-chain
// notification of a new ECDSA keep creation is seen.
func (cc *celoChain) OnBondedECDSAKeepCreated(
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
		keep, err := cc.GetKeepWithID(celoChainID(KeepAddress))
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
			if memberAddress == cc.operatorAddress() {
				thisOperatorIsMember = true
			}

			memberIDs = append(memberIDs, celoChainID(memberAddress))
		}

		handler(&chain.BondedECDSAKeepCreatedEvent{
			Keep:                 keep,
			MemberIDs:            memberIDs,
			HonestThreshold:      HonestThreshold.Uint64(),
			BlockNumber:          blockNumber,
			ThisOperatorIsMember: thisOperatorIsMember,
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
func (cc *celoChain) hasMinimumStake(address common.Address) (bool, error) {
	return cc.bondedECDSAKeepFactoryContract.HasMinimumStake(
		address,
	)
}

// BalanceOf returns the stake balance of the specified address.
func (cc *celoChain) balanceOf(address common.Address) (*big.Int, error) {
	return cc.bondedECDSAKeepFactoryContract.BalanceOf(
		address,
	)
}

// IsOperatorAuthorized checks if the factory has the authorization to
// operate on stake represented by the provided operator.
func (cc *celoChain) IsOperatorAuthorized(
	operatorID chain.ID,
) (bool, error) {
	operatorAddress, err := fromChainID(operatorID)
	if err != nil {
		return false, err
	}

	return cc.bondedECDSAKeepFactoryContract.IsOperatorAuthorized(
		operatorAddress,
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

// weiBalanceOf returns the wei balance of the given address from the latest
// known block.
func (cc *celoChain) weiBalanceOf(address common.Address) (*celo.Wei, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelCtx()

	balance, err := cc.client.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, err
	}

	return celo.WrapWei(balance), err
}
