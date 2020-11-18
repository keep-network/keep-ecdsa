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
var becdsakLogger = log.Logger("keep-contract-BondedECDSAKeep")

type BondedECDSAKeep struct {
	contract          *abi.BondedECDSAKeep
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

func NewBondedECDSAKeep(
	contractAddress common.Address,
	accountKey *keystore.Key,
	backend bind.ContractBackend,
	nonceManager *ethutil.NonceManager,
	miningWaiter *ethutil.MiningWaiter,
	transactionMutex *sync.Mutex,
) (*BondedECDSAKeep, error) {
	callerOptions := &bind.CallOpts{
		From: accountKey.Address,
	}

	transactorOptions := bind.NewKeyedTransactor(
		accountKey.PrivateKey,
	)

	randomBeaconContract, err := abi.NewBondedECDSAKeep(
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

	contractABI, err := ethereumabi.JSON(strings.NewReader(abi.BondedECDSAKeepABI))
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ABI: [%v]", err)
	}

	return &BondedECDSAKeep{
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
func (becdsak *BondedECDSAKeep) ReturnPartialSignerBonds(
	value *big.Int,

	transactionOptions ...ethutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakLogger.Debug(
		"submitting transaction returnPartialSignerBonds",
		"value: ", value,
	)

	becdsak.transactionMutex.Lock()
	defer becdsak.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsak.transactorOptions

	transactorOptions.Value = value

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsak.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsak.contract.ReturnPartialSignerBonds(
		transactorOptions,
	)
	if err != nil {
		return transaction, becdsak.errorResolver.ResolveError(
			err,
			becdsak.transactorOptions.From,
			value,
			"returnPartialSignerBonds",
		)
	}

	becdsakLogger.Infof(
		"submitted transaction returnPartialSignerBonds with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsak.miningWaiter.ForceMining(
		transaction,
		func(newGasPrice *big.Int) (*types.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.ReturnPartialSignerBonds(
				transactorOptions,
			)
			if err != nil {
				return transaction, becdsak.errorResolver.ResolveError(
					err,
					becdsak.transactorOptions.From,
					value,
					"returnPartialSignerBonds",
				)
			}

			becdsakLogger.Infof(
				"submitted transaction returnPartialSignerBonds with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	becdsak.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsak *BondedECDSAKeep) CallReturnPartialSignerBonds(
	value *big.Int,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := ethutil.CallAtBlock(
		becdsak.transactorOptions.From,
		blockNumber, value,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"returnPartialSignerBonds",
		&result,
	)

	return err
}

func (becdsak *BondedECDSAKeep) ReturnPartialSignerBondsGasEstimate() (uint64, error) {
	var result uint64

	result, err := ethutil.EstimateGas(
		becdsak.callerOptions.From,
		becdsak.contractAddress,
		"returnPartialSignerBonds",
		becdsak.contractABI,
		becdsak.transactor,
	)

	return result, err
}

// Transaction submission.
func (becdsak *BondedECDSAKeep) CloseKeep(

	transactionOptions ...ethutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakLogger.Debug(
		"submitting transaction closeKeep",
	)

	becdsak.transactionMutex.Lock()
	defer becdsak.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsak.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsak.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsak.contract.CloseKeep(
		transactorOptions,
	)
	if err != nil {
		return transaction, becdsak.errorResolver.ResolveError(
			err,
			becdsak.transactorOptions.From,
			nil,
			"closeKeep",
		)
	}

	becdsakLogger.Infof(
		"submitted transaction closeKeep with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsak.miningWaiter.ForceMining(
		transaction,
		func(newGasPrice *big.Int) (*types.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.CloseKeep(
				transactorOptions,
			)
			if err != nil {
				return transaction, becdsak.errorResolver.ResolveError(
					err,
					becdsak.transactorOptions.From,
					nil,
					"closeKeep",
				)
			}

			becdsakLogger.Infof(
				"submitted transaction closeKeep with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	becdsak.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsak *BondedECDSAKeep) CallCloseKeep(
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := ethutil.CallAtBlock(
		becdsak.transactorOptions.From,
		blockNumber, nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"closeKeep",
		&result,
	)

	return err
}

func (becdsak *BondedECDSAKeep) CloseKeepGasEstimate() (uint64, error) {
	var result uint64

	result, err := ethutil.EstimateGas(
		becdsak.callerOptions.From,
		becdsak.contractAddress,
		"closeKeep",
		becdsak.contractABI,
		becdsak.transactor,
	)

	return result, err
}

// Transaction submission.
func (becdsak *BondedECDSAKeep) SeizeSignerBonds(

	transactionOptions ...ethutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakLogger.Debug(
		"submitting transaction seizeSignerBonds",
	)

	becdsak.transactionMutex.Lock()
	defer becdsak.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsak.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsak.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsak.contract.SeizeSignerBonds(
		transactorOptions,
	)
	if err != nil {
		return transaction, becdsak.errorResolver.ResolveError(
			err,
			becdsak.transactorOptions.From,
			nil,
			"seizeSignerBonds",
		)
	}

	becdsakLogger.Infof(
		"submitted transaction seizeSignerBonds with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsak.miningWaiter.ForceMining(
		transaction,
		func(newGasPrice *big.Int) (*types.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.SeizeSignerBonds(
				transactorOptions,
			)
			if err != nil {
				return transaction, becdsak.errorResolver.ResolveError(
					err,
					becdsak.transactorOptions.From,
					nil,
					"seizeSignerBonds",
				)
			}

			becdsakLogger.Infof(
				"submitted transaction seizeSignerBonds with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	becdsak.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsak *BondedECDSAKeep) CallSeizeSignerBonds(
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := ethutil.CallAtBlock(
		becdsak.transactorOptions.From,
		blockNumber, nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"seizeSignerBonds",
		&result,
	)

	return err
}

func (becdsak *BondedECDSAKeep) SeizeSignerBondsGasEstimate() (uint64, error) {
	var result uint64

	result, err := ethutil.EstimateGas(
		becdsak.callerOptions.From,
		becdsak.contractAddress,
		"seizeSignerBonds",
		becdsak.contractABI,
		becdsak.transactor,
	)

	return result, err
}

// Transaction submission.
func (becdsak *BondedECDSAKeep) SubmitSignatureFraud(
	_v uint8,
	_r [32]uint8,
	_s [32]uint8,
	_signedDigest [32]uint8,
	_preimage []uint8,

	transactionOptions ...ethutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakLogger.Debug(
		"submitting transaction submitSignatureFraud",
		"params: ",
		fmt.Sprint(
			_v,
			_r,
			_s,
			_signedDigest,
			_preimage,
		),
	)

	becdsak.transactionMutex.Lock()
	defer becdsak.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsak.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsak.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsak.contract.SubmitSignatureFraud(
		transactorOptions,
		_v,
		_r,
		_s,
		_signedDigest,
		_preimage,
	)
	if err != nil {
		return transaction, becdsak.errorResolver.ResolveError(
			err,
			becdsak.transactorOptions.From,
			nil,
			"submitSignatureFraud",
			_v,
			_r,
			_s,
			_signedDigest,
			_preimage,
		)
	}

	becdsakLogger.Infof(
		"submitted transaction submitSignatureFraud with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsak.miningWaiter.ForceMining(
		transaction,
		func(newGasPrice *big.Int) (*types.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.SubmitSignatureFraud(
				transactorOptions,
				_v,
				_r,
				_s,
				_signedDigest,
				_preimage,
			)
			if err != nil {
				return transaction, becdsak.errorResolver.ResolveError(
					err,
					becdsak.transactorOptions.From,
					nil,
					"submitSignatureFraud",
					_v,
					_r,
					_s,
					_signedDigest,
					_preimage,
				)
			}

			becdsakLogger.Infof(
				"submitted transaction submitSignatureFraud with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	becdsak.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsak *BondedECDSAKeep) CallSubmitSignatureFraud(
	_v uint8,
	_r [32]uint8,
	_s [32]uint8,
	_signedDigest [32]uint8,
	_preimage []uint8,
	blockNumber *big.Int,
) (bool, error) {
	var result bool

	err := ethutil.CallAtBlock(
		becdsak.transactorOptions.From,
		blockNumber, nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"submitSignatureFraud",
		&result,
		_v,
		_r,
		_s,
		_signedDigest,
		_preimage,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) SubmitSignatureFraudGasEstimate(
	_v uint8,
	_r [32]uint8,
	_s [32]uint8,
	_signedDigest [32]uint8,
	_preimage []uint8,
) (uint64, error) {
	var result uint64

	result, err := ethutil.EstimateGas(
		becdsak.callerOptions.From,
		becdsak.contractAddress,
		"submitSignatureFraud",
		becdsak.contractABI,
		becdsak.transactor,
		_v,
		_r,
		_s,
		_signedDigest,
		_preimage,
	)

	return result, err
}

// Transaction submission.
func (becdsak *BondedECDSAKeep) Initialize(
	_owner common.Address,
	_members []common.Address,
	_honestThreshold *big.Int,
	_memberStake *big.Int,
	_stakeLockDuration *big.Int,
	_tokenStaking common.Address,
	_keepBonding common.Address,
	_keepFactory common.Address,

	transactionOptions ...ethutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakLogger.Debug(
		"submitting transaction initialize",
		"params: ",
		fmt.Sprint(
			_owner,
			_members,
			_honestThreshold,
			_memberStake,
			_stakeLockDuration,
			_tokenStaking,
			_keepBonding,
			_keepFactory,
		),
	)

	becdsak.transactionMutex.Lock()
	defer becdsak.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsak.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsak.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsak.contract.Initialize(
		transactorOptions,
		_owner,
		_members,
		_honestThreshold,
		_memberStake,
		_stakeLockDuration,
		_tokenStaking,
		_keepBonding,
		_keepFactory,
	)
	if err != nil {
		return transaction, becdsak.errorResolver.ResolveError(
			err,
			becdsak.transactorOptions.From,
			nil,
			"initialize",
			_owner,
			_members,
			_honestThreshold,
			_memberStake,
			_stakeLockDuration,
			_tokenStaking,
			_keepBonding,
			_keepFactory,
		)
	}

	becdsakLogger.Infof(
		"submitted transaction initialize with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsak.miningWaiter.ForceMining(
		transaction,
		func(newGasPrice *big.Int) (*types.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.Initialize(
				transactorOptions,
				_owner,
				_members,
				_honestThreshold,
				_memberStake,
				_stakeLockDuration,
				_tokenStaking,
				_keepBonding,
				_keepFactory,
			)
			if err != nil {
				return transaction, becdsak.errorResolver.ResolveError(
					err,
					becdsak.transactorOptions.From,
					nil,
					"initialize",
					_owner,
					_members,
					_honestThreshold,
					_memberStake,
					_stakeLockDuration,
					_tokenStaking,
					_keepBonding,
					_keepFactory,
				)
			}

			becdsakLogger.Infof(
				"submitted transaction initialize with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	becdsak.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsak *BondedECDSAKeep) CallInitialize(
	_owner common.Address,
	_members []common.Address,
	_honestThreshold *big.Int,
	_memberStake *big.Int,
	_stakeLockDuration *big.Int,
	_tokenStaking common.Address,
	_keepBonding common.Address,
	_keepFactory common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := ethutil.CallAtBlock(
		becdsak.transactorOptions.From,
		blockNumber, nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"initialize",
		&result,
		_owner,
		_members,
		_honestThreshold,
		_memberStake,
		_stakeLockDuration,
		_tokenStaking,
		_keepBonding,
		_keepFactory,
	)

	return err
}

func (becdsak *BondedECDSAKeep) InitializeGasEstimate(
	_owner common.Address,
	_members []common.Address,
	_honestThreshold *big.Int,
	_memberStake *big.Int,
	_stakeLockDuration *big.Int,
	_tokenStaking common.Address,
	_keepBonding common.Address,
	_keepFactory common.Address,
) (uint64, error) {
	var result uint64

	result, err := ethutil.EstimateGas(
		becdsak.callerOptions.From,
		becdsak.contractAddress,
		"initialize",
		becdsak.contractABI,
		becdsak.transactor,
		_owner,
		_members,
		_honestThreshold,
		_memberStake,
		_stakeLockDuration,
		_tokenStaking,
		_keepBonding,
		_keepFactory,
	)

	return result, err
}

// Transaction submission.
func (becdsak *BondedECDSAKeep) Sign(
	_digest [32]uint8,

	transactionOptions ...ethutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakLogger.Debug(
		"submitting transaction sign",
		"params: ",
		fmt.Sprint(
			_digest,
		),
	)

	becdsak.transactionMutex.Lock()
	defer becdsak.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsak.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsak.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsak.contract.Sign(
		transactorOptions,
		_digest,
	)
	if err != nil {
		return transaction, becdsak.errorResolver.ResolveError(
			err,
			becdsak.transactorOptions.From,
			nil,
			"sign",
			_digest,
		)
	}

	becdsakLogger.Infof(
		"submitted transaction sign with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsak.miningWaiter.ForceMining(
		transaction,
		func(newGasPrice *big.Int) (*types.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.Sign(
				transactorOptions,
				_digest,
			)
			if err != nil {
				return transaction, becdsak.errorResolver.ResolveError(
					err,
					becdsak.transactorOptions.From,
					nil,
					"sign",
					_digest,
				)
			}

			becdsakLogger.Infof(
				"submitted transaction sign with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	becdsak.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsak *BondedECDSAKeep) CallSign(
	_digest [32]uint8,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := ethutil.CallAtBlock(
		becdsak.transactorOptions.From,
		blockNumber, nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"sign",
		&result,
		_digest,
	)

	return err
}

func (becdsak *BondedECDSAKeep) SignGasEstimate(
	_digest [32]uint8,
) (uint64, error) {
	var result uint64

	result, err := ethutil.EstimateGas(
		becdsak.callerOptions.From,
		becdsak.contractAddress,
		"sign",
		becdsak.contractABI,
		becdsak.transactor,
		_digest,
	)

	return result, err
}

// Transaction submission.
func (becdsak *BondedECDSAKeep) SubmitPublicKey(
	_publicKey []uint8,

	transactionOptions ...ethutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakLogger.Debug(
		"submitting transaction submitPublicKey",
		"params: ",
		fmt.Sprint(
			_publicKey,
		),
	)

	becdsak.transactionMutex.Lock()
	defer becdsak.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsak.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsak.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsak.contract.SubmitPublicKey(
		transactorOptions,
		_publicKey,
	)
	if err != nil {
		return transaction, becdsak.errorResolver.ResolveError(
			err,
			becdsak.transactorOptions.From,
			nil,
			"submitPublicKey",
			_publicKey,
		)
	}

	becdsakLogger.Infof(
		"submitted transaction submitPublicKey with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsak.miningWaiter.ForceMining(
		transaction,
		func(newGasPrice *big.Int) (*types.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.SubmitPublicKey(
				transactorOptions,
				_publicKey,
			)
			if err != nil {
				return transaction, becdsak.errorResolver.ResolveError(
					err,
					becdsak.transactorOptions.From,
					nil,
					"submitPublicKey",
					_publicKey,
				)
			}

			becdsakLogger.Infof(
				"submitted transaction submitPublicKey with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	becdsak.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsak *BondedECDSAKeep) CallSubmitPublicKey(
	_publicKey []uint8,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := ethutil.CallAtBlock(
		becdsak.transactorOptions.From,
		blockNumber, nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"submitPublicKey",
		&result,
		_publicKey,
	)

	return err
}

func (becdsak *BondedECDSAKeep) SubmitPublicKeyGasEstimate(
	_publicKey []uint8,
) (uint64, error) {
	var result uint64

	result, err := ethutil.EstimateGas(
		becdsak.callerOptions.From,
		becdsak.contractAddress,
		"submitPublicKey",
		becdsak.contractABI,
		becdsak.transactor,
		_publicKey,
	)

	return result, err
}

// Transaction submission.
func (becdsak *BondedECDSAKeep) DistributeERC20Reward(
	_tokenAddress common.Address,
	_value *big.Int,

	transactionOptions ...ethutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakLogger.Debug(
		"submitting transaction distributeERC20Reward",
		"params: ",
		fmt.Sprint(
			_tokenAddress,
			_value,
		),
	)

	becdsak.transactionMutex.Lock()
	defer becdsak.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsak.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsak.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsak.contract.DistributeERC20Reward(
		transactorOptions,
		_tokenAddress,
		_value,
	)
	if err != nil {
		return transaction, becdsak.errorResolver.ResolveError(
			err,
			becdsak.transactorOptions.From,
			nil,
			"distributeERC20Reward",
			_tokenAddress,
			_value,
		)
	}

	becdsakLogger.Infof(
		"submitted transaction distributeERC20Reward with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsak.miningWaiter.ForceMining(
		transaction,
		func(newGasPrice *big.Int) (*types.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.DistributeERC20Reward(
				transactorOptions,
				_tokenAddress,
				_value,
			)
			if err != nil {
				return transaction, becdsak.errorResolver.ResolveError(
					err,
					becdsak.transactorOptions.From,
					nil,
					"distributeERC20Reward",
					_tokenAddress,
					_value,
				)
			}

			becdsakLogger.Infof(
				"submitted transaction distributeERC20Reward with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	becdsak.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsak *BondedECDSAKeep) CallDistributeERC20Reward(
	_tokenAddress common.Address,
	_value *big.Int,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := ethutil.CallAtBlock(
		becdsak.transactorOptions.From,
		blockNumber, nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"distributeERC20Reward",
		&result,
		_tokenAddress,
		_value,
	)

	return err
}

func (becdsak *BondedECDSAKeep) DistributeERC20RewardGasEstimate(
	_tokenAddress common.Address,
	_value *big.Int,
) (uint64, error) {
	var result uint64

	result, err := ethutil.EstimateGas(
		becdsak.callerOptions.From,
		becdsak.contractAddress,
		"distributeERC20Reward",
		becdsak.contractABI,
		becdsak.transactor,
		_tokenAddress,
		_value,
	)

	return result, err
}

// Transaction submission.
func (becdsak *BondedECDSAKeep) DistributeETHReward(
	value *big.Int,

	transactionOptions ...ethutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakLogger.Debug(
		"submitting transaction distributeETHReward",
		"value: ", value,
	)

	becdsak.transactionMutex.Lock()
	defer becdsak.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsak.transactorOptions

	transactorOptions.Value = value

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsak.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsak.contract.DistributeETHReward(
		transactorOptions,
	)
	if err != nil {
		return transaction, becdsak.errorResolver.ResolveError(
			err,
			becdsak.transactorOptions.From,
			value,
			"distributeETHReward",
		)
	}

	becdsakLogger.Infof(
		"submitted transaction distributeETHReward with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsak.miningWaiter.ForceMining(
		transaction,
		func(newGasPrice *big.Int) (*types.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.DistributeETHReward(
				transactorOptions,
			)
			if err != nil {
				return transaction, becdsak.errorResolver.ResolveError(
					err,
					becdsak.transactorOptions.From,
					value,
					"distributeETHReward",
				)
			}

			becdsakLogger.Infof(
				"submitted transaction distributeETHReward with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	becdsak.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsak *BondedECDSAKeep) CallDistributeETHReward(
	value *big.Int,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := ethutil.CallAtBlock(
		becdsak.transactorOptions.From,
		blockNumber, value,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"distributeETHReward",
		&result,
	)

	return err
}

func (becdsak *BondedECDSAKeep) DistributeETHRewardGasEstimate() (uint64, error) {
	var result uint64

	result, err := ethutil.EstimateGas(
		becdsak.callerOptions.From,
		becdsak.contractAddress,
		"distributeETHReward",
		becdsak.contractABI,
		becdsak.transactor,
	)

	return result, err
}

// Transaction submission.
func (becdsak *BondedECDSAKeep) SubmitSignature(
	_r [32]uint8,
	_s [32]uint8,
	_recoveryID uint8,

	transactionOptions ...ethutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakLogger.Debug(
		"submitting transaction submitSignature",
		"params: ",
		fmt.Sprint(
			_r,
			_s,
			_recoveryID,
		),
	)

	becdsak.transactionMutex.Lock()
	defer becdsak.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsak.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsak.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsak.contract.SubmitSignature(
		transactorOptions,
		_r,
		_s,
		_recoveryID,
	)
	if err != nil {
		return transaction, becdsak.errorResolver.ResolveError(
			err,
			becdsak.transactorOptions.From,
			nil,
			"submitSignature",
			_r,
			_s,
			_recoveryID,
		)
	}

	becdsakLogger.Infof(
		"submitted transaction submitSignature with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsak.miningWaiter.ForceMining(
		transaction,
		func(newGasPrice *big.Int) (*types.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.SubmitSignature(
				transactorOptions,
				_r,
				_s,
				_recoveryID,
			)
			if err != nil {
				return transaction, becdsak.errorResolver.ResolveError(
					err,
					becdsak.transactorOptions.From,
					nil,
					"submitSignature",
					_r,
					_s,
					_recoveryID,
				)
			}

			becdsakLogger.Infof(
				"submitted transaction submitSignature with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	becdsak.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsak *BondedECDSAKeep) CallSubmitSignature(
	_r [32]uint8,
	_s [32]uint8,
	_recoveryID uint8,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := ethutil.CallAtBlock(
		becdsak.transactorOptions.From,
		blockNumber, nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"submitSignature",
		&result,
		_r,
		_s,
		_recoveryID,
	)

	return err
}

func (becdsak *BondedECDSAKeep) SubmitSignatureGasEstimate(
	_r [32]uint8,
	_s [32]uint8,
	_recoveryID uint8,
) (uint64, error) {
	var result uint64

	result, err := ethutil.EstimateGas(
		becdsak.callerOptions.From,
		becdsak.contractAddress,
		"submitSignature",
		becdsak.contractABI,
		becdsak.transactor,
		_r,
		_s,
		_recoveryID,
	)

	return result, err
}

// Transaction submission.
func (becdsak *BondedECDSAKeep) Withdraw(
	_member common.Address,

	transactionOptions ...ethutil.TransactionOptions,
) (*types.Transaction, error) {
	becdsakLogger.Debug(
		"submitting transaction withdraw",
		"params: ",
		fmt.Sprint(
			_member,
		),
	)

	becdsak.transactionMutex.Lock()
	defer becdsak.transactionMutex.Unlock()

	// create a copy
	transactorOptions := new(bind.TransactOpts)
	*transactorOptions = *becdsak.transactorOptions

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	nonce, err := becdsak.nonceManager.CurrentNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	transactorOptions.Nonce = new(big.Int).SetUint64(nonce)

	transaction, err := becdsak.contract.Withdraw(
		transactorOptions,
		_member,
	)
	if err != nil {
		return transaction, becdsak.errorResolver.ResolveError(
			err,
			becdsak.transactorOptions.From,
			nil,
			"withdraw",
			_member,
		)
	}

	becdsakLogger.Infof(
		"submitted transaction withdraw with id: [%v] and nonce [%v]",
		transaction.Hash().Hex(),
		transaction.Nonce(),
	)

	go becdsak.miningWaiter.ForceMining(
		transaction,
		func(newGasPrice *big.Int) (*types.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.Withdraw(
				transactorOptions,
				_member,
			)
			if err != nil {
				return transaction, becdsak.errorResolver.ResolveError(
					err,
					becdsak.transactorOptions.From,
					nil,
					"withdraw",
					_member,
				)
			}

			becdsakLogger.Infof(
				"submitted transaction withdraw with id: [%v] and nonce [%v]",
				transaction.Hash().Hex(),
				transaction.Nonce(),
			)

			return transaction, nil
		},
	)

	becdsak.nonceManager.IncrementNonce()

	return transaction, err
}

// Non-mutating call, not a transaction submission.
func (becdsak *BondedECDSAKeep) CallWithdraw(
	_member common.Address,
	blockNumber *big.Int,
) error {
	var result interface{} = nil

	err := ethutil.CallAtBlock(
		becdsak.transactorOptions.From,
		blockNumber, nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"withdraw",
		&result,
		_member,
	)

	return err
}

func (becdsak *BondedECDSAKeep) WithdrawGasEstimate(
	_member common.Address,
) (uint64, error) {
	var result uint64

	result, err := ethutil.EstimateGas(
		becdsak.callerOptions.From,
		becdsak.contractAddress,
		"withdraw",
		becdsak.contractABI,
		becdsak.transactor,
		_member,
	)

	return result, err
}

// ----- Const Methods ------

func (becdsak *BondedECDSAKeep) Members(
	arg0 *big.Int,
) (common.Address, error) {
	var result common.Address
	result, err := becdsak.contract.Members(
		becdsak.callerOptions,
		arg0,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"members",
			arg0,
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) MembersAtBlock(
	arg0 *big.Int,
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"members",
		&result,
		arg0,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) Digest() ([32]uint8, error) {
	var result [32]uint8
	result, err := becdsak.contract.Digest(
		becdsak.callerOptions,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"digest",
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) DigestAtBlock(
	blockNumber *big.Int,
) ([32]uint8, error) {
	var result [32]uint8

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"digest",
		&result,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) Digests(
	arg0 [32]uint8,
) (*big.Int, error) {
	var result *big.Int
	result, err := becdsak.contract.Digests(
		becdsak.callerOptions,
		arg0,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"digests",
			arg0,
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) DigestsAtBlock(
	arg0 [32]uint8,
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"digests",
		&result,
		arg0,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) GetMembers() ([]common.Address, error) {
	var result []common.Address
	result, err := becdsak.contract.GetMembers(
		becdsak.callerOptions,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"getMembers",
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) GetMembersAtBlock(
	blockNumber *big.Int,
) ([]common.Address, error) {
	var result []common.Address

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"getMembers",
		&result,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) GetOwner() (common.Address, error) {
	var result common.Address
	result, err := becdsak.contract.GetOwner(
		becdsak.callerOptions,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"getOwner",
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) GetOwnerAtBlock(
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"getOwner",
		&result,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) IsClosed() (bool, error) {
	var result bool
	result, err := becdsak.contract.IsClosed(
		becdsak.callerOptions,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"isClosed",
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) IsClosedAtBlock(
	blockNumber *big.Int,
) (bool, error) {
	var result bool

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"isClosed",
		&result,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) IsActive() (bool, error) {
	var result bool
	result, err := becdsak.contract.IsActive(
		becdsak.callerOptions,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"isActive",
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) IsActiveAtBlock(
	blockNumber *big.Int,
) (bool, error) {
	var result bool

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"isActive",
		&result,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) IsTerminated() (bool, error) {
	var result bool
	result, err := becdsak.contract.IsTerminated(
		becdsak.callerOptions,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"isTerminated",
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) IsTerminatedAtBlock(
	blockNumber *big.Int,
) (bool, error) {
	var result bool

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"isTerminated",
		&result,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) MemberStake() (*big.Int, error) {
	var result *big.Int
	result, err := becdsak.contract.MemberStake(
		becdsak.callerOptions,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"memberStake",
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) MemberStakeAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"memberStake",
		&result,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) IsAwaitingSignature(
	_digest [32]uint8,
) (bool, error) {
	var result bool
	result, err := becdsak.contract.IsAwaitingSignature(
		becdsak.callerOptions,
		_digest,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"isAwaitingSignature",
			_digest,
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) IsAwaitingSignatureAtBlock(
	_digest [32]uint8,
	blockNumber *big.Int,
) (bool, error) {
	var result bool

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"isAwaitingSignature",
		&result,
		_digest,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) Owner() (common.Address, error) {
	var result common.Address
	result, err := becdsak.contract.Owner(
		becdsak.callerOptions,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"owner",
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) OwnerAtBlock(
	blockNumber *big.Int,
) (common.Address, error) {
	var result common.Address

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"owner",
		&result,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) PublicKey() ([]uint8, error) {
	var result []uint8
	result, err := becdsak.contract.PublicKey(
		becdsak.callerOptions,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"publicKey",
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) PublicKeyAtBlock(
	blockNumber *big.Int,
) ([]uint8, error) {
	var result []uint8

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"publicKey",
		&result,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) CheckSignatureFraud(
	_v uint8,
	_r [32]uint8,
	_s [32]uint8,
	_signedDigest [32]uint8,
	_preimage []uint8,
) (bool, error) {
	var result bool
	result, err := becdsak.contract.CheckSignatureFraud(
		becdsak.callerOptions,
		_v,
		_r,
		_s,
		_signedDigest,
		_preimage,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"checkSignatureFraud",
			_v,
			_r,
			_s,
			_signedDigest,
			_preimage,
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) CheckSignatureFraudAtBlock(
	_v uint8,
	_r [32]uint8,
	_s [32]uint8,
	_signedDigest [32]uint8,
	_preimage []uint8,
	blockNumber *big.Int,
) (bool, error) {
	var result bool

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"checkSignatureFraud",
		&result,
		_v,
		_r,
		_s,
		_signedDigest,
		_preimage,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) HonestThreshold() (*big.Int, error) {
	var result *big.Int
	result, err := becdsak.contract.HonestThreshold(
		becdsak.callerOptions,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"honestThreshold",
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) HonestThresholdAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"honestThreshold",
		&result,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) CheckBondAmount() (*big.Int, error) {
	var result *big.Int
	result, err := becdsak.contract.CheckBondAmount(
		becdsak.callerOptions,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"checkBondAmount",
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) CheckBondAmountAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"checkBondAmount",
		&result,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) GetOpenedTimestamp() (*big.Int, error) {
	var result *big.Int
	result, err := becdsak.contract.GetOpenedTimestamp(
		becdsak.callerOptions,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"getOpenedTimestamp",
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) GetOpenedTimestampAtBlock(
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"getOpenedTimestamp",
		&result,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) GetMemberETHBalance(
	_member common.Address,
) (*big.Int, error) {
	var result *big.Int
	result, err := becdsak.contract.GetMemberETHBalance(
		becdsak.callerOptions,
		_member,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"getMemberETHBalance",
			_member,
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) GetMemberETHBalanceAtBlock(
	_member common.Address,
	blockNumber *big.Int,
) (*big.Int, error) {
	var result *big.Int

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"getMemberETHBalance",
		&result,
		_member,
	)

	return result, err
}

func (becdsak *BondedECDSAKeep) GetPublicKey() ([]uint8, error) {
	var result []uint8
	result, err := becdsak.contract.GetPublicKey(
		becdsak.callerOptions,
	)

	if err != nil {
		return result, becdsak.errorResolver.ResolveError(
			err,
			becdsak.callerOptions.From,
			nil,
			"getPublicKey",
		)
	}

	return result, err
}

func (becdsak *BondedECDSAKeep) GetPublicKeyAtBlock(
	blockNumber *big.Int,
) ([]uint8, error) {
	var result []uint8

	err := ethutil.CallAtBlock(
		becdsak.callerOptions.From,
		blockNumber,
		nil,
		becdsak.contractABI,
		becdsak.caller,
		becdsak.errorResolver,
		becdsak.contractAddress,
		"getPublicKey",
		&result,
	)

	return result, err
}

// ------ Events -------

type bondedECDSAKeepKeepClosedFunc func(
	blockNumber uint64,
)

func (becdsak *BondedECDSAKeep) PastKeepClosedEvents(
	startBlock uint64,
	endBlock *uint64,
) ([]*abi.BondedECDSAKeepKeepClosed, error) {
	iterator, err := becdsak.contract.FilterKeepClosed(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past KeepClosed events: [%v]",
			err,
		)
	}

	events := make([]*abi.BondedECDSAKeepKeepClosed, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (becdsak *BondedECDSAKeep) WatchKeepClosed(
	success bondedECDSAKeepKeepClosedFunc,
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

		subscription, err := becdsak.subscribeKeepClosed(
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
				becdsakLogger.Warning(
					"subscription to event KeepClosed terminated with error; " +
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

func (becdsak *BondedECDSAKeep) subscribeKeepClosed(
	success bondedECDSAKeepKeepClosedFunc,
	fail func(err error) error,
) (subscription.EventSubscription, error) {
	eventChan := make(chan *abi.BondedECDSAKeepKeepClosed)
	eventSubscription, err := becdsak.contract.WatchKeepClosed(
		nil,
		eventChan,
	)
	if err != nil {
		close(eventChan)
		return eventSubscription, fmt.Errorf(
			"error creating watch for KeepClosed events: [%v]",
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

type bondedECDSAKeepKeepTerminatedFunc func(
	blockNumber uint64,
)

func (becdsak *BondedECDSAKeep) PastKeepTerminatedEvents(
	startBlock uint64,
	endBlock *uint64,
) ([]*abi.BondedECDSAKeepKeepTerminated, error) {
	iterator, err := becdsak.contract.FilterKeepTerminated(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past KeepTerminated events: [%v]",
			err,
		)
	}

	events := make([]*abi.BondedECDSAKeepKeepTerminated, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (becdsak *BondedECDSAKeep) WatchKeepTerminated(
	success bondedECDSAKeepKeepTerminatedFunc,
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

		subscription, err := becdsak.subscribeKeepTerminated(
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
				becdsakLogger.Warning(
					"subscription to event KeepTerminated terminated with error; " +
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

func (becdsak *BondedECDSAKeep) subscribeKeepTerminated(
	success bondedECDSAKeepKeepTerminatedFunc,
	fail func(err error) error,
) (subscription.EventSubscription, error) {
	eventChan := make(chan *abi.BondedECDSAKeepKeepTerminated)
	eventSubscription, err := becdsak.contract.WatchKeepTerminated(
		nil,
		eventChan,
	)
	if err != nil {
		close(eventChan)
		return eventSubscription, fmt.Errorf(
			"error creating watch for KeepTerminated events: [%v]",
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

type bondedECDSAKeepSignatureRequestedFunc func(
	Digest [32]uint8,
	blockNumber uint64,
)

func (becdsak *BondedECDSAKeep) PastSignatureRequestedEvents(
	startBlock uint64,
	endBlock *uint64,
	digestFilter [][32]uint8,
) ([]*abi.BondedECDSAKeepSignatureRequested, error) {
	iterator, err := becdsak.contract.FilterSignatureRequested(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		digestFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past SignatureRequested events: [%v]",
			err,
		)
	}

	events := make([]*abi.BondedECDSAKeepSignatureRequested, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (becdsak *BondedECDSAKeep) WatchSignatureRequested(
	success bondedECDSAKeepSignatureRequestedFunc,
	fail func(err error) error,
	digestFilter [][32]uint8,
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

		subscription, err := becdsak.subscribeSignatureRequested(
			success,
			failCallback,
			digestFilter,
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
				becdsakLogger.Warning(
					"subscription to event SignatureRequested terminated with error; " +
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

func (becdsak *BondedECDSAKeep) subscribeSignatureRequested(
	success bondedECDSAKeepSignatureRequestedFunc,
	fail func(err error) error,
	digestFilter [][32]uint8,
) (subscription.EventSubscription, error) {
	eventChan := make(chan *abi.BondedECDSAKeepSignatureRequested)
	eventSubscription, err := becdsak.contract.WatchSignatureRequested(
		nil,
		eventChan,
		digestFilter,
	)
	if err != nil {
		close(eventChan)
		return eventSubscription, fmt.Errorf(
			"error creating watch for SignatureRequested events: [%v]",
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
					event.Digest,
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

type bondedECDSAKeepSignatureSubmittedFunc func(
	Digest [32]uint8,
	R [32]uint8,
	S [32]uint8,
	RecoveryID uint8,
	blockNumber uint64,
)

func (becdsak *BondedECDSAKeep) PastSignatureSubmittedEvents(
	startBlock uint64,
	endBlock *uint64,
	digestFilter [][32]uint8,
) ([]*abi.BondedECDSAKeepSignatureSubmitted, error) {
	iterator, err := becdsak.contract.FilterSignatureSubmitted(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		digestFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past SignatureSubmitted events: [%v]",
			err,
		)
	}

	events := make([]*abi.BondedECDSAKeepSignatureSubmitted, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (becdsak *BondedECDSAKeep) WatchSignatureSubmitted(
	success bondedECDSAKeepSignatureSubmittedFunc,
	fail func(err error) error,
	digestFilter [][32]uint8,
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

		subscription, err := becdsak.subscribeSignatureSubmitted(
			success,
			failCallback,
			digestFilter,
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
				becdsakLogger.Warning(
					"subscription to event SignatureSubmitted terminated with error; " +
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

func (becdsak *BondedECDSAKeep) subscribeSignatureSubmitted(
	success bondedECDSAKeepSignatureSubmittedFunc,
	fail func(err error) error,
	digestFilter [][32]uint8,
) (subscription.EventSubscription, error) {
	eventChan := make(chan *abi.BondedECDSAKeepSignatureSubmitted)
	eventSubscription, err := becdsak.contract.WatchSignatureSubmitted(
		nil,
		eventChan,
		digestFilter,
	)
	if err != nil {
		close(eventChan)
		return eventSubscription, fmt.Errorf(
			"error creating watch for SignatureSubmitted events: [%v]",
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
					event.Digest,
					event.R,
					event.S,
					event.RecoveryID,
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

type bondedECDSAKeepSlashingFailedFunc func(
	blockNumber uint64,
)

func (becdsak *BondedECDSAKeep) PastSlashingFailedEvents(
	startBlock uint64,
	endBlock *uint64,
) ([]*abi.BondedECDSAKeepSlashingFailed, error) {
	iterator, err := becdsak.contract.FilterSlashingFailed(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past SlashingFailed events: [%v]",
			err,
		)
	}

	events := make([]*abi.BondedECDSAKeepSlashingFailed, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (becdsak *BondedECDSAKeep) WatchSlashingFailed(
	success bondedECDSAKeepSlashingFailedFunc,
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

		subscription, err := becdsak.subscribeSlashingFailed(
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
				becdsakLogger.Warning(
					"subscription to event SlashingFailed terminated with error; " +
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

func (becdsak *BondedECDSAKeep) subscribeSlashingFailed(
	success bondedECDSAKeepSlashingFailedFunc,
	fail func(err error) error,
) (subscription.EventSubscription, error) {
	eventChan := make(chan *abi.BondedECDSAKeepSlashingFailed)
	eventSubscription, err := becdsak.contract.WatchSlashingFailed(
		nil,
		eventChan,
	)
	if err != nil {
		close(eventChan)
		return eventSubscription, fmt.Errorf(
			"error creating watch for SlashingFailed events: [%v]",
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

type bondedECDSAKeepConflictingPublicKeySubmittedFunc func(
	SubmittingMember common.Address,
	ConflictingPublicKey []uint8,
	blockNumber uint64,
)

func (becdsak *BondedECDSAKeep) PastConflictingPublicKeySubmittedEvents(
	startBlock uint64,
	endBlock *uint64,
	submittingMemberFilter []common.Address,
) ([]*abi.BondedECDSAKeepConflictingPublicKeySubmitted, error) {
	iterator, err := becdsak.contract.FilterConflictingPublicKeySubmitted(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		submittingMemberFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past ConflictingPublicKeySubmitted events: [%v]",
			err,
		)
	}

	events := make([]*abi.BondedECDSAKeepConflictingPublicKeySubmitted, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (becdsak *BondedECDSAKeep) WatchConflictingPublicKeySubmitted(
	success bondedECDSAKeepConflictingPublicKeySubmittedFunc,
	fail func(err error) error,
	submittingMemberFilter []common.Address,
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

		subscription, err := becdsak.subscribeConflictingPublicKeySubmitted(
			success,
			failCallback,
			submittingMemberFilter,
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
				becdsakLogger.Warning(
					"subscription to event ConflictingPublicKeySubmitted terminated with error; " +
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

func (becdsak *BondedECDSAKeep) subscribeConflictingPublicKeySubmitted(
	success bondedECDSAKeepConflictingPublicKeySubmittedFunc,
	fail func(err error) error,
	submittingMemberFilter []common.Address,
) (subscription.EventSubscription, error) {
	eventChan := make(chan *abi.BondedECDSAKeepConflictingPublicKeySubmitted)
	eventSubscription, err := becdsak.contract.WatchConflictingPublicKeySubmitted(
		nil,
		eventChan,
		submittingMemberFilter,
	)
	if err != nil {
		close(eventChan)
		return eventSubscription, fmt.Errorf(
			"error creating watch for ConflictingPublicKeySubmitted events: [%v]",
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
					event.SubmittingMember,
					event.ConflictingPublicKey,
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

type bondedECDSAKeepERC20RewardDistributedFunc func(
	Token common.Address,
	Amount *big.Int,
	blockNumber uint64,
)

func (becdsak *BondedECDSAKeep) PastERC20RewardDistributedEvents(
	startBlock uint64,
	endBlock *uint64,
	tokenFilter []common.Address,
) ([]*abi.BondedECDSAKeepERC20RewardDistributed, error) {
	iterator, err := becdsak.contract.FilterERC20RewardDistributed(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		tokenFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past ERC20RewardDistributed events: [%v]",
			err,
		)
	}

	events := make([]*abi.BondedECDSAKeepERC20RewardDistributed, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (becdsak *BondedECDSAKeep) WatchERC20RewardDistributed(
	success bondedECDSAKeepERC20RewardDistributedFunc,
	fail func(err error) error,
	tokenFilter []common.Address,
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

		subscription, err := becdsak.subscribeERC20RewardDistributed(
			success,
			failCallback,
			tokenFilter,
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
				becdsakLogger.Warning(
					"subscription to event ERC20RewardDistributed terminated with error; " +
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

func (becdsak *BondedECDSAKeep) subscribeERC20RewardDistributed(
	success bondedECDSAKeepERC20RewardDistributedFunc,
	fail func(err error) error,
	tokenFilter []common.Address,
) (subscription.EventSubscription, error) {
	eventChan := make(chan *abi.BondedECDSAKeepERC20RewardDistributed)
	eventSubscription, err := becdsak.contract.WatchERC20RewardDistributed(
		nil,
		eventChan,
		tokenFilter,
	)
	if err != nil {
		close(eventChan)
		return eventSubscription, fmt.Errorf(
			"error creating watch for ERC20RewardDistributed events: [%v]",
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
					event.Token,
					event.Amount,
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

type bondedECDSAKeepETHRewardDistributedFunc func(
	Amount *big.Int,
	blockNumber uint64,
)

func (becdsak *BondedECDSAKeep) PastETHRewardDistributedEvents(
	startBlock uint64,
	endBlock *uint64,
) ([]*abi.BondedECDSAKeepETHRewardDistributed, error) {
	iterator, err := becdsak.contract.FilterETHRewardDistributed(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past ETHRewardDistributed events: [%v]",
			err,
		)
	}

	events := make([]*abi.BondedECDSAKeepETHRewardDistributed, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (becdsak *BondedECDSAKeep) WatchETHRewardDistributed(
	success bondedECDSAKeepETHRewardDistributedFunc,
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

		subscription, err := becdsak.subscribeETHRewardDistributed(
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
				becdsakLogger.Warning(
					"subscription to event ETHRewardDistributed terminated with error; " +
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

func (becdsak *BondedECDSAKeep) subscribeETHRewardDistributed(
	success bondedECDSAKeepETHRewardDistributedFunc,
	fail func(err error) error,
) (subscription.EventSubscription, error) {
	eventChan := make(chan *abi.BondedECDSAKeepETHRewardDistributed)
	eventSubscription, err := becdsak.contract.WatchETHRewardDistributed(
		nil,
		eventChan,
	)
	if err != nil {
		close(eventChan)
		return eventSubscription, fmt.Errorf(
			"error creating watch for ETHRewardDistributed events: [%v]",
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
					event.Amount,
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

type bondedECDSAKeepPublicKeyPublishedFunc func(
	PublicKey []uint8,
	blockNumber uint64,
)

func (becdsak *BondedECDSAKeep) PastPublicKeyPublishedEvents(
	startBlock uint64,
	endBlock *uint64,
) ([]*abi.BondedECDSAKeepPublicKeyPublished, error) {
	iterator, err := becdsak.contract.FilterPublicKeyPublished(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving past PublicKeyPublished events: [%v]",
			err,
		)
	}

	events := make([]*abi.BondedECDSAKeepPublicKeyPublished, 0)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
	}

	return events, nil
}

func (becdsak *BondedECDSAKeep) WatchPublicKeyPublished(
	success bondedECDSAKeepPublicKeyPublishedFunc,
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

		subscription, err := becdsak.subscribePublicKeyPublished(
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
				becdsakLogger.Warning(
					"subscription to event PublicKeyPublished terminated with error; " +
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

func (becdsak *BondedECDSAKeep) subscribePublicKeyPublished(
	success bondedECDSAKeepPublicKeyPublishedFunc,
	fail func(err error) error,
) (subscription.EventSubscription, error) {
	eventChan := make(chan *abi.BondedECDSAKeepPublicKeyPublished)
	eventSubscription, err := becdsak.contract.WatchPublicKeyPublished(
		nil,
		eventChan,
	)
	if err != nil {
		close(eventChan)
		return eventSubscription, fmt.Errorf(
			"error creating watch for PublicKeyPublished events: [%v]",
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
					event.PublicKey,
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
