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

	ethereumabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/abi"
)

// Create a package-level logger for this contract. The logger exists at
// package level so that the logger is registered at startup and can be
// included or excluded from logging at startup by name.
var becdsakvLogger = log.Logger("keep-contract-BondedECDSAKeepVendor")

const (
	// Maximum backoff time between event resubscription attempts.
	becdsakvSubscriptionBackoffMax = 2 * time.Minute

	// Threshold below which event resubscription emits an error to the logs.
	// WS connection can be dropped at any moment and event resubscription will
	// follow. However, if WS connection for event subscription is getting
	// dropped too often, it may indicate something is wrong with Ethereum
	// client. This constant defines the minimum lifetime of an event
	// subscription required before the subscription failure happens and
	// resubscription follows so that the resubscription does not emit an error
	// to the logs alerting about potential problems with Ethereum client.
	becdsakvSubscriptionAlertThreshold = 5 * time.Minute
)

type BondedECDSAKeepVendor struct {
	contract          *abi.BondedECDSAKeepVendorImplV1
	contractAddress   common.Address
	contractABI       *ethereumabi.ABI
	caller            bind.ContractCaller
	transactor        bind.ContractTransactor
	callerOptions     *bind.CallOpts
	transactorOptions *bind.TransactOpts
	errorResolver     *ethutil.ErrorResolver
	nonceManager      *ethutil.NonceManager
	miningWaiter      *ethutil.MiningWaiter

	transactionMutex *sync.Mutex
}

func NewBondedECDSAKeepVendor(
	contractAddress common.Address,
	accountKey *keystore.Key,
	backend bind.ContractBackend,
	nonceManager *ethutil.NonceManager,
	miningWaiter *ethutil.MiningWaiter,
	transactionMutex *sync.Mutex,
) (*BondedECDSAKeepVendor, error) {
	callerOptions := &bind.CallOpts{
		From: accountKey.Address,
	}

	transactorOptions := bind.NewKeyedTransactor(
		accountKey.PrivateKey,
	)

	randomBeaconContract, err := abi.NewBondedECDSAKeepVendorImplV1(
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

	contractABI, err := ethereumabi.JSON(strings.NewReader(abi.BondedECDSAKeepVendorImplV1ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ABI: [%v]", err)
	}

	return &BondedECDSAKeepVendor{
		contract:          randomBeaconContract,
		contractAddress:   contractAddress,
		contractABI:       &contractABI,
		caller:            backend,
		transactor:        backend,
		callerOptions:     callerOptions,
		transactorOptions: transactorOptions,
		errorResolver:     ethutil.NewErrorResolver(backend, &contractABI, &contractAddress),
		nonceManager:      nonceManager,
		miningWaiter:      miningWaiter,
		transactionMutex:  transactionMutex,
	}, nil
}

// ----- Non-const Methods ------

// Transaction submission.
func (becdsakv *BondedECDSAKeepVendor) CompleteFactoryUpgrade(

	transactionOptions ...ethutil.TransactionOptions,
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
		transaction,
		func(newGasPrice *big.Int) (*types.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

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

			return transaction, nil
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

	err := ethutil.CallAtBlock(
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

	result, err := ethutil.EstimateGas(
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

	transactionOptions ...ethutil.TransactionOptions,
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
		transaction,
		func(newGasPrice *big.Int) (*types.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

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

			return transaction, nil
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

	err := ethutil.CallAtBlock(
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

	result, err := ethutil.EstimateGas(
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

	transactionOptions ...ethutil.TransactionOptions,
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
		transaction,
		func(newGasPrice *big.Int) (*types.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

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

			return transaction, nil
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

	err := ethutil.CallAtBlock(
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

	result, err := ethutil.EstimateGas(
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

	err := ethutil.CallAtBlock(
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

	err := ethutil.CallAtBlock(
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

	err := ethutil.CallAtBlock(
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

type bondedECDSAKeepVendorFactoryUpgradeCompletedFunc func(
	Factory common.Address,
	blockNumber uint64,
)

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

func (becdsakv *BondedECDSAKeepVendor) WatchFactoryUpgradeCompleted(
	success bondedECDSAKeepVendorFactoryUpgradeCompletedFunc,
) subscription.EventSubscription {
	eventOccurred := make(chan *abi.BondedECDSAKeepVendorImplV1FactoryUpgradeCompleted)
	thresholdViolated := make(chan time.Duration)
	subscriptionFailed := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventOccurred:
				success(
					event.Factory,
					event.Raw.BlockNumber,
				)
			case violation := <-thresholdViolated:
				becdsakvLogger.Errorf(
					"subscription to event FactoryUpgradeCompleted had to be "+
						"retried [%v] since the last attempt; please inspect "+
						"Ethereum client connectivity",
					violation,
				)
			case err := <-subscriptionFailed:
				becdsakvLogger.Warningf(
					"subscription to event FactoryUpgradeCompleted failed "+
						"with error: [%v]; resubscription attempt will be "+
						"performed",
					err,
				)
			}
		}
	}()

	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return becdsakv.contract.WatchFactoryUpgradeCompleted(
			&bind.WatchOpts{Context: ctx},
			eventOccurred,
		)
	}

	sub := ethutil.WithResubscription(
		becdsakvSubscriptionBackoffMax,
		subscribeFn,
		becdsakvSubscriptionAlertThreshold,
		thresholdViolated,
		subscriptionFailed,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancel()
	})
}

type bondedECDSAKeepVendorFactoryUpgradeStartedFunc func(
	Factory common.Address,
	Timestamp *big.Int,
	blockNumber uint64,
)

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

func (becdsakv *BondedECDSAKeepVendor) WatchFactoryUpgradeStarted(
	success bondedECDSAKeepVendorFactoryUpgradeStartedFunc,
) subscription.EventSubscription {
	eventOccurred := make(chan *abi.BondedECDSAKeepVendorImplV1FactoryUpgradeStarted)
	thresholdViolated := make(chan time.Duration)
	subscriptionFailed := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventOccurred:
				success(
					event.Factory,
					event.Timestamp,
					event.Raw.BlockNumber,
				)
			case violation := <-thresholdViolated:
				becdsakvLogger.Errorf(
					"subscription to event FactoryUpgradeStarted had to be "+
						"retried [%v] since the last attempt; please inspect "+
						"Ethereum client connectivity",
					violation,
				)
			case err := <-subscriptionFailed:
				becdsakvLogger.Warningf(
					"subscription to event FactoryUpgradeStarted failed "+
						"with error: [%v]; resubscription attempt will be "+
						"performed",
					err,
				)
			}
		}
	}()

	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return becdsakv.contract.WatchFactoryUpgradeStarted(
			&bind.WatchOpts{Context: ctx},
			eventOccurred,
		)
	}

	sub := ethutil.WithResubscription(
		becdsakvSubscriptionBackoffMax,
		subscribeFn,
		becdsakvSubscriptionAlertThreshold,
		thresholdViolated,
		subscriptionFailed,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancel()
	})
}
