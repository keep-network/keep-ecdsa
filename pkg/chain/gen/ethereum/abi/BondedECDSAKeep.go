// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package abi

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// BondedECDSAKeepABI is the input ABI used to generate the binding from.
const BondedECDSAKeepABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"submittingMember\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"conflictingPublicKey\",\"type\":\"bytes\"}],\"name\":\"ConflictingPublicKeySubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ERC20RewardDistributed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ETHRewardDistributed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"KeepClosed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"KeepTerminated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"}],\"name\":\"PublicKeyPublished\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"digest\",\"type\":\"bytes32\"}],\"name\":\"SignatureRequested\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"digest\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"recoveryID\",\"type\":\"uint8\"}],\"name\":\"SignatureSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"SlashingFailed\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[],\"name\":\"checkBondAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"_r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_signedDigest\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"_preimage\",\"type\":\"bytes\"}],\"name\":\"checkSignatureFraud\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"_isFraud\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"closeKeep\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"digest\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"digests\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"distributeERC20Reward\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"distributeETHReward\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_member\",\"type\":\"address\"}],\"name\":\"getMemberETHBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getMembers\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOpenedTimestamp\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getPublicKey\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"honestThreshold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"_members\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"_honestThreshold\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_memberStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_stakeLockDuration\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_tokenStaking\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_keepBonding\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"_keepFactory\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"isActive\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_digest\",\"type\":\"bytes32\"}],\"name\":\"isAwaitingSignature\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"isClosed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"isTerminated\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"memberStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"members\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"publicKey\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"returnPartialSignerBonds\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"seizeSignerBonds\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_digest\",\"type\":\"bytes32\"}],\"name\":\"sign\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_publicKey\",\"type\":\"bytes\"}],\"name\":\"submitPublicKey\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"_recoveryID\",\"type\":\"uint8\"}],\"name\":\"submitSignature\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"_v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"_r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_signedDigest\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"_preimage\",\"type\":\"bytes\"}],\"name\":\"submitSignatureFraud\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"_isFraud\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_member\",\"type\":\"address\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// BondedECDSAKeep is an auto generated Go binding around an Ethereum contract.
type BondedECDSAKeep struct {
	BondedECDSAKeepCaller     // Read-only binding to the contract
	BondedECDSAKeepTransactor // Write-only binding to the contract
	BondedECDSAKeepFilterer   // Log filterer for contract events
}

// BondedECDSAKeepCaller is an auto generated read-only Go binding around an Ethereum contract.
type BondedECDSAKeepCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BondedECDSAKeepTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BondedECDSAKeepTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BondedECDSAKeepFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BondedECDSAKeepFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BondedECDSAKeepSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BondedECDSAKeepSession struct {
	Contract     *BondedECDSAKeep  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BondedECDSAKeepCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BondedECDSAKeepCallerSession struct {
	Contract *BondedECDSAKeepCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// BondedECDSAKeepTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BondedECDSAKeepTransactorSession struct {
	Contract     *BondedECDSAKeepTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// BondedECDSAKeepRaw is an auto generated low-level Go binding around an Ethereum contract.
type BondedECDSAKeepRaw struct {
	Contract *BondedECDSAKeep // Generic contract binding to access the raw methods on
}

// BondedECDSAKeepCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BondedECDSAKeepCallerRaw struct {
	Contract *BondedECDSAKeepCaller // Generic read-only contract binding to access the raw methods on
}

// BondedECDSAKeepTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BondedECDSAKeepTransactorRaw struct {
	Contract *BondedECDSAKeepTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBondedECDSAKeep creates a new instance of BondedECDSAKeep, bound to a specific deployed contract.
func NewBondedECDSAKeep(address common.Address, backend bind.ContractBackend) (*BondedECDSAKeep, error) {
	contract, err := bindBondedECDSAKeep(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BondedECDSAKeep{BondedECDSAKeepCaller: BondedECDSAKeepCaller{contract: contract}, BondedECDSAKeepTransactor: BondedECDSAKeepTransactor{contract: contract}, BondedECDSAKeepFilterer: BondedECDSAKeepFilterer{contract: contract}}, nil
}

// NewBondedECDSAKeepCaller creates a new read-only instance of BondedECDSAKeep, bound to a specific deployed contract.
func NewBondedECDSAKeepCaller(address common.Address, caller bind.ContractCaller) (*BondedECDSAKeepCaller, error) {
	contract, err := bindBondedECDSAKeep(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BondedECDSAKeepCaller{contract: contract}, nil
}

// NewBondedECDSAKeepTransactor creates a new write-only instance of BondedECDSAKeep, bound to a specific deployed contract.
func NewBondedECDSAKeepTransactor(address common.Address, transactor bind.ContractTransactor) (*BondedECDSAKeepTransactor, error) {
	contract, err := bindBondedECDSAKeep(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BondedECDSAKeepTransactor{contract: contract}, nil
}

// NewBondedECDSAKeepFilterer creates a new log filterer instance of BondedECDSAKeep, bound to a specific deployed contract.
func NewBondedECDSAKeepFilterer(address common.Address, filterer bind.ContractFilterer) (*BondedECDSAKeepFilterer, error) {
	contract, err := bindBondedECDSAKeep(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BondedECDSAKeepFilterer{contract: contract}, nil
}

// bindBondedECDSAKeep binds a generic wrapper to an already deployed contract.
func bindBondedECDSAKeep(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BondedECDSAKeepABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BondedECDSAKeep *BondedECDSAKeepRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BondedECDSAKeep.Contract.BondedECDSAKeepCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BondedECDSAKeep *BondedECDSAKeepRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.BondedECDSAKeepTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BondedECDSAKeep *BondedECDSAKeepRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.BondedECDSAKeepTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BondedECDSAKeep *BondedECDSAKeepCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BondedECDSAKeep.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BondedECDSAKeep *BondedECDSAKeepTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BondedECDSAKeep *BondedECDSAKeepTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.contract.Transact(opts, method, params...)
}

// CheckBondAmount is a free data retrieval call binding the contract method 0xdc3d6da8.
//
// Solidity: function checkBondAmount() view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepCaller) CheckBondAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "checkBondAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CheckBondAmount is a free data retrieval call binding the contract method 0xdc3d6da8.
//
// Solidity: function checkBondAmount() view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepSession) CheckBondAmount() (*big.Int, error) {
	return _BondedECDSAKeep.Contract.CheckBondAmount(&_BondedECDSAKeep.CallOpts)
}

// CheckBondAmount is a free data retrieval call binding the contract method 0xdc3d6da8.
//
// Solidity: function checkBondAmount() view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) CheckBondAmount() (*big.Int, error) {
	return _BondedECDSAKeep.Contract.CheckBondAmount(&_BondedECDSAKeep.CallOpts)
}

// CheckSignatureFraud is a free data retrieval call binding the contract method 0xbf9c8301.
//
// Solidity: function checkSignatureFraud(uint8 _v, bytes32 _r, bytes32 _s, bytes32 _signedDigest, bytes _preimage) view returns(bool _isFraud)
func (_BondedECDSAKeep *BondedECDSAKeepCaller) CheckSignatureFraud(opts *bind.CallOpts, _v uint8, _r [32]byte, _s [32]byte, _signedDigest [32]byte, _preimage []byte) (bool, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "checkSignatureFraud", _v, _r, _s, _signedDigest, _preimage)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CheckSignatureFraud is a free data retrieval call binding the contract method 0xbf9c8301.
//
// Solidity: function checkSignatureFraud(uint8 _v, bytes32 _r, bytes32 _s, bytes32 _signedDigest, bytes _preimage) view returns(bool _isFraud)
func (_BondedECDSAKeep *BondedECDSAKeepSession) CheckSignatureFraud(_v uint8, _r [32]byte, _s [32]byte, _signedDigest [32]byte, _preimage []byte) (bool, error) {
	return _BondedECDSAKeep.Contract.CheckSignatureFraud(&_BondedECDSAKeep.CallOpts, _v, _r, _s, _signedDigest, _preimage)
}

// CheckSignatureFraud is a free data retrieval call binding the contract method 0xbf9c8301.
//
// Solidity: function checkSignatureFraud(uint8 _v, bytes32 _r, bytes32 _s, bytes32 _signedDigest, bytes _preimage) view returns(bool _isFraud)
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) CheckSignatureFraud(_v uint8, _r [32]byte, _s [32]byte, _signedDigest [32]byte, _preimage []byte) (bool, error) {
	return _BondedECDSAKeep.Contract.CheckSignatureFraud(&_BondedECDSAKeep.CallOpts, _v, _r, _s, _signedDigest, _preimage)
}

// Digest is a free data retrieval call binding the contract method 0x52a82b65.
//
// Solidity: function digest() view returns(bytes32)
func (_BondedECDSAKeep *BondedECDSAKeepCaller) Digest(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "digest")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// Digest is a free data retrieval call binding the contract method 0x52a82b65.
//
// Solidity: function digest() view returns(bytes32)
func (_BondedECDSAKeep *BondedECDSAKeepSession) Digest() ([32]byte, error) {
	return _BondedECDSAKeep.Contract.Digest(&_BondedECDSAKeep.CallOpts)
}

// Digest is a free data retrieval call binding the contract method 0x52a82b65.
//
// Solidity: function digest() view returns(bytes32)
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) Digest() ([32]byte, error) {
	return _BondedECDSAKeep.Contract.Digest(&_BondedECDSAKeep.CallOpts)
}

// Digests is a free data retrieval call binding the contract method 0x01ac4293.
//
// Solidity: function digests(bytes32 ) view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepCaller) Digests(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "digests", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Digests is a free data retrieval call binding the contract method 0x01ac4293.
//
// Solidity: function digests(bytes32 ) view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepSession) Digests(arg0 [32]byte) (*big.Int, error) {
	return _BondedECDSAKeep.Contract.Digests(&_BondedECDSAKeep.CallOpts, arg0)
}

// Digests is a free data retrieval call binding the contract method 0x01ac4293.
//
// Solidity: function digests(bytes32 ) view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) Digests(arg0 [32]byte) (*big.Int, error) {
	return _BondedECDSAKeep.Contract.Digests(&_BondedECDSAKeep.CallOpts, arg0)
}

// GetMemberETHBalance is a free data retrieval call binding the contract method 0xd5cc8b0f.
//
// Solidity: function getMemberETHBalance(address _member) view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepCaller) GetMemberETHBalance(opts *bind.CallOpts, _member common.Address) (*big.Int, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "getMemberETHBalance", _member)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMemberETHBalance is a free data retrieval call binding the contract method 0xd5cc8b0f.
//
// Solidity: function getMemberETHBalance(address _member) view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepSession) GetMemberETHBalance(_member common.Address) (*big.Int, error) {
	return _BondedECDSAKeep.Contract.GetMemberETHBalance(&_BondedECDSAKeep.CallOpts, _member)
}

// GetMemberETHBalance is a free data retrieval call binding the contract method 0xd5cc8b0f.
//
// Solidity: function getMemberETHBalance(address _member) view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) GetMemberETHBalance(_member common.Address) (*big.Int, error) {
	return _BondedECDSAKeep.Contract.GetMemberETHBalance(&_BondedECDSAKeep.CallOpts, _member)
}

// GetMembers is a free data retrieval call binding the contract method 0x9eab5253.
//
// Solidity: function getMembers() view returns(address[])
func (_BondedECDSAKeep *BondedECDSAKeepCaller) GetMembers(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "getMembers")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetMembers is a free data retrieval call binding the contract method 0x9eab5253.
//
// Solidity: function getMembers() view returns(address[])
func (_BondedECDSAKeep *BondedECDSAKeepSession) GetMembers() ([]common.Address, error) {
	return _BondedECDSAKeep.Contract.GetMembers(&_BondedECDSAKeep.CallOpts)
}

// GetMembers is a free data retrieval call binding the contract method 0x9eab5253.
//
// Solidity: function getMembers() view returns(address[])
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) GetMembers() ([]common.Address, error) {
	return _BondedECDSAKeep.Contract.GetMembers(&_BondedECDSAKeep.CallOpts)
}

// GetOpenedTimestamp is a free data retrieval call binding the contract method 0xf4c2b4c1.
//
// Solidity: function getOpenedTimestamp() view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepCaller) GetOpenedTimestamp(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "getOpenedTimestamp")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetOpenedTimestamp is a free data retrieval call binding the contract method 0xf4c2b4c1.
//
// Solidity: function getOpenedTimestamp() view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepSession) GetOpenedTimestamp() (*big.Int, error) {
	return _BondedECDSAKeep.Contract.GetOpenedTimestamp(&_BondedECDSAKeep.CallOpts)
}

// GetOpenedTimestamp is a free data retrieval call binding the contract method 0xf4c2b4c1.
//
// Solidity: function getOpenedTimestamp() view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) GetOpenedTimestamp() (*big.Int, error) {
	return _BondedECDSAKeep.Contract.GetOpenedTimestamp(&_BondedECDSAKeep.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_BondedECDSAKeep *BondedECDSAKeepCaller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "getOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_BondedECDSAKeep *BondedECDSAKeepSession) GetOwner() (common.Address, error) {
	return _BondedECDSAKeep.Contract.GetOwner(&_BondedECDSAKeep.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) GetOwner() (common.Address, error) {
	return _BondedECDSAKeep.Contract.GetOwner(&_BondedECDSAKeep.CallOpts)
}

// GetPublicKey is a free data retrieval call binding the contract method 0x2e334452.
//
// Solidity: function getPublicKey() view returns(bytes)
func (_BondedECDSAKeep *BondedECDSAKeepCaller) GetPublicKey(opts *bind.CallOpts) ([]byte, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "getPublicKey")

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetPublicKey is a free data retrieval call binding the contract method 0x2e334452.
//
// Solidity: function getPublicKey() view returns(bytes)
func (_BondedECDSAKeep *BondedECDSAKeepSession) GetPublicKey() ([]byte, error) {
	return _BondedECDSAKeep.Contract.GetPublicKey(&_BondedECDSAKeep.CallOpts)
}

// GetPublicKey is a free data retrieval call binding the contract method 0x2e334452.
//
// Solidity: function getPublicKey() view returns(bytes)
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) GetPublicKey() ([]byte, error) {
	return _BondedECDSAKeep.Contract.GetPublicKey(&_BondedECDSAKeep.CallOpts)
}

// HonestThreshold is a free data retrieval call binding the contract method 0x6806db1f.
//
// Solidity: function honestThreshold() view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepCaller) HonestThreshold(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "honestThreshold")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// HonestThreshold is a free data retrieval call binding the contract method 0x6806db1f.
//
// Solidity: function honestThreshold() view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepSession) HonestThreshold() (*big.Int, error) {
	return _BondedECDSAKeep.Contract.HonestThreshold(&_BondedECDSAKeep.CallOpts)
}

// HonestThreshold is a free data retrieval call binding the contract method 0x6806db1f.
//
// Solidity: function honestThreshold() view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) HonestThreshold() (*big.Int, error) {
	return _BondedECDSAKeep.Contract.HonestThreshold(&_BondedECDSAKeep.CallOpts)
}

// IsActive is a free data retrieval call binding the contract method 0x22f3e2d4.
//
// Solidity: function isActive() view returns(bool)
func (_BondedECDSAKeep *BondedECDSAKeepCaller) IsActive(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "isActive")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsActive is a free data retrieval call binding the contract method 0x22f3e2d4.
//
// Solidity: function isActive() view returns(bool)
func (_BondedECDSAKeep *BondedECDSAKeepSession) IsActive() (bool, error) {
	return _BondedECDSAKeep.Contract.IsActive(&_BondedECDSAKeep.CallOpts)
}

// IsActive is a free data retrieval call binding the contract method 0x22f3e2d4.
//
// Solidity: function isActive() view returns(bool)
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) IsActive() (bool, error) {
	return _BondedECDSAKeep.Contract.IsActive(&_BondedECDSAKeep.CallOpts)
}

// IsAwaitingSignature is a free data retrieval call binding the contract method 0xcb7cf187.
//
// Solidity: function isAwaitingSignature(bytes32 _digest) view returns(bool)
func (_BondedECDSAKeep *BondedECDSAKeepCaller) IsAwaitingSignature(opts *bind.CallOpts, _digest [32]byte) (bool, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "isAwaitingSignature", _digest)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAwaitingSignature is a free data retrieval call binding the contract method 0xcb7cf187.
//
// Solidity: function isAwaitingSignature(bytes32 _digest) view returns(bool)
func (_BondedECDSAKeep *BondedECDSAKeepSession) IsAwaitingSignature(_digest [32]byte) (bool, error) {
	return _BondedECDSAKeep.Contract.IsAwaitingSignature(&_BondedECDSAKeep.CallOpts, _digest)
}

// IsAwaitingSignature is a free data retrieval call binding the contract method 0xcb7cf187.
//
// Solidity: function isAwaitingSignature(bytes32 _digest) view returns(bool)
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) IsAwaitingSignature(_digest [32]byte) (bool, error) {
	return _BondedECDSAKeep.Contract.IsAwaitingSignature(&_BondedECDSAKeep.CallOpts, _digest)
}

// IsClosed is a free data retrieval call binding the contract method 0xc2b6b58c.
//
// Solidity: function isClosed() view returns(bool)
func (_BondedECDSAKeep *BondedECDSAKeepCaller) IsClosed(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "isClosed")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsClosed is a free data retrieval call binding the contract method 0xc2b6b58c.
//
// Solidity: function isClosed() view returns(bool)
func (_BondedECDSAKeep *BondedECDSAKeepSession) IsClosed() (bool, error) {
	return _BondedECDSAKeep.Contract.IsClosed(&_BondedECDSAKeep.CallOpts)
}

// IsClosed is a free data retrieval call binding the contract method 0xc2b6b58c.
//
// Solidity: function isClosed() view returns(bool)
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) IsClosed() (bool, error) {
	return _BondedECDSAKeep.Contract.IsClosed(&_BondedECDSAKeep.CallOpts)
}

// IsTerminated is a free data retrieval call binding the contract method 0xd1cc9976.
//
// Solidity: function isTerminated() view returns(bool)
func (_BondedECDSAKeep *BondedECDSAKeepCaller) IsTerminated(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "isTerminated")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsTerminated is a free data retrieval call binding the contract method 0xd1cc9976.
//
// Solidity: function isTerminated() view returns(bool)
func (_BondedECDSAKeep *BondedECDSAKeepSession) IsTerminated() (bool, error) {
	return _BondedECDSAKeep.Contract.IsTerminated(&_BondedECDSAKeep.CallOpts)
}

// IsTerminated is a free data retrieval call binding the contract method 0xd1cc9976.
//
// Solidity: function isTerminated() view returns(bool)
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) IsTerminated() (bool, error) {
	return _BondedECDSAKeep.Contract.IsTerminated(&_BondedECDSAKeep.CallOpts)
}

// MemberStake is a free data retrieval call binding the contract method 0xc9de240d.
//
// Solidity: function memberStake() view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepCaller) MemberStake(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "memberStake")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MemberStake is a free data retrieval call binding the contract method 0xc9de240d.
//
// Solidity: function memberStake() view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepSession) MemberStake() (*big.Int, error) {
	return _BondedECDSAKeep.Contract.MemberStake(&_BondedECDSAKeep.CallOpts)
}

// MemberStake is a free data retrieval call binding the contract method 0xc9de240d.
//
// Solidity: function memberStake() view returns(uint256)
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) MemberStake() (*big.Int, error) {
	return _BondedECDSAKeep.Contract.MemberStake(&_BondedECDSAKeep.CallOpts)
}

// Members is a free data retrieval call binding the contract method 0x5daf08ca.
//
// Solidity: function members(uint256 ) view returns(address)
func (_BondedECDSAKeep *BondedECDSAKeepCaller) Members(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "members", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Members is a free data retrieval call binding the contract method 0x5daf08ca.
//
// Solidity: function members(uint256 ) view returns(address)
func (_BondedECDSAKeep *BondedECDSAKeepSession) Members(arg0 *big.Int) (common.Address, error) {
	return _BondedECDSAKeep.Contract.Members(&_BondedECDSAKeep.CallOpts, arg0)
}

// Members is a free data retrieval call binding the contract method 0x5daf08ca.
//
// Solidity: function members(uint256 ) view returns(address)
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) Members(arg0 *big.Int) (common.Address, error) {
	return _BondedECDSAKeep.Contract.Members(&_BondedECDSAKeep.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BondedECDSAKeep *BondedECDSAKeepCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BondedECDSAKeep *BondedECDSAKeepSession) Owner() (common.Address, error) {
	return _BondedECDSAKeep.Contract.Owner(&_BondedECDSAKeep.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) Owner() (common.Address, error) {
	return _BondedECDSAKeep.Contract.Owner(&_BondedECDSAKeep.CallOpts)
}

// PublicKey is a free data retrieval call binding the contract method 0x63ffab31.
//
// Solidity: function publicKey() view returns(bytes)
func (_BondedECDSAKeep *BondedECDSAKeepCaller) PublicKey(opts *bind.CallOpts) ([]byte, error) {
	var out []interface{}
	err := _BondedECDSAKeep.contract.Call(opts, &out, "publicKey")

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// PublicKey is a free data retrieval call binding the contract method 0x63ffab31.
//
// Solidity: function publicKey() view returns(bytes)
func (_BondedECDSAKeep *BondedECDSAKeepSession) PublicKey() ([]byte, error) {
	return _BondedECDSAKeep.Contract.PublicKey(&_BondedECDSAKeep.CallOpts)
}

// PublicKey is a free data retrieval call binding the contract method 0x63ffab31.
//
// Solidity: function publicKey() view returns(bytes)
func (_BondedECDSAKeep *BondedECDSAKeepCallerSession) PublicKey() ([]byte, error) {
	return _BondedECDSAKeep.Contract.PublicKey(&_BondedECDSAKeep.CallOpts)
}

// CloseKeep is a paid mutator transaction binding the contract method 0xa15c3bbb.
//
// Solidity: function closeKeep() returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactor) CloseKeep(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BondedECDSAKeep.contract.Transact(opts, "closeKeep")
}

// CloseKeep is a paid mutator transaction binding the contract method 0xa15c3bbb.
//
// Solidity: function closeKeep() returns()
func (_BondedECDSAKeep *BondedECDSAKeepSession) CloseKeep() (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.CloseKeep(&_BondedECDSAKeep.TransactOpts)
}

// CloseKeep is a paid mutator transaction binding the contract method 0xa15c3bbb.
//
// Solidity: function closeKeep() returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactorSession) CloseKeep() (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.CloseKeep(&_BondedECDSAKeep.TransactOpts)
}

// DistributeERC20Reward is a paid mutator transaction binding the contract method 0x5a89f810.
//
// Solidity: function distributeERC20Reward(address _tokenAddress, uint256 _value) returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactor) DistributeERC20Reward(opts *bind.TransactOpts, _tokenAddress common.Address, _value *big.Int) (*types.Transaction, error) {
	return _BondedECDSAKeep.contract.Transact(opts, "distributeERC20Reward", _tokenAddress, _value)
}

// DistributeERC20Reward is a paid mutator transaction binding the contract method 0x5a89f810.
//
// Solidity: function distributeERC20Reward(address _tokenAddress, uint256 _value) returns()
func (_BondedECDSAKeep *BondedECDSAKeepSession) DistributeERC20Reward(_tokenAddress common.Address, _value *big.Int) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.DistributeERC20Reward(&_BondedECDSAKeep.TransactOpts, _tokenAddress, _value)
}

// DistributeERC20Reward is a paid mutator transaction binding the contract method 0x5a89f810.
//
// Solidity: function distributeERC20Reward(address _tokenAddress, uint256 _value) returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactorSession) DistributeERC20Reward(_tokenAddress common.Address, _value *big.Int) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.DistributeERC20Reward(&_BondedECDSAKeep.TransactOpts, _tokenAddress, _value)
}

// DistributeETHReward is a paid mutator transaction binding the contract method 0x2930e170.
//
// Solidity: function distributeETHReward() payable returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactor) DistributeETHReward(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BondedECDSAKeep.contract.Transact(opts, "distributeETHReward")
}

// DistributeETHReward is a paid mutator transaction binding the contract method 0x2930e170.
//
// Solidity: function distributeETHReward() payable returns()
func (_BondedECDSAKeep *BondedECDSAKeepSession) DistributeETHReward() (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.DistributeETHReward(&_BondedECDSAKeep.TransactOpts)
}

// DistributeETHReward is a paid mutator transaction binding the contract method 0x2930e170.
//
// Solidity: function distributeETHReward() payable returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactorSession) DistributeETHReward() (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.DistributeETHReward(&_BondedECDSAKeep.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x79a80491.
//
// Solidity: function initialize(address _owner, address[] _members, uint256 _honestThreshold, uint256 _memberStake, uint256 _stakeLockDuration, address _tokenStaking, address _keepBonding, address _keepFactory) returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactor) Initialize(opts *bind.TransactOpts, _owner common.Address, _members []common.Address, _honestThreshold *big.Int, _memberStake *big.Int, _stakeLockDuration *big.Int, _tokenStaking common.Address, _keepBonding common.Address, _keepFactory common.Address) (*types.Transaction, error) {
	return _BondedECDSAKeep.contract.Transact(opts, "initialize", _owner, _members, _honestThreshold, _memberStake, _stakeLockDuration, _tokenStaking, _keepBonding, _keepFactory)
}

// Initialize is a paid mutator transaction binding the contract method 0x79a80491.
//
// Solidity: function initialize(address _owner, address[] _members, uint256 _honestThreshold, uint256 _memberStake, uint256 _stakeLockDuration, address _tokenStaking, address _keepBonding, address _keepFactory) returns()
func (_BondedECDSAKeep *BondedECDSAKeepSession) Initialize(_owner common.Address, _members []common.Address, _honestThreshold *big.Int, _memberStake *big.Int, _stakeLockDuration *big.Int, _tokenStaking common.Address, _keepBonding common.Address, _keepFactory common.Address) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.Initialize(&_BondedECDSAKeep.TransactOpts, _owner, _members, _honestThreshold, _memberStake, _stakeLockDuration, _tokenStaking, _keepBonding, _keepFactory)
}

// Initialize is a paid mutator transaction binding the contract method 0x79a80491.
//
// Solidity: function initialize(address _owner, address[] _members, uint256 _honestThreshold, uint256 _memberStake, uint256 _stakeLockDuration, address _tokenStaking, address _keepBonding, address _keepFactory) returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactorSession) Initialize(_owner common.Address, _members []common.Address, _honestThreshold *big.Int, _memberStake *big.Int, _stakeLockDuration *big.Int, _tokenStaking common.Address, _keepBonding common.Address, _keepFactory common.Address) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.Initialize(&_BondedECDSAKeep.TransactOpts, _owner, _members, _honestThreshold, _memberStake, _stakeLockDuration, _tokenStaking, _keepBonding, _keepFactory)
}

// ReturnPartialSignerBonds is a paid mutator transaction binding the contract method 0x6ed15f94.
//
// Solidity: function returnPartialSignerBonds() payable returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactor) ReturnPartialSignerBonds(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BondedECDSAKeep.contract.Transact(opts, "returnPartialSignerBonds")
}

// ReturnPartialSignerBonds is a paid mutator transaction binding the contract method 0x6ed15f94.
//
// Solidity: function returnPartialSignerBonds() payable returns()
func (_BondedECDSAKeep *BondedECDSAKeepSession) ReturnPartialSignerBonds() (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.ReturnPartialSignerBonds(&_BondedECDSAKeep.TransactOpts)
}

// ReturnPartialSignerBonds is a paid mutator transaction binding the contract method 0x6ed15f94.
//
// Solidity: function returnPartialSignerBonds() payable returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactorSession) ReturnPartialSignerBonds() (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.ReturnPartialSignerBonds(&_BondedECDSAKeep.TransactOpts)
}

// SeizeSignerBonds is a paid mutator transaction binding the contract method 0x07acd5cb.
//
// Solidity: function seizeSignerBonds() returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactor) SeizeSignerBonds(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BondedECDSAKeep.contract.Transact(opts, "seizeSignerBonds")
}

// SeizeSignerBonds is a paid mutator transaction binding the contract method 0x07acd5cb.
//
// Solidity: function seizeSignerBonds() returns()
func (_BondedECDSAKeep *BondedECDSAKeepSession) SeizeSignerBonds() (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.SeizeSignerBonds(&_BondedECDSAKeep.TransactOpts)
}

// SeizeSignerBonds is a paid mutator transaction binding the contract method 0x07acd5cb.
//
// Solidity: function seizeSignerBonds() returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactorSession) SeizeSignerBonds() (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.SeizeSignerBonds(&_BondedECDSAKeep.TransactOpts)
}

// Sign is a paid mutator transaction binding the contract method 0x799cd333.
//
// Solidity: function sign(bytes32 _digest) returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactor) Sign(opts *bind.TransactOpts, _digest [32]byte) (*types.Transaction, error) {
	return _BondedECDSAKeep.contract.Transact(opts, "sign", _digest)
}

// Sign is a paid mutator transaction binding the contract method 0x799cd333.
//
// Solidity: function sign(bytes32 _digest) returns()
func (_BondedECDSAKeep *BondedECDSAKeepSession) Sign(_digest [32]byte) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.Sign(&_BondedECDSAKeep.TransactOpts, _digest)
}

// Sign is a paid mutator transaction binding the contract method 0x799cd333.
//
// Solidity: function sign(bytes32 _digest) returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactorSession) Sign(_digest [32]byte) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.Sign(&_BondedECDSAKeep.TransactOpts, _digest)
}

// SubmitPublicKey is a paid mutator transaction binding the contract method 0xabd14f37.
//
// Solidity: function submitPublicKey(bytes _publicKey) returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactor) SubmitPublicKey(opts *bind.TransactOpts, _publicKey []byte) (*types.Transaction, error) {
	return _BondedECDSAKeep.contract.Transact(opts, "submitPublicKey", _publicKey)
}

// SubmitPublicKey is a paid mutator transaction binding the contract method 0xabd14f37.
//
// Solidity: function submitPublicKey(bytes _publicKey) returns()
func (_BondedECDSAKeep *BondedECDSAKeepSession) SubmitPublicKey(_publicKey []byte) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.SubmitPublicKey(&_BondedECDSAKeep.TransactOpts, _publicKey)
}

// SubmitPublicKey is a paid mutator transaction binding the contract method 0xabd14f37.
//
// Solidity: function submitPublicKey(bytes _publicKey) returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactorSession) SubmitPublicKey(_publicKey []byte) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.SubmitPublicKey(&_BondedECDSAKeep.TransactOpts, _publicKey)
}

// SubmitSignature is a paid mutator transaction binding the contract method 0x7df2b357.
//
// Solidity: function submitSignature(bytes32 _r, bytes32 _s, uint8 _recoveryID) returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactor) SubmitSignature(opts *bind.TransactOpts, _r [32]byte, _s [32]byte, _recoveryID uint8) (*types.Transaction, error) {
	return _BondedECDSAKeep.contract.Transact(opts, "submitSignature", _r, _s, _recoveryID)
}

// SubmitSignature is a paid mutator transaction binding the contract method 0x7df2b357.
//
// Solidity: function submitSignature(bytes32 _r, bytes32 _s, uint8 _recoveryID) returns()
func (_BondedECDSAKeep *BondedECDSAKeepSession) SubmitSignature(_r [32]byte, _s [32]byte, _recoveryID uint8) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.SubmitSignature(&_BondedECDSAKeep.TransactOpts, _r, _s, _recoveryID)
}

// SubmitSignature is a paid mutator transaction binding the contract method 0x7df2b357.
//
// Solidity: function submitSignature(bytes32 _r, bytes32 _s, uint8 _recoveryID) returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactorSession) SubmitSignature(_r [32]byte, _s [32]byte, _recoveryID uint8) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.SubmitSignature(&_BondedECDSAKeep.TransactOpts, _r, _s, _recoveryID)
}

// SubmitSignatureFraud is a paid mutator transaction binding the contract method 0xf15d1a90.
//
// Solidity: function submitSignatureFraud(uint8 _v, bytes32 _r, bytes32 _s, bytes32 _signedDigest, bytes _preimage) returns(bool _isFraud)
func (_BondedECDSAKeep *BondedECDSAKeepTransactor) SubmitSignatureFraud(opts *bind.TransactOpts, _v uint8, _r [32]byte, _s [32]byte, _signedDigest [32]byte, _preimage []byte) (*types.Transaction, error) {
	return _BondedECDSAKeep.contract.Transact(opts, "submitSignatureFraud", _v, _r, _s, _signedDigest, _preimage)
}

// SubmitSignatureFraud is a paid mutator transaction binding the contract method 0xf15d1a90.
//
// Solidity: function submitSignatureFraud(uint8 _v, bytes32 _r, bytes32 _s, bytes32 _signedDigest, bytes _preimage) returns(bool _isFraud)
func (_BondedECDSAKeep *BondedECDSAKeepSession) SubmitSignatureFraud(_v uint8, _r [32]byte, _s [32]byte, _signedDigest [32]byte, _preimage []byte) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.SubmitSignatureFraud(&_BondedECDSAKeep.TransactOpts, _v, _r, _s, _signedDigest, _preimage)
}

// SubmitSignatureFraud is a paid mutator transaction binding the contract method 0xf15d1a90.
//
// Solidity: function submitSignatureFraud(uint8 _v, bytes32 _r, bytes32 _s, bytes32 _signedDigest, bytes _preimage) returns(bool _isFraud)
func (_BondedECDSAKeep *BondedECDSAKeepTransactorSession) SubmitSignatureFraud(_v uint8, _r [32]byte, _s [32]byte, _signedDigest [32]byte, _preimage []byte) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.SubmitSignatureFraud(&_BondedECDSAKeep.TransactOpts, _v, _r, _s, _signedDigest, _preimage)
}

// Withdraw is a paid mutator transaction binding the contract method 0x51cff8d9.
//
// Solidity: function withdraw(address _member) returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactor) Withdraw(opts *bind.TransactOpts, _member common.Address) (*types.Transaction, error) {
	return _BondedECDSAKeep.contract.Transact(opts, "withdraw", _member)
}

// Withdraw is a paid mutator transaction binding the contract method 0x51cff8d9.
//
// Solidity: function withdraw(address _member) returns()
func (_BondedECDSAKeep *BondedECDSAKeepSession) Withdraw(_member common.Address) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.Withdraw(&_BondedECDSAKeep.TransactOpts, _member)
}

// Withdraw is a paid mutator transaction binding the contract method 0x51cff8d9.
//
// Solidity: function withdraw(address _member) returns()
func (_BondedECDSAKeep *BondedECDSAKeepTransactorSession) Withdraw(_member common.Address) (*types.Transaction, error) {
	return _BondedECDSAKeep.Contract.Withdraw(&_BondedECDSAKeep.TransactOpts, _member)
}

// BondedECDSAKeepConflictingPublicKeySubmittedIterator is returned from FilterConflictingPublicKeySubmitted and is used to iterate over the raw logs and unpacked data for ConflictingPublicKeySubmitted events raised by the BondedECDSAKeep contract.
type BondedECDSAKeepConflictingPublicKeySubmittedIterator struct {
	Event *BondedECDSAKeepConflictingPublicKeySubmitted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BondedECDSAKeepConflictingPublicKeySubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BondedECDSAKeepConflictingPublicKeySubmitted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BondedECDSAKeepConflictingPublicKeySubmitted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BondedECDSAKeepConflictingPublicKeySubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BondedECDSAKeepConflictingPublicKeySubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BondedECDSAKeepConflictingPublicKeySubmitted represents a ConflictingPublicKeySubmitted event raised by the BondedECDSAKeep contract.
type BondedECDSAKeepConflictingPublicKeySubmitted struct {
	SubmittingMember     common.Address
	ConflictingPublicKey []byte
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterConflictingPublicKeySubmitted is a free log retrieval operation binding the contract event 0x99d98e35ad6445ac964c46a75c7f748e8f390ebdca5a924cd8f92d674fa34ff7.
//
// Solidity: event ConflictingPublicKeySubmitted(address indexed submittingMember, bytes conflictingPublicKey)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) FilterConflictingPublicKeySubmitted(opts *bind.FilterOpts, submittingMember []common.Address) (*BondedECDSAKeepConflictingPublicKeySubmittedIterator, error) {

	var submittingMemberRule []interface{}
	for _, submittingMemberItem := range submittingMember {
		submittingMemberRule = append(submittingMemberRule, submittingMemberItem)
	}

	logs, sub, err := _BondedECDSAKeep.contract.FilterLogs(opts, "ConflictingPublicKeySubmitted", submittingMemberRule)
	if err != nil {
		return nil, err
	}
	return &BondedECDSAKeepConflictingPublicKeySubmittedIterator{contract: _BondedECDSAKeep.contract, event: "ConflictingPublicKeySubmitted", logs: logs, sub: sub}, nil
}

// WatchConflictingPublicKeySubmitted is a free log subscription operation binding the contract event 0x99d98e35ad6445ac964c46a75c7f748e8f390ebdca5a924cd8f92d674fa34ff7.
//
// Solidity: event ConflictingPublicKeySubmitted(address indexed submittingMember, bytes conflictingPublicKey)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) WatchConflictingPublicKeySubmitted(opts *bind.WatchOpts, sink chan<- *BondedECDSAKeepConflictingPublicKeySubmitted, submittingMember []common.Address) (event.Subscription, error) {

	var submittingMemberRule []interface{}
	for _, submittingMemberItem := range submittingMember {
		submittingMemberRule = append(submittingMemberRule, submittingMemberItem)
	}

	logs, sub, err := _BondedECDSAKeep.contract.WatchLogs(opts, "ConflictingPublicKeySubmitted", submittingMemberRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BondedECDSAKeepConflictingPublicKeySubmitted)
				if err := _BondedECDSAKeep.contract.UnpackLog(event, "ConflictingPublicKeySubmitted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseConflictingPublicKeySubmitted is a log parse operation binding the contract event 0x99d98e35ad6445ac964c46a75c7f748e8f390ebdca5a924cd8f92d674fa34ff7.
//
// Solidity: event ConflictingPublicKeySubmitted(address indexed submittingMember, bytes conflictingPublicKey)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) ParseConflictingPublicKeySubmitted(log types.Log) (*BondedECDSAKeepConflictingPublicKeySubmitted, error) {
	event := new(BondedECDSAKeepConflictingPublicKeySubmitted)
	if err := _BondedECDSAKeep.contract.UnpackLog(event, "ConflictingPublicKeySubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BondedECDSAKeepERC20RewardDistributedIterator is returned from FilterERC20RewardDistributed and is used to iterate over the raw logs and unpacked data for ERC20RewardDistributed events raised by the BondedECDSAKeep contract.
type BondedECDSAKeepERC20RewardDistributedIterator struct {
	Event *BondedECDSAKeepERC20RewardDistributed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BondedECDSAKeepERC20RewardDistributedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BondedECDSAKeepERC20RewardDistributed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BondedECDSAKeepERC20RewardDistributed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BondedECDSAKeepERC20RewardDistributedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BondedECDSAKeepERC20RewardDistributedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BondedECDSAKeepERC20RewardDistributed represents a ERC20RewardDistributed event raised by the BondedECDSAKeep contract.
type BondedECDSAKeepERC20RewardDistributed struct {
	Token  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterERC20RewardDistributed is a free log retrieval operation binding the contract event 0xb69f5873bb2e9e1cc495d5c23d2995010c3b5cdd1756e3cada2bc3f2150902cc.
//
// Solidity: event ERC20RewardDistributed(address indexed token, uint256 amount)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) FilterERC20RewardDistributed(opts *bind.FilterOpts, token []common.Address) (*BondedECDSAKeepERC20RewardDistributedIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _BondedECDSAKeep.contract.FilterLogs(opts, "ERC20RewardDistributed", tokenRule)
	if err != nil {
		return nil, err
	}
	return &BondedECDSAKeepERC20RewardDistributedIterator{contract: _BondedECDSAKeep.contract, event: "ERC20RewardDistributed", logs: logs, sub: sub}, nil
}

// WatchERC20RewardDistributed is a free log subscription operation binding the contract event 0xb69f5873bb2e9e1cc495d5c23d2995010c3b5cdd1756e3cada2bc3f2150902cc.
//
// Solidity: event ERC20RewardDistributed(address indexed token, uint256 amount)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) WatchERC20RewardDistributed(opts *bind.WatchOpts, sink chan<- *BondedECDSAKeepERC20RewardDistributed, token []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _BondedECDSAKeep.contract.WatchLogs(opts, "ERC20RewardDistributed", tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BondedECDSAKeepERC20RewardDistributed)
				if err := _BondedECDSAKeep.contract.UnpackLog(event, "ERC20RewardDistributed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseERC20RewardDistributed is a log parse operation binding the contract event 0xb69f5873bb2e9e1cc495d5c23d2995010c3b5cdd1756e3cada2bc3f2150902cc.
//
// Solidity: event ERC20RewardDistributed(address indexed token, uint256 amount)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) ParseERC20RewardDistributed(log types.Log) (*BondedECDSAKeepERC20RewardDistributed, error) {
	event := new(BondedECDSAKeepERC20RewardDistributed)
	if err := _BondedECDSAKeep.contract.UnpackLog(event, "ERC20RewardDistributed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BondedECDSAKeepETHRewardDistributedIterator is returned from FilterETHRewardDistributed and is used to iterate over the raw logs and unpacked data for ETHRewardDistributed events raised by the BondedECDSAKeep contract.
type BondedECDSAKeepETHRewardDistributedIterator struct {
	Event *BondedECDSAKeepETHRewardDistributed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BondedECDSAKeepETHRewardDistributedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BondedECDSAKeepETHRewardDistributed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BondedECDSAKeepETHRewardDistributed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BondedECDSAKeepETHRewardDistributedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BondedECDSAKeepETHRewardDistributedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BondedECDSAKeepETHRewardDistributed represents a ETHRewardDistributed event raised by the BondedECDSAKeep contract.
type BondedECDSAKeepETHRewardDistributed struct {
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterETHRewardDistributed is a free log retrieval operation binding the contract event 0xa9e4160b29b5c7db7fa61c512c4b45e7c3451c3331537f065a3417778cea5096.
//
// Solidity: event ETHRewardDistributed(uint256 amount)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) FilterETHRewardDistributed(opts *bind.FilterOpts) (*BondedECDSAKeepETHRewardDistributedIterator, error) {

	logs, sub, err := _BondedECDSAKeep.contract.FilterLogs(opts, "ETHRewardDistributed")
	if err != nil {
		return nil, err
	}
	return &BondedECDSAKeepETHRewardDistributedIterator{contract: _BondedECDSAKeep.contract, event: "ETHRewardDistributed", logs: logs, sub: sub}, nil
}

// WatchETHRewardDistributed is a free log subscription operation binding the contract event 0xa9e4160b29b5c7db7fa61c512c4b45e7c3451c3331537f065a3417778cea5096.
//
// Solidity: event ETHRewardDistributed(uint256 amount)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) WatchETHRewardDistributed(opts *bind.WatchOpts, sink chan<- *BondedECDSAKeepETHRewardDistributed) (event.Subscription, error) {

	logs, sub, err := _BondedECDSAKeep.contract.WatchLogs(opts, "ETHRewardDistributed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BondedECDSAKeepETHRewardDistributed)
				if err := _BondedECDSAKeep.contract.UnpackLog(event, "ETHRewardDistributed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseETHRewardDistributed is a log parse operation binding the contract event 0xa9e4160b29b5c7db7fa61c512c4b45e7c3451c3331537f065a3417778cea5096.
//
// Solidity: event ETHRewardDistributed(uint256 amount)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) ParseETHRewardDistributed(log types.Log) (*BondedECDSAKeepETHRewardDistributed, error) {
	event := new(BondedECDSAKeepETHRewardDistributed)
	if err := _BondedECDSAKeep.contract.UnpackLog(event, "ETHRewardDistributed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BondedECDSAKeepKeepClosedIterator is returned from FilterKeepClosed and is used to iterate over the raw logs and unpacked data for KeepClosed events raised by the BondedECDSAKeep contract.
type BondedECDSAKeepKeepClosedIterator struct {
	Event *BondedECDSAKeepKeepClosed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BondedECDSAKeepKeepClosedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BondedECDSAKeepKeepClosed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BondedECDSAKeepKeepClosed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BondedECDSAKeepKeepClosedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BondedECDSAKeepKeepClosedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BondedECDSAKeepKeepClosed represents a KeepClosed event raised by the BondedECDSAKeep contract.
type BondedECDSAKeepKeepClosed struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterKeepClosed is a free log retrieval operation binding the contract event 0x400fd7ee62b209afddce9dfbca204b2124c135597dff0ac92e9844e2b08927f6.
//
// Solidity: event KeepClosed()
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) FilterKeepClosed(opts *bind.FilterOpts) (*BondedECDSAKeepKeepClosedIterator, error) {

	logs, sub, err := _BondedECDSAKeep.contract.FilterLogs(opts, "KeepClosed")
	if err != nil {
		return nil, err
	}
	return &BondedECDSAKeepKeepClosedIterator{contract: _BondedECDSAKeep.contract, event: "KeepClosed", logs: logs, sub: sub}, nil
}

// WatchKeepClosed is a free log subscription operation binding the contract event 0x400fd7ee62b209afddce9dfbca204b2124c135597dff0ac92e9844e2b08927f6.
//
// Solidity: event KeepClosed()
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) WatchKeepClosed(opts *bind.WatchOpts, sink chan<- *BondedECDSAKeepKeepClosed) (event.Subscription, error) {

	logs, sub, err := _BondedECDSAKeep.contract.WatchLogs(opts, "KeepClosed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BondedECDSAKeepKeepClosed)
				if err := _BondedECDSAKeep.contract.UnpackLog(event, "KeepClosed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseKeepClosed is a log parse operation binding the contract event 0x400fd7ee62b209afddce9dfbca204b2124c135597dff0ac92e9844e2b08927f6.
//
// Solidity: event KeepClosed()
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) ParseKeepClosed(log types.Log) (*BondedECDSAKeepKeepClosed, error) {
	event := new(BondedECDSAKeepKeepClosed)
	if err := _BondedECDSAKeep.contract.UnpackLog(event, "KeepClosed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BondedECDSAKeepKeepTerminatedIterator is returned from FilterKeepTerminated and is used to iterate over the raw logs and unpacked data for KeepTerminated events raised by the BondedECDSAKeep contract.
type BondedECDSAKeepKeepTerminatedIterator struct {
	Event *BondedECDSAKeepKeepTerminated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BondedECDSAKeepKeepTerminatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BondedECDSAKeepKeepTerminated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BondedECDSAKeepKeepTerminated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BondedECDSAKeepKeepTerminatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BondedECDSAKeepKeepTerminatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BondedECDSAKeepKeepTerminated represents a KeepTerminated event raised by the BondedECDSAKeep contract.
type BondedECDSAKeepKeepTerminated struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterKeepTerminated is a free log retrieval operation binding the contract event 0x39f530c1293a870138e53618b826819a76f1fe86b5d500ba4622f9e8354a846a.
//
// Solidity: event KeepTerminated()
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) FilterKeepTerminated(opts *bind.FilterOpts) (*BondedECDSAKeepKeepTerminatedIterator, error) {

	logs, sub, err := _BondedECDSAKeep.contract.FilterLogs(opts, "KeepTerminated")
	if err != nil {
		return nil, err
	}
	return &BondedECDSAKeepKeepTerminatedIterator{contract: _BondedECDSAKeep.contract, event: "KeepTerminated", logs: logs, sub: sub}, nil
}

// WatchKeepTerminated is a free log subscription operation binding the contract event 0x39f530c1293a870138e53618b826819a76f1fe86b5d500ba4622f9e8354a846a.
//
// Solidity: event KeepTerminated()
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) WatchKeepTerminated(opts *bind.WatchOpts, sink chan<- *BondedECDSAKeepKeepTerminated) (event.Subscription, error) {

	logs, sub, err := _BondedECDSAKeep.contract.WatchLogs(opts, "KeepTerminated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BondedECDSAKeepKeepTerminated)
				if err := _BondedECDSAKeep.contract.UnpackLog(event, "KeepTerminated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseKeepTerminated is a log parse operation binding the contract event 0x39f530c1293a870138e53618b826819a76f1fe86b5d500ba4622f9e8354a846a.
//
// Solidity: event KeepTerminated()
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) ParseKeepTerminated(log types.Log) (*BondedECDSAKeepKeepTerminated, error) {
	event := new(BondedECDSAKeepKeepTerminated)
	if err := _BondedECDSAKeep.contract.UnpackLog(event, "KeepTerminated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BondedECDSAKeepPublicKeyPublishedIterator is returned from FilterPublicKeyPublished and is used to iterate over the raw logs and unpacked data for PublicKeyPublished events raised by the BondedECDSAKeep contract.
type BondedECDSAKeepPublicKeyPublishedIterator struct {
	Event *BondedECDSAKeepPublicKeyPublished // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BondedECDSAKeepPublicKeyPublishedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BondedECDSAKeepPublicKeyPublished)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BondedECDSAKeepPublicKeyPublished)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BondedECDSAKeepPublicKeyPublishedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BondedECDSAKeepPublicKeyPublishedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BondedECDSAKeepPublicKeyPublished represents a PublicKeyPublished event raised by the BondedECDSAKeep contract.
type BondedECDSAKeepPublicKeyPublished struct {
	PublicKey []byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterPublicKeyPublished is a free log retrieval operation binding the contract event 0xf62bba8b270bef3e8d0fcebc1f86567664da8ccbd03e8509d6231cc8d63f4b31.
//
// Solidity: event PublicKeyPublished(bytes publicKey)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) FilterPublicKeyPublished(opts *bind.FilterOpts) (*BondedECDSAKeepPublicKeyPublishedIterator, error) {

	logs, sub, err := _BondedECDSAKeep.contract.FilterLogs(opts, "PublicKeyPublished")
	if err != nil {
		return nil, err
	}
	return &BondedECDSAKeepPublicKeyPublishedIterator{contract: _BondedECDSAKeep.contract, event: "PublicKeyPublished", logs: logs, sub: sub}, nil
}

// WatchPublicKeyPublished is a free log subscription operation binding the contract event 0xf62bba8b270bef3e8d0fcebc1f86567664da8ccbd03e8509d6231cc8d63f4b31.
//
// Solidity: event PublicKeyPublished(bytes publicKey)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) WatchPublicKeyPublished(opts *bind.WatchOpts, sink chan<- *BondedECDSAKeepPublicKeyPublished) (event.Subscription, error) {

	logs, sub, err := _BondedECDSAKeep.contract.WatchLogs(opts, "PublicKeyPublished")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BondedECDSAKeepPublicKeyPublished)
				if err := _BondedECDSAKeep.contract.UnpackLog(event, "PublicKeyPublished", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePublicKeyPublished is a log parse operation binding the contract event 0xf62bba8b270bef3e8d0fcebc1f86567664da8ccbd03e8509d6231cc8d63f4b31.
//
// Solidity: event PublicKeyPublished(bytes publicKey)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) ParsePublicKeyPublished(log types.Log) (*BondedECDSAKeepPublicKeyPublished, error) {
	event := new(BondedECDSAKeepPublicKeyPublished)
	if err := _BondedECDSAKeep.contract.UnpackLog(event, "PublicKeyPublished", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BondedECDSAKeepSignatureRequestedIterator is returned from FilterSignatureRequested and is used to iterate over the raw logs and unpacked data for SignatureRequested events raised by the BondedECDSAKeep contract.
type BondedECDSAKeepSignatureRequestedIterator struct {
	Event *BondedECDSAKeepSignatureRequested // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BondedECDSAKeepSignatureRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BondedECDSAKeepSignatureRequested)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BondedECDSAKeepSignatureRequested)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BondedECDSAKeepSignatureRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BondedECDSAKeepSignatureRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BondedECDSAKeepSignatureRequested represents a SignatureRequested event raised by the BondedECDSAKeep contract.
type BondedECDSAKeepSignatureRequested struct {
	Digest [32]byte
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterSignatureRequested is a free log retrieval operation binding the contract event 0x34f611bedd4f8c135323bbfc4921e3f6e4feb7eef591036eed6af5462e6cfab0.
//
// Solidity: event SignatureRequested(bytes32 indexed digest)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) FilterSignatureRequested(opts *bind.FilterOpts, digest [][32]byte) (*BondedECDSAKeepSignatureRequestedIterator, error) {

	var digestRule []interface{}
	for _, digestItem := range digest {
		digestRule = append(digestRule, digestItem)
	}

	logs, sub, err := _BondedECDSAKeep.contract.FilterLogs(opts, "SignatureRequested", digestRule)
	if err != nil {
		return nil, err
	}
	return &BondedECDSAKeepSignatureRequestedIterator{contract: _BondedECDSAKeep.contract, event: "SignatureRequested", logs: logs, sub: sub}, nil
}

// WatchSignatureRequested is a free log subscription operation binding the contract event 0x34f611bedd4f8c135323bbfc4921e3f6e4feb7eef591036eed6af5462e6cfab0.
//
// Solidity: event SignatureRequested(bytes32 indexed digest)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) WatchSignatureRequested(opts *bind.WatchOpts, sink chan<- *BondedECDSAKeepSignatureRequested, digest [][32]byte) (event.Subscription, error) {

	var digestRule []interface{}
	for _, digestItem := range digest {
		digestRule = append(digestRule, digestItem)
	}

	logs, sub, err := _BondedECDSAKeep.contract.WatchLogs(opts, "SignatureRequested", digestRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BondedECDSAKeepSignatureRequested)
				if err := _BondedECDSAKeep.contract.UnpackLog(event, "SignatureRequested", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSignatureRequested is a log parse operation binding the contract event 0x34f611bedd4f8c135323bbfc4921e3f6e4feb7eef591036eed6af5462e6cfab0.
//
// Solidity: event SignatureRequested(bytes32 indexed digest)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) ParseSignatureRequested(log types.Log) (*BondedECDSAKeepSignatureRequested, error) {
	event := new(BondedECDSAKeepSignatureRequested)
	if err := _BondedECDSAKeep.contract.UnpackLog(event, "SignatureRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BondedECDSAKeepSignatureSubmittedIterator is returned from FilterSignatureSubmitted and is used to iterate over the raw logs and unpacked data for SignatureSubmitted events raised by the BondedECDSAKeep contract.
type BondedECDSAKeepSignatureSubmittedIterator struct {
	Event *BondedECDSAKeepSignatureSubmitted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BondedECDSAKeepSignatureSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BondedECDSAKeepSignatureSubmitted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BondedECDSAKeepSignatureSubmitted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BondedECDSAKeepSignatureSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BondedECDSAKeepSignatureSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BondedECDSAKeepSignatureSubmitted represents a SignatureSubmitted event raised by the BondedECDSAKeep contract.
type BondedECDSAKeepSignatureSubmitted struct {
	Digest     [32]byte
	R          [32]byte
	S          [32]byte
	RecoveryID uint8
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterSignatureSubmitted is a free log retrieval operation binding the contract event 0xb19546e9e0b503d103dd4ae295f4d526e9115adf7c902ead329b1f2404efd35f.
//
// Solidity: event SignatureSubmitted(bytes32 indexed digest, bytes32 r, bytes32 s, uint8 recoveryID)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) FilterSignatureSubmitted(opts *bind.FilterOpts, digest [][32]byte) (*BondedECDSAKeepSignatureSubmittedIterator, error) {

	var digestRule []interface{}
	for _, digestItem := range digest {
		digestRule = append(digestRule, digestItem)
	}

	logs, sub, err := _BondedECDSAKeep.contract.FilterLogs(opts, "SignatureSubmitted", digestRule)
	if err != nil {
		return nil, err
	}
	return &BondedECDSAKeepSignatureSubmittedIterator{contract: _BondedECDSAKeep.contract, event: "SignatureSubmitted", logs: logs, sub: sub}, nil
}

// WatchSignatureSubmitted is a free log subscription operation binding the contract event 0xb19546e9e0b503d103dd4ae295f4d526e9115adf7c902ead329b1f2404efd35f.
//
// Solidity: event SignatureSubmitted(bytes32 indexed digest, bytes32 r, bytes32 s, uint8 recoveryID)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) WatchSignatureSubmitted(opts *bind.WatchOpts, sink chan<- *BondedECDSAKeepSignatureSubmitted, digest [][32]byte) (event.Subscription, error) {

	var digestRule []interface{}
	for _, digestItem := range digest {
		digestRule = append(digestRule, digestItem)
	}

	logs, sub, err := _BondedECDSAKeep.contract.WatchLogs(opts, "SignatureSubmitted", digestRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BondedECDSAKeepSignatureSubmitted)
				if err := _BondedECDSAKeep.contract.UnpackLog(event, "SignatureSubmitted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSignatureSubmitted is a log parse operation binding the contract event 0xb19546e9e0b503d103dd4ae295f4d526e9115adf7c902ead329b1f2404efd35f.
//
// Solidity: event SignatureSubmitted(bytes32 indexed digest, bytes32 r, bytes32 s, uint8 recoveryID)
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) ParseSignatureSubmitted(log types.Log) (*BondedECDSAKeepSignatureSubmitted, error) {
	event := new(BondedECDSAKeepSignatureSubmitted)
	if err := _BondedECDSAKeep.contract.UnpackLog(event, "SignatureSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BondedECDSAKeepSlashingFailedIterator is returned from FilterSlashingFailed and is used to iterate over the raw logs and unpacked data for SlashingFailed events raised by the BondedECDSAKeep contract.
type BondedECDSAKeepSlashingFailedIterator struct {
	Event *BondedECDSAKeepSlashingFailed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BondedECDSAKeepSlashingFailedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BondedECDSAKeepSlashingFailed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BondedECDSAKeepSlashingFailed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BondedECDSAKeepSlashingFailedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BondedECDSAKeepSlashingFailedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BondedECDSAKeepSlashingFailed represents a SlashingFailed event raised by the BondedECDSAKeep contract.
type BondedECDSAKeepSlashingFailed struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterSlashingFailed is a free log retrieval operation binding the contract event 0x9caa8d499152520a1b3e11b41d51a51e5d1699294ebccdb9de0faa824dba8aae.
//
// Solidity: event SlashingFailed()
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) FilterSlashingFailed(opts *bind.FilterOpts) (*BondedECDSAKeepSlashingFailedIterator, error) {

	logs, sub, err := _BondedECDSAKeep.contract.FilterLogs(opts, "SlashingFailed")
	if err != nil {
		return nil, err
	}
	return &BondedECDSAKeepSlashingFailedIterator{contract: _BondedECDSAKeep.contract, event: "SlashingFailed", logs: logs, sub: sub}, nil
}

// WatchSlashingFailed is a free log subscription operation binding the contract event 0x9caa8d499152520a1b3e11b41d51a51e5d1699294ebccdb9de0faa824dba8aae.
//
// Solidity: event SlashingFailed()
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) WatchSlashingFailed(opts *bind.WatchOpts, sink chan<- *BondedECDSAKeepSlashingFailed) (event.Subscription, error) {

	logs, sub, err := _BondedECDSAKeep.contract.WatchLogs(opts, "SlashingFailed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BondedECDSAKeepSlashingFailed)
				if err := _BondedECDSAKeep.contract.UnpackLog(event, "SlashingFailed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSlashingFailed is a log parse operation binding the contract event 0x9caa8d499152520a1b3e11b41d51a51e5d1699294ebccdb9de0faa824dba8aae.
//
// Solidity: event SlashingFailed()
func (_BondedECDSAKeep *BondedECDSAKeepFilterer) ParseSlashingFailed(log types.Log) (*BondedECDSAKeepSlashingFailed, error) {
	event := new(BondedECDSAKeepSlashingFailed)
	if err := _BondedECDSAKeep.contract.UnpackLog(event, "SlashingFailed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
