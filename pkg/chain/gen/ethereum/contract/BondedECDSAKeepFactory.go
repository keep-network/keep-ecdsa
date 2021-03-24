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
var becdsakfLogger = log.Logger("keep-contract-BondedECDSAKeepFactory")

type BondedECDSAKeepFactory struct {
	contract          *abi.BondedECDSAKeepFactory
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

func NewBondedECDSAKeepFactory(
	contractAddress common.Address,
	chainId *big.Int,
	accountKey *keystore.Key,
	backend bind.ContractBackend,
	nonceManager *ethlike.NonceManager,
	miningWaiter *ethlike.MiningWaiter,
	blockCounter *ethlike.BlockCounter,
	transactionMutex *sync.Mutex,
) (*BondedECDSAKeepFactory, error) {
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

	contract, err := abi.NewBondedECDSAKeepFactory(
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

	contractABI, err := hostchainabi.JSON(strings.NewReader(abi.BondedECDSAKeepFactoryABI))
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ABI: [%v]", err)
	}

	return &BondedECDSAKeepFactory{
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
func (becdsakf *BondedECDSAKeepFactory) BeaconCallback(
	_relayEntry *big.Int,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakfLogger.Debug(
		"submitting transaction beaconCallback",
		"params: ",
		fmt.Sprint(
			_relayEntry,
		),
	)

	becdsakf.transactionMutex.Lock()
	defer becdsakf.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsakf.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsakf.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsakf.contract.BeaconCallback(
		transactorOptions,
		_relayEntry,
	)
	if err != nil {
		return transaction, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.transactorOptions.From,
			nil,
			"beaconCallback",
			_relayEntry,
		)
	}

	becdsakfLogger.Infof(
		"submitted transaction beaconCallback with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsakf.miningWaiter.ForceMining(
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsakf.contract.BeaconCallback(
				transactorOptions,
				_relayEntry,
			)
			if err != nil {
				return nil, becdsakf.errorResolver.ResolveError(
					err,
					becdsakf.transactorOptions.From,
					nil,
					"beaconCallback",
					_relayEntry,
				)
			}

			becdsakfLogger.Infof(
				"submitted transaction beaconCallback with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
		},
	)

	becdsakf.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsakf *BondedECDSAKeepFactory) CallBeaconCallback(
	_relayEntry *big.Int,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		becdsakf.transactorOptions.From,
		blockNumber, nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"beaconCallback",
		&result,
		_relayEntry,
	)

	return err
}

func (becdsakf *BondedECDSAKeepFactory) BeaconCallbackGasEstimate(
	_relayEntry *big.Int,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		becdsakf.callerOptions.From,
		becdsakf.contractAddress,
		"beaconCallback",
		becdsakf.contractABI,
		becdsakf.transactor,
		_relayEntry,
	)

	return result, err
}

// Transaction submission.
func (becdsakf *BondedECDSAKeepFactory) CreateSortitionPool(
	_application common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakfLogger.Debug(
		"submitting transaction createSortitionPool",
		"params: ",
		fmt.Sprint(
			_application,
		),
	)

	becdsakf.transactionMutex.Lock()
	defer becdsakf.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsakf.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsakf.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsakf.contract.CreateSortitionPool(
		transactorOptions,
		_application,
	)
	if err != nil {
		return transaction, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.transactorOptions.From,
			nil,
			"createSortitionPool",
			_application,
		)
	}

	becdsakfLogger.Infof(
		"submitted transaction createSortitionPool with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsakf.miningWaiter.ForceMining(
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsakf.contract.CreateSortitionPool(
				transactorOptions,
				_application,
			)
			if err != nil {
				return nil, becdsakf.errorResolver.ResolveError(
					err,
					becdsakf.transactorOptions.From,
					nil,
					"createSortitionPool",
					_application,
				)
			}

			becdsakfLogger.Infof(
				"submitted transaction createSortitionPool with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
		},
	)

	becdsakf.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsakf *BondedECDSAKeepFactory) CallCreateSortitionPool(
	_application common.Address,
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		becdsakf.transactorOptions.From,
		blockNumber, nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"createSortitionPool",
		&result,
		_application,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) CreateSortitionPoolGasEstimate(
	_application common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		becdsakf.callerOptions.From,
		becdsakf.contractAddress,
		"createSortitionPool",
		becdsakf.contractABI,
		becdsakf.transactor,
		_application,
	)

	return result, err
}

// Transaction submission.
func (becdsakf *BondedECDSAKeepFactory) IsRecognized(
	_delegatedAuthorityRecipient common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakfLogger.Debug(
		"submitting transaction isRecognized",
		"params: ",
		fmt.Sprint(
			_delegatedAuthorityRecipient,
		),
	)

	becdsakf.transactionMutex.Lock()
	defer becdsakf.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsakf.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsakf.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsakf.contract.IsRecognized(
		transactorOptions,
		_delegatedAuthorityRecipient,
	)
	if err != nil {
		return transaction, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.transactorOptions.From,
			nil,
			"isRecognized",
			_delegatedAuthorityRecipient,
		)
	}

	becdsakfLogger.Infof(
		"submitted transaction isRecognized with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsakf.miningWaiter.ForceMining(
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsakf.contract.IsRecognized(
				transactorOptions,
				_delegatedAuthorityRecipient,
			)
			if err != nil {
				return nil, becdsakf.errorResolver.ResolveError(
					err,
					becdsakf.transactorOptions.From,
					nil,
					"isRecognized",
					_delegatedAuthorityRecipient,
				)
			}

			becdsakfLogger.Infof(
				"submitted transaction isRecognized with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
		},
	)

	becdsakf.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsakf *BondedECDSAKeepFactory) CallIsRecognized(
	_delegatedAuthorityRecipient common.Address,
	blockNumber *big.Int,
) (bool, error) {
	var result bool

	err := chainutil.CallAtBlock(
		becdsakf.transactorOptions.From,
		blockNumber, nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"isRecognized",
		&result,
		_delegatedAuthorityRecipient,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) IsRecognizedGasEstimate(
	_delegatedAuthorityRecipient common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		becdsakf.callerOptions.From,
		becdsakf.contractAddress,
		"isRecognized",
		becdsakf.contractABI,
		becdsakf.transactor,
		_delegatedAuthorityRecipient,
	)

	return result, err
}

// Transaction submission.
func (becdsakf *BondedECDSAKeepFactory) OpenKeep(
	_groupSize *big.Int,
	_honestThreshold *big.Int,
	_owner common.Address,
	_bond *big.Int,
	_stakeLockDuration *big.Int,
	value *big.Int,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakfLogger.Debug(
		"submitting transaction openKeep",
		"params: ",
		fmt.Sprint(
			_groupSize,
			_honestThreshold,
			_owner,
			_bond,
			_stakeLockDuration,
		),
		"value: ", value,
	)

	becdsakf.transactionMutex.Lock()
	defer becdsakf.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsakf.transactorOptions

	transactorOptions.Value = value

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsakf.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsakf.contract.OpenKeep(
		transactorOptions,
		_groupSize,
		_honestThreshold,
		_owner,
		_bond,
		_stakeLockDuration,
	)
	if err != nil {
		return transaction, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.transactorOptions.From,
			value,
			"openKeep",
			_groupSize,
			_honestThreshold,
			_owner,
			_bond,
			_stakeLockDuration,
		)
	}

	becdsakfLogger.Infof(
		"submitted transaction openKeep with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsakf.miningWaiter.ForceMining(
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsakf.contract.OpenKeep(
				transactorOptions,
				_groupSize,
				_honestThreshold,
				_owner,
				_bond,
				_stakeLockDuration,
			)
			if err != nil {
				return nil, becdsakf.errorResolver.ResolveError(
					err,
					becdsakf.transactorOptions.From,
					value,
					"openKeep",
					_groupSize,
					_honestThreshold,
					_owner,
					_bond,
					_stakeLockDuration,
				)
			}

			becdsakfLogger.Infof(
				"submitted transaction openKeep with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
		},
	)

	becdsakf.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsakf *BondedECDSAKeepFactory) CallOpenKeep(
	_groupSize *big.Int,
	_honestThreshold *big.Int,
	_owner common.Address,
	_bond *big.Int,
	_stakeLockDuration *big.Int,
	value *big.Int,
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		becdsakf.transactorOptions.From,
		blockNumber, value,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"openKeep",
		&result,
		_groupSize,
		_honestThreshold,
		_owner,
		_bond,
		_stakeLockDuration,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) OpenKeepGasEstimate(
	_groupSize *big.Int,
	_honestThreshold *big.Int,
	_owner common.Address,
	_bond *big.Int,
	_stakeLockDuration *big.Int,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		becdsakf.callerOptions.From,
		becdsakf.contractAddress,
		"openKeep",
		becdsakf.contractABI,
		becdsakf.transactor,
		_groupSize,
		_honestThreshold,
		_owner,
		_bond,
		_stakeLockDuration,
	)

	return result, err
}

// Transaction submission.
func (becdsakf *BondedECDSAKeepFactory) RegisterMemberCandidate(
	_application common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakfLogger.Debug(
		"submitting transaction registerMemberCandidate",
		"params: ",
		fmt.Sprint(
			_application,
		),
	)

	becdsakf.transactionMutex.Lock()
	defer becdsakf.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsakf.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsakf.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsakf.contract.RegisterMemberCandidate(
		transactorOptions,
		_application,
	)
	if err != nil {
		return transaction, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.transactorOptions.From,
			nil,
			"registerMemberCandidate",
			_application,
		)
	}

	becdsakfLogger.Infof(
		"submitted transaction registerMemberCandidate with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsakf.miningWaiter.ForceMining(
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsakf.contract.RegisterMemberCandidate(
				transactorOptions,
				_application,
			)
			if err != nil {
				return nil, becdsakf.errorResolver.ResolveError(
					err,
					becdsakf.transactorOptions.From,
					nil,
					"registerMemberCandidate",
					_application,
				)
			}

			becdsakfLogger.Infof(
				"submitted transaction registerMemberCandidate with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
		},
	)

	becdsakf.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsakf *BondedECDSAKeepFactory) CallRegisterMemberCandidate(
	_application common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		becdsakf.transactorOptions.From,
		blockNumber, nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"registerMemberCandidate",
		&result,
		_application,
	)

	return err
}

func (becdsakf *BondedECDSAKeepFactory) RegisterMemberCandidateGasEstimate(
	_application common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		becdsakf.callerOptions.From,
		becdsakf.contractAddress,
		"registerMemberCandidate",
		becdsakf.contractABI,
		becdsakf.transactor,
		_application,
	)

	return result, err
}

// Transaction submission.
func (becdsakf *BondedECDSAKeepFactory) RequestNewGroupSelectionSeed(
	value *big.Int,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakfLogger.Debug(
		"submitting transaction requestNewGroupSelectionSeed",
		"value: ", value,
	)

	becdsakf.transactionMutex.Lock()
	defer becdsakf.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsakf.transactorOptions

	transactorOptions.Value = value

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsakf.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsakf.contract.RequestNewGroupSelectionSeed(
		transactorOptions,
	)
	if err != nil {
		return transaction, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.transactorOptions.From,
			value,
			"requestNewGroupSelectionSeed",
		)
	}

	becdsakfLogger.Infof(
		"submitted transaction requestNewGroupSelectionSeed with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsakf.miningWaiter.ForceMining(
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsakf.contract.RequestNewGroupSelectionSeed(
				transactorOptions,
			)
			if err != nil {
				return nil, becdsakf.errorResolver.ResolveError(
					err,
					becdsakf.transactorOptions.From,
					value,
					"requestNewGroupSelectionSeed",
				)
			}

			becdsakfLogger.Infof(
				"submitted transaction requestNewGroupSelectionSeed with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
		},
	)

	becdsakf.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsakf *BondedECDSAKeepFactory) CallRequestNewGroupSelectionSeed(
	value *big.Int,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		becdsakf.transactorOptions.From,
		blockNumber, value,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"requestNewGroupSelectionSeed",
		&result,
	)

	return err
}

func (becdsakf *BondedECDSAKeepFactory) RequestNewGroupSelectionSeedGasEstimate() (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		becdsakf.callerOptions.From,
		becdsakf.contractAddress,
		"requestNewGroupSelectionSeed",
		becdsakf.contractABI,
		becdsakf.transactor,
	)

	return result, err
}

// Transaction submission.
func (becdsakf *BondedECDSAKeepFactory) SetMinimumBondableValue(
	_minimumBondableValue *big.Int,
	_groupSize *big.Int,
	_honestThreshold *big.Int,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakfLogger.Debug(
		"submitting transaction setMinimumBondableValue",
		"params: ",
		fmt.Sprint(
			_minimumBondableValue,
			_groupSize,
			_honestThreshold,
		),
	)

	becdsakf.transactionMutex.Lock()
	defer becdsakf.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsakf.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsakf.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsakf.contract.SetMinimumBondableValue(
		transactorOptions,
		_minimumBondableValue,
		_groupSize,
		_honestThreshold,
	)
	if err != nil {
		return transaction, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.transactorOptions.From,
			nil,
			"setMinimumBondableValue",
			_minimumBondableValue,
			_groupSize,
			_honestThreshold,
		)
	}

	becdsakfLogger.Infof(
		"submitted transaction setMinimumBondableValue with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsakf.miningWaiter.ForceMining(
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsakf.contract.SetMinimumBondableValue(
				transactorOptions,
				_minimumBondableValue,
				_groupSize,
				_honestThreshold,
			)
			if err != nil {
				return nil, becdsakf.errorResolver.ResolveError(
					err,
					becdsakf.transactorOptions.From,
					nil,
					"setMinimumBondableValue",
					_minimumBondableValue,
					_groupSize,
					_honestThreshold,
				)
			}

			becdsakfLogger.Infof(
				"submitted transaction setMinimumBondableValue with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
		},
	)

	becdsakf.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsakf *BondedECDSAKeepFactory) CallSetMinimumBondableValue(
	_minimumBondableValue *big.Int,
	_groupSize *big.Int,
	_honestThreshold *big.Int,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		becdsakf.transactorOptions.From,
		blockNumber, nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"setMinimumBondableValue",
		&result,
		_minimumBondableValue,
		_groupSize,
		_honestThreshold,
	)

	return err
}

func (becdsakf *BondedECDSAKeepFactory) SetMinimumBondableValueGasEstimate(
	_minimumBondableValue *big.Int,
	_groupSize *big.Int,
	_honestThreshold *big.Int,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		becdsakf.callerOptions.From,
		becdsakf.contractAddress,
		"setMinimumBondableValue",
		becdsakf.contractABI,
		becdsakf.transactor,
		_minimumBondableValue,
		_groupSize,
		_honestThreshold,
	)

	return result, err
}

// Transaction submission.
func (becdsakf *BondedECDSAKeepFactory) UpdateOperatorStatus(
	_operator common.Address,
	_application common.Address,

	transactionOptions ...chainutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakfLogger.Debug(
		"submitting transaction updateOperatorStatus",
		"params: ",
		fmt.Sprint(
			_operator,
			_application,
		),
	)

	becdsakf.transactionMutex.Lock()
	defer becdsakf.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsakf.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsakf.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsakf.contract.UpdateOperatorStatus(
		transactorOptions,
		_operator,
		_application,
	)
	if err != nil {
		return transaction, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.transactorOptions.From,
			nil,
			"updateOperatorStatus",
			_operator,
			_application,
		)
	}

	becdsakfLogger.Infof(
		"submitted transaction updateOperatorStatus with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsakf.miningWaiter.ForceMining(
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsakf.contract.UpdateOperatorStatus(
				transactorOptions,
				_operator,
				_application,
			)
			if err != nil {
				return nil, becdsakf.errorResolver.ResolveError(
					err,
					becdsakf.transactorOptions.From,
					nil,
					"updateOperatorStatus",
					_operator,
					_application,
				)
			}

			becdsakfLogger.Infof(
				"submitted transaction updateOperatorStatus with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
		},
	)

	becdsakf.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsakf *BondedECDSAKeepFactory) CallUpdateOperatorStatus(
	_operator common.Address,
	_application common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := chainutil.CallAtBlock(
		becdsakf.transactorOptions.From,
		blockNumber, nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"updateOperatorStatus",
		&result,
		_operator,
		_application,
	)

	return err
}

func (becdsakf *BondedECDSAKeepFactory) UpdateOperatorStatusGasEstimate(
	_operator common.Address,
	_application common.Address,
) (uint64, error) {
	var result uint64

	result, err := chainutil.EstimateGas(
		becdsakf.callerOptions.From,
		becdsakf.contractAddress,
		"updateOperatorStatus",
		becdsakf.contractABI,
		becdsakf.transactor,
		_operator,
		_application,
	)

	return result, err
}

// ----- Const Methods ------

func (becdsakf *BondedECDSAKeepFactory) BalanceOf(
	_operator common.Address,
) (*big.Int, error) {
	var result *big.Int
	result, err := becdsakf.contract.BalanceOf(
		becdsakf.callerOptions,
		_operator,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"balanceOf",
			_operator,
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) BalanceOfAtBlock(
	_operator common.Address,
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"balanceOf",
		&result,
		_operator,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) CallbackGas() (*big.Int, error) {
	var result *big.Int
	result, err := becdsakf.contract.CallbackGas(
		becdsakf.callerOptions,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"callbackGas",
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) CallbackGasAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"callbackGas",
		&result,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) GetKeepAtIndex(
	index *big.Int,
) (common.Address, error) {
	var result common.Address
	result, err := becdsakf.contract.GetKeepAtIndex(
		becdsakf.callerOptions,
		index,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"getKeepAtIndex",
			index,
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) GetKeepAtIndexAtBlock(
	index *big.Int,
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"getKeepAtIndex",
		&result,
		index,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) GetKeepCount() (*big.Int, error) {
	var result *big.Int
	result, err := becdsakf.contract.GetKeepCount(
		becdsakf.callerOptions,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"getKeepCount",
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) GetKeepCountAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"getKeepCount",
		&result,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) GetKeepOpenedTimestamp(
	_keep common.Address,
) (*big.Int, error) {
	var result *big.Int
	result, err := becdsakf.contract.GetKeepOpenedTimestamp(
		becdsakf.callerOptions,
		_keep,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"getKeepOpenedTimestamp",
			_keep,
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) GetKeepOpenedTimestampAtBlock(
	_keep common.Address,
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"getKeepOpenedTimestamp",
		&result,
		_keep,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) GetSortitionPool(
	_application common.Address,
) (common.Address, error) {
	var result common.Address
	result, err := becdsakf.contract.GetSortitionPool(
		becdsakf.callerOptions,
		_application,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"getSortitionPool",
			_application,
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) GetSortitionPoolAtBlock(
	_application common.Address,
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"getSortitionPool",
		&result,
		_application,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) GetSortitionPoolWeight(
	_application common.Address,
) (*big.Int, error) {
	var result *big.Int
	result, err := becdsakf.contract.GetSortitionPoolWeight(
		becdsakf.callerOptions,
		_application,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"getSortitionPoolWeight",
			_application,
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) GetSortitionPoolWeightAtBlock(
	_application common.Address,
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"getSortitionPoolWeight",
		&result,
		_application,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) GroupSelectionSeed() (*big.Int, error) {
	var result *big.Int
	result, err := becdsakf.contract.GroupSelectionSeed(
		becdsakf.callerOptions,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"groupSelectionSeed",
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) GroupSelectionSeedAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"groupSelectionSeed",
		&result,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) HasMinimumStake(
	_operator common.Address,
) (bool, error) {
	var result bool
	result, err := becdsakf.contract.HasMinimumStake(
		becdsakf.callerOptions,
		_operator,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"hasMinimumStake",
			_operator,
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) HasMinimumStakeAtBlock(
	_operator common.Address,
	blockNumber *big.Int,
) (bool, error) {
	var result bool

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"hasMinimumStake",
		&result,
		_operator,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) IsOperatorAuthorized(
	_operator common.Address,
) (bool, error) {
	var result bool
	result, err := becdsakf.contract.IsOperatorAuthorized(
		becdsakf.callerOptions,
		_operator,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"isOperatorAuthorized",
			_operator,
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) IsOperatorAuthorizedAtBlock(
	_operator common.Address,
	blockNumber *big.Int,
) (bool, error) {
	var result bool

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"isOperatorAuthorized",
		&result,
		_operator,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) IsOperatorEligible(
	_operator common.Address,
	_application common.Address,
) (bool, error) {
	var result bool
	result, err := becdsakf.contract.IsOperatorEligible(
		becdsakf.callerOptions,
		_operator,
		_application,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"isOperatorEligible",
			_operator,
			_application,
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) IsOperatorEligibleAtBlock(
	_operator common.Address,
	_application common.Address,
	blockNumber *big.Int,
) (bool, error) {
	var result bool

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"isOperatorEligible",
		&result,
		_operator,
		_application,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) IsOperatorRegistered(
	_operator common.Address,
	_application common.Address,
) (bool, error) {
	var result bool
	result, err := becdsakf.contract.IsOperatorRegistered(
		becdsakf.callerOptions,
		_operator,
		_application,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"isOperatorRegistered",
			_operator,
			_application,
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) IsOperatorRegisteredAtBlock(
	_operator common.Address,
	_application common.Address,
	blockNumber *big.Int,
) (bool, error) {
	var result bool

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"isOperatorRegistered",
		&result,
		_operator,
		_application,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) IsOperatorUpToDate(
	_operator common.Address,
	_application common.Address,
) (bool, error) {
	var result bool
	result, err := becdsakf.contract.IsOperatorUpToDate(
		becdsakf.callerOptions,
		_operator,
		_application,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"isOperatorUpToDate",
			_operator,
			_application,
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) IsOperatorUpToDateAtBlock(
	_operator common.Address,
	_application common.Address,
	blockNumber *big.Int,
) (bool, error) {
	var result bool

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"isOperatorUpToDate",
		&result,
		_operator,
		_application,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) Keeps(
	arg0 *big.Int,
) (common.Address, error) {
	var result common.Address
	result, err := becdsakf.contract.Keeps(
		becdsakf.callerOptions,
		arg0,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"keeps",
			arg0,
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) KeepsAtBlock(
	arg0 *big.Int,
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"keeps",
		&result,
		arg0,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) MasterKeepAddress() (common.Address, error) {
	var result common.Address
	result, err := becdsakf.contract.MasterKeepAddress(
		becdsakf.callerOptions,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"masterKeepAddress",
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) MasterKeepAddressAtBlock(
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"masterKeepAddress",
		&result,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) MinimumBond() (*big.Int, error) {
	var result *big.Int
	result, err := becdsakf.contract.MinimumBond(
		becdsakf.callerOptions,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"minimumBond",
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) MinimumBondAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"minimumBond",
		&result,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) NewEntryFeeEstimate() (*big.Int, error) {
	var result *big.Int
	result, err := becdsakf.contract.NewEntryFeeEstimate(
		becdsakf.callerOptions,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"newEntryFeeEstimate",
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) NewEntryFeeEstimateAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"newEntryFeeEstimate",
		&result,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) NewGroupSelectionSeedFee() (*big.Int, error) {
	var result *big.Int
	result, err := becdsakf.contract.NewGroupSelectionSeedFee(
		becdsakf.callerOptions,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"newGroupSelectionSeedFee",
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) NewGroupSelectionSeedFeeAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"newGroupSelectionSeedFee",
		&result,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) OpenKeepFeeEstimate() (*big.Int, error) {
	var result *big.Int
	result, err := becdsakf.contract.OpenKeepFeeEstimate(
		becdsakf.callerOptions,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"openKeepFeeEstimate",
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) OpenKeepFeeEstimateAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"openKeepFeeEstimate",
		&result,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) PoolStakeWeightDivisor() (*big.Int, error) {
	var result *big.Int
	result, err := becdsakf.contract.PoolStakeWeightDivisor(
		becdsakf.callerOptions,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"poolStakeWeightDivisor",
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) PoolStakeWeightDivisorAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"poolStakeWeightDivisor",
		&result,
	)

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) ReseedPool() (*big.Int, error) {
	var result *big.Int
	result, err := becdsakf.contract.ReseedPool(
		becdsakf.callerOptions,
	)

	if err != nil {
		return result, becdsakf.errorResolver.ResolveError(
			err,
			becdsakf.callerOptions.From,
			nil,
			"reseedPool",
		)
	}

	return result, err
}

func (becdsakf *BondedECDSAKeepFactory) ReseedPoolAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := chainutil.CallAtBlock(
		becdsakf.callerOptions.From,
		blockNumber,
		nil,
		becdsakf.contractABI,
		becdsakf.caller,
		becdsakf.errorResolver,
		becdsakf.contractAddress,
		"reseedPool",
		&result,
	)

	return result, err
}

// ------ Events -------

func (becdsakf *BondedECDSAKeepFactory) BondedECDSAKeepCreated(
	opts *ethlike.SubscribeOpts,
	keepAddressFilter []common.Address,
	ownerFilter []common.Address,
	applicationFilter []common.Address,
) *BecdsakfBondedECDSAKeepCreatedSubscription {
	if opts == nil {
		opts = new(ethlike.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BecdsakfBondedECDSAKeepCreatedSubscription{
		becdsakf,
		opts,
		keepAddressFilter,
		ownerFilter,
		applicationFilter,
	}
}

type BecdsakfBondedECDSAKeepCreatedSubscription struct {
	contract          *BondedECDSAKeepFactory
	opts              *ethlike.SubscribeOpts
	keepAddressFilter []common.Address
	ownerFilter       []common.Address
	applicationFilter []common.Address
}

type bondedECDSAKeepFactoryBondedECDSAKeepCreatedFunc func(
	KeepAddress common.Address,
	Members []common.Address,
	Owner common.Address,
	Application common.Address,
	HonestThreshold *big.Int,
	blockNumber uint64,
)

func (becdsakcs *BecdsakfBondedECDSAKeepCreatedSubscription) OnEvent(
	handler bondedECDSAKeepFactoryBondedECDSAKeepCreatedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BondedECDSAKeepFactoryBondedECDSAKeepCreated)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.KeepAddress,
					event.Members,
					event.Owner,
					event.Application,
					event.HonestThreshold,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := becdsakcs.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (becdsakcs *BecdsakfBondedECDSAKeepCreatedSubscription) Pipe(
	sink chan *abi.BondedECDSAKeepFactoryBondedECDSAKeepCreated,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(becdsakcs.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := becdsakcs.contract.blockCounter.CurrentBlock()
				if err != nil {
					becdsakfLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - becdsakcs.opts.PastBlocks

				becdsakfLogger.Infof(
					"subscription monitoring fetching past BondedECDSAKeepCreated events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := becdsakcs.contract.PastBondedECDSAKeepCreatedEvents(
					fromBlock,
					nil,
					becdsakcs.keepAddressFilter,
					becdsakcs.ownerFilter,
					becdsakcs.applicationFilter,
				)
				if err != nil {
					becdsakfLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				becdsakfLogger.Infof(
					"subscription monitoring fetched [%v] past BondedECDSAKeepCreated events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := becdsakcs.contract.watchBondedECDSAKeepCreated(
		sink,
		becdsakcs.keepAddressFilter,
		becdsakcs.ownerFilter,
		becdsakcs.applicationFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (becdsakf *BondedECDSAKeepFactory) watchBondedECDSAKeepCreated(
	sink chan *abi.BondedECDSAKeepFactoryBondedECDSAKeepCreated,
	keepAddressFilter []common.Address,
	ownerFilter []common.Address,
	applicationFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return becdsakf.contract.WatchBondedECDSAKeepCreated(
			&bind.WatchOpts{Context: ctx},
			sink,
			keepAddressFilter,
			ownerFilter,
			applicationFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		becdsakfLogger.Errorf(
			"subscription to event BondedECDSAKeepCreated had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		becdsakfLogger.Errorf(
			"subscription to event BondedECDSAKeepCreated failed "+
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

func (becdsakf *BondedECDSAKeepFactory) PastBondedECDSAKeepCreatedEvents(
	startBlock uint64,
	endBlock *uint64,
	keepAddressFilter []common.Address,
	ownerFilter []common.Address,
	applicationFilter []common.Address,
) ([]*abi.BondedECDSAKeepFactoryBondedECDSAKeepCreated, error) {
	iterator, err := becdsakf.contract.FilterBondedECDSAKeepCreated(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		keepAddressFilter,
		ownerFilter,
		applicationFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past BondedECDSAKeepCreated events: [%v]",
			err,
		)
	}

	events := make([]*abi.BondedECDSAKeepFactoryBondedECDSAKeepCreated, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (becdsakf *BondedECDSAKeepFactory) SortitionPoolCreated(
	opts *ethlike.SubscribeOpts,
	applicationFilter []common.Address,
) *BecdsakfSortitionPoolCreatedSubscription {
	if opts == nil {
		opts = new(ethlike.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BecdsakfSortitionPoolCreatedSubscription{
		becdsakf,
		opts,
		applicationFilter,
	}
}

type BecdsakfSortitionPoolCreatedSubscription struct {
	contract          *BondedECDSAKeepFactory
	opts              *ethlike.SubscribeOpts
	applicationFilter []common.Address
}

type bondedECDSAKeepFactorySortitionPoolCreatedFunc func(
	Application common.Address,
	SortitionPool common.Address,
	blockNumber uint64,
)

func (spcs *BecdsakfSortitionPoolCreatedSubscription) OnEvent(
	handler bondedECDSAKeepFactorySortitionPoolCreatedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BondedECDSAKeepFactorySortitionPoolCreated)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Application,
					event.SortitionPool,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := spcs.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (spcs *BecdsakfSortitionPoolCreatedSubscription) Pipe(
	sink chan *abi.BondedECDSAKeepFactorySortitionPoolCreated,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(spcs.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := spcs.contract.blockCounter.CurrentBlock()
				if err != nil {
					becdsakfLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - spcs.opts.PastBlocks

				becdsakfLogger.Infof(
					"subscription monitoring fetching past SortitionPoolCreated events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := spcs.contract.PastSortitionPoolCreatedEvents(
					fromBlock,
					nil,
					spcs.applicationFilter,
				)
				if err != nil {
					becdsakfLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				becdsakfLogger.Infof(
					"subscription monitoring fetched [%v] past SortitionPoolCreated events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := spcs.contract.watchSortitionPoolCreated(
		sink,
		spcs.applicationFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (becdsakf *BondedECDSAKeepFactory) watchSortitionPoolCreated(
	sink chan *abi.BondedECDSAKeepFactorySortitionPoolCreated,
	applicationFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return becdsakf.contract.WatchSortitionPoolCreated(
			&bind.WatchOpts{Context: ctx},
			sink,
			applicationFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		becdsakfLogger.Errorf(
			"subscription to event SortitionPoolCreated had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		becdsakfLogger.Errorf(
			"subscription to event SortitionPoolCreated failed "+
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

func (becdsakf *BondedECDSAKeepFactory) PastSortitionPoolCreatedEvents(
	startBlock uint64,
	endBlock *uint64,
	applicationFilter []common.Address,
) ([]*abi.BondedECDSAKeepFactorySortitionPoolCreated, error) {
	iterator, err := becdsakf.contract.FilterSortitionPoolCreated(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		applicationFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past SortitionPoolCreated events: [%v]",
			err,
		)
	}

	events := make([]*abi.BondedECDSAKeepFactorySortitionPoolCreated, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}
