// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	hostchainabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"

	"github.com/ipfs/go-log"

	chainutil "github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/ethereum/abi"
)

// Create a package-level logger for this contract. The logger exists at
// package level so that the logger is registered at startup and can be
// included or excluded from logging at startup by name.
var becdsakvLogger = log.Logger("keep-contract-BondedECDSAKeepVendor")

type BondedECDSAKeepVendor struct {
	contract          *abi.BondedECDSAKeepVendorImplV1
	contractAddress   common.Address
	contractABI       *hostchainabi.ABI
	caller            bind.ContractCaller
	transactor        bind.ContractTransactor
	callerOptions     *bind.CallOpts
	transactorOptions *bind.TransactOpts
	errorResolver     *chainutil.ErrorResolver
	nonceManager      *ethlike.NonceManager
	miningWaiter      *ethlike.MiningWaiter
	blockCounter      *ethlike.BlockCounter

	transactionMutex *sync.Mutex
}

func NewBondedECDSAKeepVendor(
	contractAddress common.Address,
	chainId *big.Int,
	accountKey *keystore.Key,
	backend bind.ContractBackend,
	nonceManager *ethlike.NonceManager,
	miningWaiter *ethlike.MiningWaiter,
	blockCounter *ethlike.BlockCounter,
	transactionMutex *sync.Mutex,
) (*BondedECDSAKeepVendor, error) {
	callerOptions := &bind.CallOpts{
		From: accountKey.Address,
	}

	// FIXME Switch to bind.NewKeyedTransactorWithChainID when
	// FIXME celo-org/celo-blockchain merges in changes from upstream
	// FIXME ethereum/go-ethereum beyond v1.9.25.
	transactorOptions, err := chainutil.NewKeyedTransactorWithChainID(
		accountKey.PrivateKey,
		chainId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate transactor: [%v]", err)
	}

	contract, err := abi.NewBondedECDSAKeepVendorImplV1(
		contractAddress,
		backend,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to instantiate contract at address: %s [%v]",
			contractAddress.String(),
			err,
		)
	}

	contractABI, err := hostchainabi.JSON(strings.NewReader(abi.BondedECDSAKeepVendorImplV1ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ABI: [%v]", err)
	}

	return &BondedECDSAKeepVendor{
		contract:          contract,
		contractAddress:   contractAddress,
		contractABI:       &contractABI,
		caller:            backend,
		transactor:        backend,
		callerOptions:     callerOptions,
		transactorOptions: transactorOptions,
		errorResolver:     chainutil.NewErrorResolver(backend, &contractABI, &contractAddress),
		nonceManager:      nonceManager,
		miningWaiter:      miningWaiter,
		blockCounter:      blockCounter,
		transactionMutex:  transactionMutex,
	}, nil
}

// ----- Non-const Methods ------

// Transaction submission.
func (becdsakv *BondedECDSAKeepVendor) CompleteFactoryUpgrade(

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakvLogger.Debug(
		"submitting transaction completeFactoryUpgrade",
	)

	becdsakv.transactionMutex.Lock()
	defer becdsakv.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsakv.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsakv.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsakv.contract.CompleteFactoryUpgrade(
		transactorOptions,
	)
	if err != nil {
		return transaction, becdsakv.errorResolver.ResolveError(
			err,
			becdsakv.transactorOptions.From,
			nil,
			"completeFactoryUpgrade",
		)
	}

	becdsakvLogger.Infof(
		"submitted transaction completeFactoryUpgrade with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsakv.miningWaiter.ForceMining(
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsakv.contract.CompleteFactoryUpgrade(
				transactorOptions,
			)
			if err != nil {
				return nil, becdsakv.errorResolver.ResolveError(
					err,
					becdsakv.transactorOptions.From,
					nil,
					"completeFactoryUpgrade",
				)
			}

			becdsakvLogger.Infof(
				"submitted transaction completeFactoryUpgrade with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
		},
	)

	becdsakv.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsakv *BondedECDSAKeepVendor) CallCompleteFactoryUpgrade(
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		becdsakv.transactorOptions.From,
		blockNumber, nil,
		becdsakv.contractABI,
		becdsakv.caller,
		becdsakv.errorResolver,
		becdsakv.contractAddress,
		"completeFactoryUpgrade",
		&result,
	)

	return err
}

func (becdsakv *BondedECDSAKeepVendor) CompleteFactoryUpgradeGasEstimate() (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		becdsakv.callerOptions.From,
		becdsakv.contractAddress,
		"completeFactoryUpgrade",
		becdsakv.contractABI,
		becdsakv.transactor,
	)

	return result, err
}

// Transaction submission.
func (becdsakv *BondedECDSAKeepVendor) Initialize(
	registryAddress common.Address,
	factory common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakvLogger.Debug(
		"submitting transaction initialize",
		"params: ",
		fmt.Sprint(
			registryAddress,
			factory,
		),
	)

	becdsakv.transactionMutex.Lock()
	defer becdsakv.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsakv.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsakv.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsakv.contract.Initialize(
		transactorOptions,
		registryAddress,
		factory,
	)
	if err != nil {
		return transaction, becdsakv.errorResolver.ResolveError(
			err,
			becdsakv.transactorOptions.From,
			nil,
			"initialize",
			registryAddress,
			factory,
		)
	}

	becdsakvLogger.Infof(
		"submitted transaction initialize with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsakv.miningWaiter.ForceMining(
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsakv.contract.Initialize(
				transactorOptions,
				registryAddress,
				factory,
			)
			if err != nil {
				return nil, becdsakv.errorResolver.ResolveError(
					err,
					becdsakv.transactorOptions.From,
					nil,
					"initialize",
					registryAddress,
					factory,
				)
			}

			becdsakvLogger.Infof(
				"submitted transaction initialize with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
		},
	)

	becdsakv.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsakv *BondedECDSAKeepVendor) CallInitialize(
	registryAddress common.Address,
	factory common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		becdsakv.transactorOptions.From,
		blockNumber, nil,
		becdsakv.contractABI,
		becdsakv.caller,
		becdsakv.errorResolver,
		becdsakv.contractAddress,
		"initialize",
		&result,
		registryAddress,
		factory,
	)

	return err
}

func (becdsakv *BondedECDSAKeepVendor) InitializeGasEstimate(
	registryAddress common.Address,
	factory common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		becdsakv.callerOptions.From,
		becdsakv.contractAddress,
		"initialize",
		becdsakv.contractABI,
		becdsakv.transactor,
		registryAddress,
		factory,
	)

	return result, err
}

// Transaction submission.
func (becdsakv *BondedECDSAKeepVendor) UpgradeFactory(
	_factory common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakvLogger.Debug(
		"submitting transaction upgradeFactory",
		"params: ",
		fmt.Sprint(
			_factory,
		),
	)

	becdsakv.transactionMutex.Lock()
	defer becdsakv.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsakv.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsakv.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsakv.contract.UpgradeFactory(
		transactorOptions,
		_factory,
	)
	if err != nil {
		return transaction, becdsakv.errorResolver.ResolveError(
			err,
			becdsakv.transactorOptions.From,
			nil,
			"upgradeFactory",
			_factory,
		)
	}

	becdsakvLogger.Infof(
		"submitted transaction upgradeFactory with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsakv.miningWaiter.ForceMining(
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsakv.contract.UpgradeFactory(
				transactorOptions,
				_factory,
			)
			if err != nil {
				return nil, becdsakv.errorResolver.ResolveError(
					err,
					becdsakv.transactorOptions.From,
					nil,
					"upgradeFactory",
					_factory,
				)
			}

			becdsakvLogger.Infof(
				"submitted transaction upgradeFactory with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
		},
	)

	becdsakv.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsakv *BondedECDSAKeepVendor) CallUpgradeFactory(
	_factory common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		becdsakv.transactorOptions.From,
		blockNumber, nil,
		becdsakv.contractABI,
		becdsakv.caller,
		becdsakv.errorResolver,
		becdsakv.contractAddress,
		"upgradeFactory",
		&result,
		_factory,
	)

	return err
}

func (becdsakv *BondedECDSAKeepVendor) UpgradeFactoryGasEstimate(
	_factory common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		becdsakv.callerOptions.From,
		becdsakv.contractAddress,
		"upgradeFactory",
		becdsakv.contractABI,
		becdsakv.transactor,
		_factory,
	)

	return result, err
}

// ----- Const Methods ------

func (becdsakv *BondedECDSAKeepVendor) FactoryUpgradeTimeDelay() (*big.Int, error) {
	var result *big.Int
	result, err := becdsakv.contract.FactoryUpgradeTimeDelay(
		becdsakv.callerOptions,
	)

	if err != nil {
		return result, becdsakv.errorResolver.ResolveError(
			err,
			becdsakv.callerOptions.From,
			nil,
			"factoryUpgradeTimeDelay",
		)
	}

	return result, err
}

func (becdsakv *BondedECDSAKeepVendor) FactoryUpgradeTimeDelayAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		becdsakv.callerOptions.From,
		blockNumber,
		nil,
		becdsakv.contractABI,
		becdsakv.caller,
		becdsakv.errorResolver,
		becdsakv.contractAddress,
		"factoryUpgradeTimeDelay",
		&result,
	)

	return result, err
}

func (becdsakv *BondedECDSAKeepVendor) Initialized() (bool, error) {
	var result bool
	result, err := becdsakv.contract.Initialized(
		becdsakv.callerOptions,
	)

	if err != nil {
		return result, becdsakv.errorResolver.ResolveError(
			err,
			becdsakv.callerOptions.From,
			nil,
			"initialized",
		)
	}

	return result, err
}

func (becdsakv *BondedECDSAKeepVendor) InitializedAtBlock(
	blockNumber *big.Int,
) (bool, error) {
	var result bool

	err := chainutil.CallAtBlock(
		becdsakv.callerOptions.From,
		blockNumber,
		nil,
		becdsakv.contractABI,
		becdsakv.caller,
		becdsakv.errorResolver,
		becdsakv.contractAddress,
		"initialized",
		&result,
	)

	return result, err
}

func (becdsakv *BondedECDSAKeepVendor) SelectFactory() (common.Address, error) {
	var result common.Address
	result, err := becdsakv.contract.SelectFactory(
		becdsakv.callerOptions,
	)

	if err != nil {
		return result, becdsakv.errorResolver.ResolveError(
			err,
			becdsakv.callerOptions.From,
			nil,
			"selectFactory",
		)
	}

	return result, err
}

func (becdsakv *BondedECDSAKeepVendor) SelectFactoryAtBlock(
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		becdsakv.callerOptions.From,
		blockNumber,
		nil,
		becdsakv.contractABI,
		becdsakv.caller,
		becdsakv.errorResolver,
		becdsakv.contractAddress,
		"selectFactory",
		&result,
	)

	return result, err
}

// ------ Events -------

func (becdsakv *BondedECDSAKeepVendor) FactoryUpgradeCompleted(
	opts *ethlike.SubscribeOpts,
) *BecdsakvFactoryUpgradeCompletedSubscription {
	if opts == nil {
		opts = new(ethlike.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BecdsakvFactoryUpgradeCompletedSubscription{
		becdsakv,
		opts,
	}
}

type BecdsakvFactoryUpgradeCompletedSubscription struct {
	contract *BondedECDSAKeepVendor
	opts     *ethlike.SubscribeOpts
}

type bondedECDSAKeepVendorFactoryUpgradeCompletedFunc func(
	Factory common.Address,
	blockNumber uint64,
)

func (fucs *BecdsakvFactoryUpgradeCompletedSubscription) OnEvent(
	handler bondedECDSAKeepVendorFactoryUpgradeCompletedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BondedECDSAKeepVendorImplV1FactoryUpgradeCompleted)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Factory,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := fucs.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (fucs *BecdsakvFactoryUpgradeCompletedSubscription) Pipe(
	sink chan *abi.BondedECDSAKeepVendorImplV1FactoryUpgradeCompleted,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(fucs.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := fucs.contract.blockCounter.CurrentBlock()
				if err != nil {
					becdsakvLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - fucs.opts.PastBlocks

				becdsakvLogger.Infof(
					"subscription monitoring fetching past FactoryUpgradeCompleted events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := fucs.contract.PastFactoryUpgradeCompletedEvents(
					fromBlock,
					nil,
				)
				if err != nil {
					becdsakvLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				becdsakvLogger.Infof(
					"subscription monitoring fetched [%v] past FactoryUpgradeCompleted events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := fucs.contract.watchFactoryUpgradeCompleted(
		sink,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (becdsakv *BondedECDSAKeepVendor) watchFactoryUpgradeCompleted(
	sink chan *abi.BondedECDSAKeepVendorImplV1FactoryUpgradeCompleted,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return becdsakv.contract.WatchFactoryUpgradeCompleted(
			&bind.WatchOpts{Context: ctx},
			sink,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		becdsakvLogger.Errorf(
			"subscription to event FactoryUpgradeCompleted had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		becdsakvLogger.Errorf(
			"subscription to event FactoryUpgradeCompleted failed "+
				"with error: [%v]; resubscription attempt will be "+
				"performed",
			err,
		)
	}

	return chainutil.WithResubscription(
		chainutil.SubscriptionBackoffMax,
		subscribeFn,
		chainutil.SubscriptionAlertThreshold,
		thresholdViolatedFn,
		subscriptionFailedFn,
	)
}

func (becdsakv *BondedECDSAKeepVendor) PastFactoryUpgradeCompletedEvents(
	startBlock uint64,
	endBlock *uint64,
) ([]*abi.BondedECDSAKeepVendorImplV1FactoryUpgradeCompleted, error) {
	iterator, err := becdsakv.contract.FilterFactoryUpgradeCompleted(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past FactoryUpgradeCompleted events: [%v]",
			err,
		)
	}

	events := make([]*abi.BondedECDSAKeepVendorImplV1FactoryUpgradeCompleted, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (becdsakv *BondedECDSAKeepVendor) FactoryUpgradeStarted(
	opts *ethlike.SubscribeOpts,
) *BecdsakvFactoryUpgradeStartedSubscription {
	if opts == nil {
		opts = new(ethlike.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BecdsakvFactoryUpgradeStartedSubscription{
		becdsakv,
		opts,
	}
}

type BecdsakvFactoryUpgradeStartedSubscription struct {
	contract *BondedECDSAKeepVendor
	opts     *ethlike.SubscribeOpts
}

type bondedECDSAKeepVendorFactoryUpgradeStartedFunc func(
	Factory common.Address,
	Timestamp *big.Int,
	blockNumber uint64,
)

func (fuss *BecdsakvFactoryUpgradeStartedSubscription) OnEvent(
	handler bondedECDSAKeepVendorFactoryUpgradeStartedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BondedECDSAKeepVendorImplV1FactoryUpgradeStarted)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Factory,
					event.Timestamp,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := fuss.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (fuss *BecdsakvFactoryUpgradeStartedSubscription) Pipe(
	sink chan *abi.BondedECDSAKeepVendorImplV1FactoryUpgradeStarted,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(fuss.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := fuss.contract.blockCounter.CurrentBlock()
				if err != nil {
					becdsakvLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - fuss.opts.PastBlocks

				becdsakvLogger.Infof(
					"subscription monitoring fetching past FactoryUpgradeStarted events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := fuss.contract.PastFactoryUpgradeStartedEvents(
					fromBlock,
					nil,
				)
				if err != nil {
					becdsakvLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				becdsakvLogger.Infof(
					"subscription monitoring fetched [%v] past FactoryUpgradeStarted events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := fuss.contract.watchFactoryUpgradeStarted(
		sink,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (becdsakv *BondedECDSAKeepVendor) watchFactoryUpgradeStarted(
	sink chan *abi.BondedECDSAKeepVendorImplV1FactoryUpgradeStarted,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return becdsakv.contract.WatchFactoryUpgradeStarted(
			&bind.WatchOpts{Context: ctx},
			sink,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		becdsakvLogger.Errorf(
			"subscription to event FactoryUpgradeStarted had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		becdsakvLogger.Errorf(
			"subscription to event FactoryUpgradeStarted failed "+
				"with error: [%v]; resubscription attempt will be "+
				"performed",
			err,
		)
	}

	return chainutil.WithResubscription(
		chainutil.SubscriptionBackoffMax,
		subscribeFn,
		chainutil.SubscriptionAlertThreshold,
		thresholdViolatedFn,
		subscriptionFailedFn,
	)
}

func (becdsakv *BondedECDSAKeepVendor) PastFactoryUpgradeStartedEvents(
	startBlock uint64,
	endBlock *uint64,
) ([]*abi.BondedECDSAKeepVendorImplV1FactoryUpgradeStarted, error) {
	iterator, err := becdsakv.contract.FilterFactoryUpgradeStarted(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past FactoryUpgradeStarted events: [%v]",
			err,
		)
	}

	events := make([]*abi.BondedECDSAKeepVendorImplV1FactoryUpgradeStarted, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}
