// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
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

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/abi"
)

// Create a package-level logger for this contract. The logger exists at
// package level so that the logger is registered at startup and can be
// included or excluded from logging at startup by name.
var becdsakvLogger = log.Logger("keep-contract-BondedECDSAKeepVendor")

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

// ----- Const Methods ------

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
	fail func(err error) error,
) (subscription.EventSubscription, error) {
	errorChan := make(chan error)
	unsubscribeChan := make(chan struct{})

	// Delay which must be preserved before a new resubscription attempt.
	// There is no sense to resubscribe immediately after the fail of current
	// subscription because the publisher must have some time to recover.
	retryDelay := 5 * time.Second

	watch := func() {
		failCallback := func(err error) error {
			fail(err)
			errorChan <- err // trigger resubscription signal
			return err
		}

		subscription, err := becdsakv.subscribeFactoryUpgradeCompleted(
			success,
			failCallback,
		)
		if err != nil {
			errorChan <- err // trigger resubscription signal
			return
		}

		// wait for unsubscription signal
		<-unsubscribeChan
		subscription.Unsubscribe()
	}

	// trigger the resubscriber goroutine
	go func() {
		go watch() // trigger first subscription

		for {
			select {
			case <-errorChan:
				becdsakvLogger.Warning(
					"subscription to event FactoryUpgradeCompleted terminated with error; " +
						"resubscription attempt will be performed after the retry delay",
				)
				time.Sleep(retryDelay)
				go watch()
			case <-unsubscribeChan:
				// shutdown the resubscriber goroutine on unsubscribe signal
				return
			}
		}
	}()

	// closing the unsubscribeChan will trigger a unsubscribe signal and
	// run unsubscription for all subscription instances
	unsubscribeCallback := func() {
		close(unsubscribeChan)
	}

	return subscription.NewEventSubscription(unsubscribeCallback), nil
}

func (becdsakv *BondedECDSAKeepVendor) subscribeFactoryUpgradeCompleted(
	success bondedECDSAKeepVendorFactoryUpgradeCompletedFunc,
	fail func(err error) error,
) (subscription.EventSubscription, error) {
	eventChan := make(chan *abi.BondedECDSAKeepVendorImplV1FactoryUpgradeCompleted)
	eventSubscription, err := becdsakv.contract.WatchFactoryUpgradeCompleted(
		nil,
		eventChan,
	)
	if err != nil {
		close(eventChan)
		return eventSubscription, fmt.Errorf(
			"error creating watch for FactoryUpgradeCompleted events: [%v]",
			err,
		)
	}

	var subscriptionMutex = &sync.Mutex{}

	go func() {
		for {
			select {
			case event, subscribed := <-eventChan:
				subscriptionMutex.Lock()
				// if eventChan has been closed, it means we have unsubscribed
				if !subscribed {
					subscriptionMutex.Unlock()
					return
				}
				success(
					event.Factory,
					event.Raw.BlockNumber,
				)
				subscriptionMutex.Unlock()
			case ee := <-eventSubscription.Err():
				fail(ee)
				return
			}
		}
	}()

	unsubscribeCallback := func() {
		subscriptionMutex.Lock()
		defer subscriptionMutex.Unlock()

		eventSubscription.Unsubscribe()
		close(eventChan)
	}

	return subscription.NewEventSubscription(unsubscribeCallback), nil
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
	fail func(err error) error,
) (subscription.EventSubscription, error) {
	errorChan := make(chan error)
	unsubscribeChan := make(chan struct{})

	// Delay which must be preserved before a new resubscription attempt.
	// There is no sense to resubscribe immediately after the fail of current
	// subscription because the publisher must have some time to recover.
	retryDelay := 5 * time.Second

	watch := func() {
		failCallback := func(err error) error {
			fail(err)
			errorChan <- err // trigger resubscription signal
			return err
		}

		subscription, err := becdsakv.subscribeFactoryUpgradeStarted(
			success,
			failCallback,
		)
		if err != nil {
			errorChan <- err // trigger resubscription signal
			return
		}

		// wait for unsubscription signal
		<-unsubscribeChan
		subscription.Unsubscribe()
	}

	// trigger the resubscriber goroutine
	go func() {
		go watch() // trigger first subscription

		for {
			select {
			case <-errorChan:
				becdsakvLogger.Warning(
					"subscription to event FactoryUpgradeStarted terminated with error; " +
						"resubscription attempt will be performed after the retry delay",
				)
				time.Sleep(retryDelay)
				go watch()
			case <-unsubscribeChan:
				// shutdown the resubscriber goroutine on unsubscribe signal
				return
			}
		}
	}()

	// closing the unsubscribeChan will trigger a unsubscribe signal and
	// run unsubscription for all subscription instances
	unsubscribeCallback := func() {
		close(unsubscribeChan)
	}

	return subscription.NewEventSubscription(unsubscribeCallback), nil
}

func (becdsakv *BondedECDSAKeepVendor) subscribeFactoryUpgradeStarted(
	success bondedECDSAKeepVendorFactoryUpgradeStartedFunc,
	fail func(err error) error,
) (subscription.EventSubscription, error) {
	eventChan := make(chan *abi.BondedECDSAKeepVendorImplV1FactoryUpgradeStarted)
	eventSubscription, err := becdsakv.contract.WatchFactoryUpgradeStarted(
		nil,
		eventChan,
	)
	if err != nil {
		close(eventChan)
		return eventSubscription, fmt.Errorf(
			"error creating watch for FactoryUpgradeStarted events: [%v]",
			err,
		)
	}

	var subscriptionMutex = &sync.Mutex{}

	go func() {
		for {
			select {
			case event, subscribed := <-eventChan:
				subscriptionMutex.Lock()
				// if eventChan has been closed, it means we have unsubscribed
				if !subscribed {
					subscriptionMutex.Unlock()
					return
				}
				success(
					event.Factory,
					event.Timestamp,
					event.Raw.BlockNumber,
				)
				subscriptionMutex.Unlock()
			case ee := <-eventSubscription.Err():
				fail(ee)
				return
			}
		}
	}()

	unsubscribeCallback := func() {
		subscriptionMutex.Lock()
		defer subscriptionMutex.Unlock()

		eventSubscription.Unsubscribe()
		close(eventChan)
	}

	return subscription.NewEventSubscription(unsubscribeCallback), nil
}
