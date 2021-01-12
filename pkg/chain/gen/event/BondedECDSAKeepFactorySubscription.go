package event

import (
	"context"
	"math/big"
	"time"

	"github.com/ipfs/go-log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/chain/ethereum/blockcounter"
	"github.com/keep-network/keep-core/pkg/subscription"
	genContract "github.com/keep-network/keep-ecdsa/pkg/chain/gen/contract"
)

var becdsakfsLogger = log.Logger("keep-subscription-BondedECDSAKeepFactory")

type BondedECDSAKeepFactorySubscription struct {
	contract     *genContract.BondedECDSAKeepFactory
	blockCounter *blockcounter.EthereumBlockCounter
}

func NewBondedECDSAKeepFactorySubscription(
	contract *genContract.BondedECDSAKeepFactory,
	blockCounter *blockcounter.EthereumBlockCounter,
) *BondedECDSAKeepFactorySubscription {
	return &BondedECDSAKeepFactorySubscription{
		contract,
		blockCounter,
	}
}

type SubscribeOpts {
	TickDuration time.Duration
	BlocksBack uint64
}

const (
	DefaultTickDuration = 15 * time.Minute
	DefaultBlocksBack = 100
)

// TODO:
// - SubscribeOps
// - Deduplicator ? 
// - Watch* should accept a chan with abi

func (becdsakfs *BondedECDSAKeepFactorySubscription) SubscribeBondedECDSAKeepCreated(
	opts *SubscribeOpts,
	onEvent chan<- *abi.BondedECDSAKeepFactoryBondedECDSAKeepCreated,
	keepAddressFilter []common.Address,
	ownerFilter []common.Address,
	applicationFilter []common.Address,
) subscription.EventSubscription {
	if opts == nil {
		opts = new(SubscribeOpts)
	}
	if opts.TickDuration == 0 {
		opts.TickDuration = DefaultTickDuration
	}
	if opts.BlocksBack == 0 {
		opts.BlocksBack = DefaultBlocksBack
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(opts.TickDuration)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				lastBlock, err := becdsakfs.blockCounter.CurrentBlock()
				if err != nil {
					becdsakfsLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				events, err := becdsakfs.contract.PastBondedECDSAKeepCreatedEvents(
					lastBlock-opts.BlocksBack,
					nil,
					keepAddressFilter,
					ownerFilter,
					applicationFilter,
				)
				if err != nil {
					becdsakfsLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}

				for _, event := range events {
					onEvent <- &BondedECDSAKeepCreatedEvent{
						event.KeepAddress,
						event.Members,
						event.Owner,
						event.Application,
						event.HonestThreshold,
						event.Raw.BlockNumber,
					}
				}
			}
		}
	}()

	sub := becdsakfs.contract.WatchBondedECDSAKeepCreated(
		// TODO: Watch* should accept a chan instead of a handler func
		func(
			KeepAddress common.Address,
			Members []common.Address,
			Owner common.Address,
			Application common.Address,
			HonestThreshold *big.Int,
			blockNumber uint64,
		) {
			onEvent <- &BondedECDSAKeepCreatedEvent{
				KeepAddress,
				Members,
				Owner,
				Application,
				HonestThreshold,
				blockNumber,
			}
		},
		keepAddressFilter,
		ownerFilter,
		applicationFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancel()
	})
}
