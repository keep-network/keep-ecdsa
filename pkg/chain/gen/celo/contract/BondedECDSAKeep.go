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

	hostchainabi "github.com/celo-org/celo-blockchain/accounts/abi"
	"github.com/celo-org/celo-blockchain/accounts/abi/bind"
	"github.com/celo-org/celo-blockchain/accounts/keystore"
	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/celo-org/celo-blockchain/event"

	"github.com/ipfs/go-log"

	chainutil "github.com/keep-network/keep-common/pkg/chain/celo/celoutil"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/celo/abi"
)

// Create a package-level logger for this contract. The logger exists at
// package level so that the logger is registered at startup and can be
// included or excluded from logging at startup by name.
var becdsakLogger = log.Logger("keep-contract-BondedECDSAKeep")

type BondedECDSAKeep struct {
	contract          *abi.BondedECDSAKeep
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

func NewBondedECDSAKeep(
	contractAddress common.Address,
	chainId *big.Int,
	accountKey *keystore.Key,
	backend bind.ContractBackend,
	nonceManager *ethlike.NonceManager,
	miningWaiter *ethlike.MiningWaiter,
	blockCounter *ethlike.BlockCounter,
	transactionMutex *sync.Mutex,
) (*BondedECDSAKeep, error) {
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

	contract, err := abi.NewBondedECDSAKeep(
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

	contractABI, err := hostchainabi.JSON(strings.NewReader(abi.BondedECDSAKeepABI))
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ABI: [%v]", err)
	}

	return &BondedECDSAKeep{
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
func (becdsak *BondedECDSAKeep) CloseKeep(

	transactionOptions ...chainutil.TransactionOptions,
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
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.CloseKeep(
				transactorOptions,
			)
			if err != nil {
				return nil, becdsak.errorResolver.ResolveError(
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

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
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

	err := chainutil.CallAtBlock(
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

	result, err := chainutil.EstimateGas(
		becdsak.callerOptions.From,
		becdsak.contractAddress,
		"closeKeep",
		becdsak.contractABI,
		becdsak.transactor,
	)

	return result, err
}

// Transaction submission.
func (becdsak *BondedECDSAKeep) DistributeERC20Reward(
	_tokenAddress common.Address,
	_value *big.Int,

	transactionOptions ...chainutil.TransactionOptions,
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
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.DistributeERC20Reward(
				transactorOptions,
				_tokenAddress,
				_value,
			)
			if err != nil {
				return nil, becdsak.errorResolver.ResolveError(
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

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
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

	err := chainutil.CallAtBlock(
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

	result, err := chainutil.EstimateGas(
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

	transactionOptions ...chainutil.TransactionOptions,
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
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.DistributeETHReward(
				transactorOptions,
			)
			if err != nil {
				return nil, becdsak.errorResolver.ResolveError(
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

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
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

	err := chainutil.CallAtBlock(
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

	result, err := chainutil.EstimateGas(
		becdsak.callerOptions.From,
		becdsak.contractAddress,
		"distributeETHReward",
		becdsak.contractABI,
		becdsak.transactor,
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

	transactionOptions ...chainutil.TransactionOptions,
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
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
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
				return nil, becdsak.errorResolver.ResolveError(
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

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
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

	err := chainutil.CallAtBlock(
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

	result, err := chainutil.EstimateGas(
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
func (becdsak *BondedECDSAKeep) ReturnPartialSignerBonds(
	value *big.Int,

	transactionOptions ...chainutil.TransactionOptions,
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
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.ReturnPartialSignerBonds(
				transactorOptions,
			)
			if err != nil {
				return nil, becdsak.errorResolver.ResolveError(
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

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
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

	err := chainutil.CallAtBlock(
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

	result, err := chainutil.EstimateGas(
		becdsak.callerOptions.From,
		becdsak.contractAddress,
		"returnPartialSignerBonds",
		becdsak.contractABI,
		becdsak.transactor,
	)

	return result, err
}

// Transaction submission.
func (becdsak *BondedECDSAKeep) SeizeSignerBonds(

	transactionOptions ...chainutil.TransactionOptions,
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
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.SeizeSignerBonds(
				transactorOptions,
			)
			if err != nil {
				return nil, becdsak.errorResolver.ResolveError(
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

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
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

	err := chainutil.CallAtBlock(
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

	result, err := chainutil.EstimateGas(
		becdsak.callerOptions.From,
		becdsak.contractAddress,
		"seizeSignerBonds",
		becdsak.contractABI,
		becdsak.transactor,
	)

	return result, err
}

// Transaction submission.
func (becdsak *BondedECDSAKeep) Sign(
	_digest [32]uint8,

	transactionOptions ...chainutil.TransactionOptions,
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
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.Sign(
				transactorOptions,
				_digest,
			)
			if err != nil {
				return nil, becdsak.errorResolver.ResolveError(
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

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
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

	err := chainutil.CallAtBlock(
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

	result, err := chainutil.EstimateGas(
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

	transactionOptions ...chainutil.TransactionOptions,
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
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.SubmitPublicKey(
				transactorOptions,
				_publicKey,
			)
			if err != nil {
				return nil, becdsak.errorResolver.ResolveError(
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

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
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

	err := chainutil.CallAtBlock(
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

	result, err := chainutil.EstimateGas(
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
func (becdsak *BondedECDSAKeep) SubmitSignature(
	_r [32]uint8,
	_s [32]uint8,
	_recoveryID uint8,

	transactionOptions ...chainutil.TransactionOptions,
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
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.SubmitSignature(
				transactorOptions,
				_r,
				_s,
				_recoveryID,
			)
			if err != nil {
				return nil, becdsak.errorResolver.ResolveError(
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

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
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

	err := chainutil.CallAtBlock(
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

	result, err := chainutil.EstimateGas(
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
func (becdsak *BondedECDSAKeep) SubmitSignatureFraud(
	_v uint8,
	_r [32]uint8,
	_s [32]uint8,
	_signedDigest [32]uint8,
	_preimage []uint8,

	transactionOptions ...chainutil.TransactionOptions,
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
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
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
				return nil, becdsak.errorResolver.ResolveError(
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

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
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

	err := chainutil.CallAtBlock(
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

	result, err := chainutil.EstimateGas(
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
func (becdsak *BondedECDSAKeep) Withdraw(
	_member common.Address,

	transactionOptions ...chainutil.TransactionOptions,
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
		&ethlike.Transaction{
			Hash:     ethlike.Hash(transaction.Hash()),
			GasPrice: transaction.GasPrice(),
		},
		func(newGasPrice *big.Int) (*ethlike.Transaction, error) {
			transactorOptions.GasLimit = transaction.Gas()
			transactorOptions.GasPrice = newGasPrice

			transaction, err := becdsak.contract.Withdraw(
				transactorOptions,
				_member,
			)
			if err != nil {
				return nil, becdsak.errorResolver.ResolveError(
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

			return &ethlike.Transaction{
				Hash:     ethlike.Hash(transaction.Hash()),
				GasPrice: transaction.GasPrice(),
			}, nil
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

	err := chainutil.CallAtBlock(
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

	result, err := chainutil.EstimateGas(
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

	err := chainutil.CallAtBlock(
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

	err := chainutil.CallAtBlock(
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

	err := chainutil.CallAtBlock(
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

	err := chainutil.CallAtBlock(
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

	err := chainutil.CallAtBlock(
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

	err := chainutil.CallAtBlock(
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

	err := chainutil.CallAtBlock(
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

	err := chainutil.CallAtBlock(
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

	err := chainutil.CallAtBlock(
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

	err := chainutil.CallAtBlock(
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

	err := chainutil.CallAtBlock(
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

	err := chainutil.CallAtBlock(
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

	err := chainutil.CallAtBlock(
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

	err := chainutil.CallAtBlock(
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

	err := chainutil.CallAtBlock(
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

	err := chainutil.CallAtBlock(
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

	err := chainutil.CallAtBlock(
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

	err := chainutil.CallAtBlock(
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

// ------ Events -------

func (becdsak *BondedECDSAKeep) ConflictingPublicKeySubmitted(
	opts *ethlike.SubscribeOpts,
	submittingMemberFilter []common.Address,
) *BecdsakConflictingPublicKeySubmittedSubscription {
	if opts == nil {
		opts = new(ethlike.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BecdsakConflictingPublicKeySubmittedSubscription{
		becdsak,
		opts,
		submittingMemberFilter,
	}
}

type BecdsakConflictingPublicKeySubmittedSubscription struct {
	contract               *BondedECDSAKeep
	opts                   *ethlike.SubscribeOpts
	submittingMemberFilter []common.Address
}

type bondedECDSAKeepConflictingPublicKeySubmittedFunc func(
	SubmittingMember common.Address,
	ConflictingPublicKey []uint8,
	blockNumber uint64,
)

func (cpkss *BecdsakConflictingPublicKeySubmittedSubscription) OnEvent(
	handler bondedECDSAKeepConflictingPublicKeySubmittedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BondedECDSAKeepConflictingPublicKeySubmitted)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.SubmittingMember,
					event.ConflictingPublicKey,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := cpkss.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (cpkss *BecdsakConflictingPublicKeySubmittedSubscription) Pipe(
	sink chan *abi.BondedECDSAKeepConflictingPublicKeySubmitted,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(cpkss.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := cpkss.contract.blockCounter.CurrentBlock()
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - cpkss.opts.PastBlocks

				becdsakLogger.Infof(
					"subscription monitoring fetching past ConflictingPublicKeySubmitted events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := cpkss.contract.PastConflictingPublicKeySubmittedEvents(
					fromBlock,
					nil,
					cpkss.submittingMemberFilter,
				)
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				becdsakLogger.Infof(
					"subscription monitoring fetched [%v] past ConflictingPublicKeySubmitted events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := cpkss.contract.watchConflictingPublicKeySubmitted(
		sink,
		cpkss.submittingMemberFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (becdsak *BondedECDSAKeep) watchConflictingPublicKeySubmitted(
	sink chan *abi.BondedECDSAKeepConflictingPublicKeySubmitted,
	submittingMemberFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return becdsak.contract.WatchConflictingPublicKeySubmitted(
			&bind.WatchOpts{Context: ctx},
			sink,
			submittingMemberFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		becdsakLogger.Errorf(
			"subscription to event ConflictingPublicKeySubmitted had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		becdsakLogger.Errorf(
			"subscription to event ConflictingPublicKeySubmitted failed "+
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

func (becdsak *BondedECDSAKeep) ERC20RewardDistributed(
	opts *ethlike.SubscribeOpts,
	tokenFilter []common.Address,
) *BecdsakERC20RewardDistributedSubscription {
	if opts == nil {
		opts = new(ethlike.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BecdsakERC20RewardDistributedSubscription{
		becdsak,
		opts,
		tokenFilter,
	}
}

type BecdsakERC20RewardDistributedSubscription struct {
	contract    *BondedECDSAKeep
	opts        *ethlike.SubscribeOpts
	tokenFilter []common.Address
}

type bondedECDSAKeepERC20RewardDistributedFunc func(
	Token common.Address,
	Amount *big.Int,
	blockNumber uint64,
)

func (ercrds *BecdsakERC20RewardDistributedSubscription) OnEvent(
	handler bondedECDSAKeepERC20RewardDistributedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BondedECDSAKeepERC20RewardDistributed)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Token,
					event.Amount,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := ercrds.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (ercrds *BecdsakERC20RewardDistributedSubscription) Pipe(
	sink chan *abi.BondedECDSAKeepERC20RewardDistributed,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(ercrds.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := ercrds.contract.blockCounter.CurrentBlock()
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - ercrds.opts.PastBlocks

				becdsakLogger.Infof(
					"subscription monitoring fetching past ERC20RewardDistributed events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := ercrds.contract.PastERC20RewardDistributedEvents(
					fromBlock,
					nil,
					ercrds.tokenFilter,
				)
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				becdsakLogger.Infof(
					"subscription monitoring fetched [%v] past ERC20RewardDistributed events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := ercrds.contract.watchERC20RewardDistributed(
		sink,
		ercrds.tokenFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (becdsak *BondedECDSAKeep) watchERC20RewardDistributed(
	sink chan *abi.BondedECDSAKeepERC20RewardDistributed,
	tokenFilter []common.Address,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return becdsak.contract.WatchERC20RewardDistributed(
			&bind.WatchOpts{Context: ctx},
			sink,
			tokenFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		becdsakLogger.Errorf(
			"subscription to event ERC20RewardDistributed had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		becdsakLogger.Errorf(
			"subscription to event ERC20RewardDistributed failed "+
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

func (becdsak *BondedECDSAKeep) ETHRewardDistributed(
	opts *ethlike.SubscribeOpts,
) *BecdsakETHRewardDistributedSubscription {
	if opts == nil {
		opts = new(ethlike.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BecdsakETHRewardDistributedSubscription{
		becdsak,
		opts,
	}
}

type BecdsakETHRewardDistributedSubscription struct {
	contract *BondedECDSAKeep
	opts     *ethlike.SubscribeOpts
}

type bondedECDSAKeepETHRewardDistributedFunc func(
	Amount *big.Int,
	blockNumber uint64,
)

func (ethrds *BecdsakETHRewardDistributedSubscription) OnEvent(
	handler bondedECDSAKeepETHRewardDistributedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BondedECDSAKeepETHRewardDistributed)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Amount,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := ethrds.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (ethrds *BecdsakETHRewardDistributedSubscription) Pipe(
	sink chan *abi.BondedECDSAKeepETHRewardDistributed,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(ethrds.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := ethrds.contract.blockCounter.CurrentBlock()
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - ethrds.opts.PastBlocks

				becdsakLogger.Infof(
					"subscription monitoring fetching past ETHRewardDistributed events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := ethrds.contract.PastETHRewardDistributedEvents(
					fromBlock,
					nil,
				)
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				becdsakLogger.Infof(
					"subscription monitoring fetched [%v] past ETHRewardDistributed events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := ethrds.contract.watchETHRewardDistributed(
		sink,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (becdsak *BondedECDSAKeep) watchETHRewardDistributed(
	sink chan *abi.BondedECDSAKeepETHRewardDistributed,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return becdsak.contract.WatchETHRewardDistributed(
			&bind.WatchOpts{Context: ctx},
			sink,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		becdsakLogger.Errorf(
			"subscription to event ETHRewardDistributed had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		becdsakLogger.Errorf(
			"subscription to event ETHRewardDistributed failed "+
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

func (becdsak *BondedECDSAKeep) KeepClosed(
	opts *ethlike.SubscribeOpts,
) *BecdsakKeepClosedSubscription {
	if opts == nil {
		opts = new(ethlike.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BecdsakKeepClosedSubscription{
		becdsak,
		opts,
	}
}

type BecdsakKeepClosedSubscription struct {
	contract *BondedECDSAKeep
	opts     *ethlike.SubscribeOpts
}

type bondedECDSAKeepKeepClosedFunc func(
	blockNumber uint64,
)

func (kcs *BecdsakKeepClosedSubscription) OnEvent(
	handler bondedECDSAKeepKeepClosedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BondedECDSAKeepKeepClosed)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := kcs.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (kcs *BecdsakKeepClosedSubscription) Pipe(
	sink chan *abi.BondedECDSAKeepKeepClosed,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(kcs.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := kcs.contract.blockCounter.CurrentBlock()
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - kcs.opts.PastBlocks

				becdsakLogger.Infof(
					"subscription monitoring fetching past KeepClosed events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := kcs.contract.PastKeepClosedEvents(
					fromBlock,
					nil,
				)
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				becdsakLogger.Infof(
					"subscription monitoring fetched [%v] past KeepClosed events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := kcs.contract.watchKeepClosed(
		sink,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (becdsak *BondedECDSAKeep) watchKeepClosed(
	sink chan *abi.BondedECDSAKeepKeepClosed,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return becdsak.contract.WatchKeepClosed(
			&bind.WatchOpts{Context: ctx},
			sink,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		becdsakLogger.Errorf(
			"subscription to event KeepClosed had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		becdsakLogger.Errorf(
			"subscription to event KeepClosed failed "+
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

func (becdsak *BondedECDSAKeep) KeepTerminated(
	opts *ethlike.SubscribeOpts,
) *BecdsakKeepTerminatedSubscription {
	if opts == nil {
		opts = new(ethlike.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BecdsakKeepTerminatedSubscription{
		becdsak,
		opts,
	}
}

type BecdsakKeepTerminatedSubscription struct {
	contract *BondedECDSAKeep
	opts     *ethlike.SubscribeOpts
}

type bondedECDSAKeepKeepTerminatedFunc func(
	blockNumber uint64,
)

func (kts *BecdsakKeepTerminatedSubscription) OnEvent(
	handler bondedECDSAKeepKeepTerminatedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BondedECDSAKeepKeepTerminated)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := kts.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (kts *BecdsakKeepTerminatedSubscription) Pipe(
	sink chan *abi.BondedECDSAKeepKeepTerminated,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(kts.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := kts.contract.blockCounter.CurrentBlock()
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - kts.opts.PastBlocks

				becdsakLogger.Infof(
					"subscription monitoring fetching past KeepTerminated events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := kts.contract.PastKeepTerminatedEvents(
					fromBlock,
					nil,
				)
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				becdsakLogger.Infof(
					"subscription monitoring fetched [%v] past KeepTerminated events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := kts.contract.watchKeepTerminated(
		sink,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (becdsak *BondedECDSAKeep) watchKeepTerminated(
	sink chan *abi.BondedECDSAKeepKeepTerminated,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return becdsak.contract.WatchKeepTerminated(
			&bind.WatchOpts{Context: ctx},
			sink,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		becdsakLogger.Errorf(
			"subscription to event KeepTerminated had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		becdsakLogger.Errorf(
			"subscription to event KeepTerminated failed "+
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

func (becdsak *BondedECDSAKeep) PublicKeyPublished(
	opts *ethlike.SubscribeOpts,
) *BecdsakPublicKeyPublishedSubscription {
	if opts == nil {
		opts = new(ethlike.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BecdsakPublicKeyPublishedSubscription{
		becdsak,
		opts,
	}
}

type BecdsakPublicKeyPublishedSubscription struct {
	contract *BondedECDSAKeep
	opts     *ethlike.SubscribeOpts
}

type bondedECDSAKeepPublicKeyPublishedFunc func(
	PublicKey []uint8,
	blockNumber uint64,
)

func (pkps *BecdsakPublicKeyPublishedSubscription) OnEvent(
	handler bondedECDSAKeepPublicKeyPublishedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BondedECDSAKeepPublicKeyPublished)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.PublicKey,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := pkps.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (pkps *BecdsakPublicKeyPublishedSubscription) Pipe(
	sink chan *abi.BondedECDSAKeepPublicKeyPublished,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(pkps.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := pkps.contract.blockCounter.CurrentBlock()
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - pkps.opts.PastBlocks

				becdsakLogger.Infof(
					"subscription monitoring fetching past PublicKeyPublished events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := pkps.contract.PastPublicKeyPublishedEvents(
					fromBlock,
					nil,
				)
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				becdsakLogger.Infof(
					"subscription monitoring fetched [%v] past PublicKeyPublished events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := pkps.contract.watchPublicKeyPublished(
		sink,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (becdsak *BondedECDSAKeep) watchPublicKeyPublished(
	sink chan *abi.BondedECDSAKeepPublicKeyPublished,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return becdsak.contract.WatchPublicKeyPublished(
			&bind.WatchOpts{Context: ctx},
			sink,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		becdsakLogger.Errorf(
			"subscription to event PublicKeyPublished had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		becdsakLogger.Errorf(
			"subscription to event PublicKeyPublished failed "+
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

func (becdsak *BondedECDSAKeep) SignatureRequested(
	opts *ethlike.SubscribeOpts,
	digestFilter [][32]uint8,
) *BecdsakSignatureRequestedSubscription {
	if opts == nil {
		opts = new(ethlike.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BecdsakSignatureRequestedSubscription{
		becdsak,
		opts,
		digestFilter,
	}
}

type BecdsakSignatureRequestedSubscription struct {
	contract     *BondedECDSAKeep
	opts         *ethlike.SubscribeOpts
	digestFilter [][32]uint8
}

type bondedECDSAKeepSignatureRequestedFunc func(
	Digest [32]uint8,
	blockNumber uint64,
)

func (srs *BecdsakSignatureRequestedSubscription) OnEvent(
	handler bondedECDSAKeepSignatureRequestedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BondedECDSAKeepSignatureRequested)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Digest,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := srs.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (srs *BecdsakSignatureRequestedSubscription) Pipe(
	sink chan *abi.BondedECDSAKeepSignatureRequested,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(srs.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := srs.contract.blockCounter.CurrentBlock()
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - srs.opts.PastBlocks

				becdsakLogger.Infof(
					"subscription monitoring fetching past SignatureRequested events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := srs.contract.PastSignatureRequestedEvents(
					fromBlock,
					nil,
					srs.digestFilter,
				)
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				becdsakLogger.Infof(
					"subscription monitoring fetched [%v] past SignatureRequested events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := srs.contract.watchSignatureRequested(
		sink,
		srs.digestFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (becdsak *BondedECDSAKeep) watchSignatureRequested(
	sink chan *abi.BondedECDSAKeepSignatureRequested,
	digestFilter [][32]uint8,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return becdsak.contract.WatchSignatureRequested(
			&bind.WatchOpts{Context: ctx},
			sink,
			digestFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		becdsakLogger.Errorf(
			"subscription to event SignatureRequested had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		becdsakLogger.Errorf(
			"subscription to event SignatureRequested failed "+
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

func (becdsak *BondedECDSAKeep) SignatureSubmitted(
	opts *ethlike.SubscribeOpts,
	digestFilter [][32]uint8,
) *BecdsakSignatureSubmittedSubscription {
	if opts == nil {
		opts = new(ethlike.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BecdsakSignatureSubmittedSubscription{
		becdsak,
		opts,
		digestFilter,
	}
}

type BecdsakSignatureSubmittedSubscription struct {
	contract     *BondedECDSAKeep
	opts         *ethlike.SubscribeOpts
	digestFilter [][32]uint8
}

type bondedECDSAKeepSignatureSubmittedFunc func(
	Digest [32]uint8,
	R [32]uint8,
	S [32]uint8,
	RecoveryID uint8,
	blockNumber uint64,
)

func (sss *BecdsakSignatureSubmittedSubscription) OnEvent(
	handler bondedECDSAKeepSignatureSubmittedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BondedECDSAKeepSignatureSubmitted)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Digest,
					event.R,
					event.S,
					event.RecoveryID,
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := sss.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (sss *BecdsakSignatureSubmittedSubscription) Pipe(
	sink chan *abi.BondedECDSAKeepSignatureSubmitted,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(sss.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := sss.contract.blockCounter.CurrentBlock()
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - sss.opts.PastBlocks

				becdsakLogger.Infof(
					"subscription monitoring fetching past SignatureSubmitted events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := sss.contract.PastSignatureSubmittedEvents(
					fromBlock,
					nil,
					sss.digestFilter,
				)
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				becdsakLogger.Infof(
					"subscription monitoring fetched [%v] past SignatureSubmitted events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := sss.contract.watchSignatureSubmitted(
		sink,
		sss.digestFilter,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (becdsak *BondedECDSAKeep) watchSignatureSubmitted(
	sink chan *abi.BondedECDSAKeepSignatureSubmitted,
	digestFilter [][32]uint8,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return becdsak.contract.WatchSignatureSubmitted(
			&bind.WatchOpts{Context: ctx},
			sink,
			digestFilter,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		becdsakLogger.Errorf(
			"subscription to event SignatureSubmitted had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		becdsakLogger.Errorf(
			"subscription to event SignatureSubmitted failed "+
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

func (becdsak *BondedECDSAKeep) SlashingFailed(
	opts *ethlike.SubscribeOpts,
) *BecdsakSlashingFailedSubscription {
	if opts == nil {
		opts = new(ethlike.SubscribeOpts)
	}
	if opts.Tick == 0 {
		opts.Tick = chainutil.DefaultSubscribeOptsTick
	}
	if opts.PastBlocks == 0 {
		opts.PastBlocks = chainutil.DefaultSubscribeOptsPastBlocks
	}

	return &BecdsakSlashingFailedSubscription{
		becdsak,
		opts,
	}
}

type BecdsakSlashingFailedSubscription struct {
	contract *BondedECDSAKeep
	opts     *ethlike.SubscribeOpts
}

type bondedECDSAKeepSlashingFailedFunc func(
	blockNumber uint64,
)

func (sfs *BecdsakSlashingFailedSubscription) OnEvent(
	handler bondedECDSAKeepSlashingFailedFunc,
) subscription.EventSubscription {
	eventChan := make(chan *abi.BondedECDSAKeepSlashingFailed)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventChan:
				handler(
					event.Raw.BlockNumber,
				)
			}
		}
	}()

	sub := sfs.Pipe(eventChan)
	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (sfs *BecdsakSlashingFailedSubscription) Pipe(
	sink chan *abi.BondedECDSAKeepSlashingFailed,
) subscription.EventSubscription {
	ctx, cancelCtx := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(sfs.opts.Tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lastBlock, err := sfs.contract.blockCounter.CurrentBlock()
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
				}
				fromBlock := lastBlock - sfs.opts.PastBlocks

				becdsakLogger.Infof(
					"subscription monitoring fetching past SlashingFailed events "+
						"starting from block [%v]",
					fromBlock,
				)
				events, err := sfs.contract.PastSlashingFailedEvents(
					fromBlock,
					nil,
				)
				if err != nil {
					becdsakLogger.Errorf(
						"subscription failed to pull events: [%v]",
						err,
					)
					continue
				}
				becdsakLogger.Infof(
					"subscription monitoring fetched [%v] past SlashingFailed events",
					len(events),
				)

				for _, event := range events {
					sink <- event
				}
			}
		}
	}()

	sub := sfs.contract.watchSlashingFailed(
		sink,
	)

	return subscription.NewEventSubscription(func() {
		sub.Unsubscribe()
		cancelCtx()
	})
}

func (becdsak *BondedECDSAKeep) watchSlashingFailed(
	sink chan *abi.BondedECDSAKeepSlashingFailed,
) event.Subscription {
	subscribeFn := func(ctx context.Context) (event.Subscription, error) {
		return becdsak.contract.WatchSlashingFailed(
			&bind.WatchOpts{Context: ctx},
			sink,
		)
	}

	thresholdViolatedFn := func(elapsed time.Duration) {
		becdsakLogger.Errorf(
			"subscription to event SlashingFailed had to be "+
				"retried [%s] since the last attempt; please inspect "+
				"host chain connectivity",
			elapsed,
		)
	}

	subscriptionFailedFn := func(err error) {
		becdsakLogger.Errorf(
			"subscription to event SlashingFailed failed "+
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
