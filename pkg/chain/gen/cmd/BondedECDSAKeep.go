// Code generated - DO NOT EDIT.
// This file is a generated command and any manual changes will be lost.

package cmd

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/cmd"
	"github.com/keep-network/keep-ecdsa/internal/config"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/contract"

	"github.com/urfave/cli"
)

var BondedECDSAKeepCommand cli.Command

var bondedECDSAKeepDescription = `The bonded-e-c-d-s-a-keep command allows calling the BondedECDSAKeep contract on an
	Ethereum network. It has subcommands corresponding to each contract method,
	which respectively each take parameters based on the contract method's
	parameters.

	Subcommands will submit a non-mutating call to the network and output the
	result.

	All subcommands can be called against a specific block by passing the
	-b/--block flag.

	All subcommands can be used to investigate the result of a previous
	transaction that called that same method by passing the -t/--transaction
	flag with the transaction hash.

	Subcommands for mutating methods may be submitted as a mutating transaction
	by passing the -s/--submit flag. In this mode, this command will terminate
	successfully once the transaction has been submitted, but will not wait for
	the transaction to be included in a block. They return the transaction hash.

	Calls that require ether to be paid will get 0 ether by default, which can
	be changed by passing the -v/--value flag.`

func init() {
	AvailableCommands = append(AvailableCommands, cli.Command{
		Name:        "bonded-e-c-d-s-a-keep",
		Usage:       `Provides access to the BondedECDSAKeep contract.`,
		Description: bondedECDSAKeepDescription,
		Subcommands: []cli.Command{{
			Name:      "is-active",
			Usage:     "Calls the constant method isActive on the BondedECDSAKeep contract.",
			ArgsUsage: "",
			Action:    becdsakIsActive,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "check-bond-amount",
			Usage:     "Calls the constant method checkBondAmount on the BondedECDSAKeep contract.",
			ArgsUsage: "",
			Action:    becdsakCheckBondAmount,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "get-member-e-t-h-balance",
			Usage:     "Calls the constant method getMemberETHBalance on the BondedECDSAKeep contract.",
			ArgsUsage: "[_member] ",
			Action:    becdsakGetMemberETHBalance,
			Before:    cmd.ArgCountChecker(1),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "members",
			Usage:     "Calls the constant method members on the BondedECDSAKeep contract.",
			ArgsUsage: "[arg0] ",
			Action:    becdsakMembers,
			Before:    cmd.ArgCountChecker(1),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "get-owner",
			Usage:     "Calls the constant method getOwner on the BondedECDSAKeep contract.",
			ArgsUsage: "",
			Action:    becdsakGetOwner,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "honest-threshold",
			Usage:     "Calls the constant method honestThreshold on the BondedECDSAKeep contract.",
			ArgsUsage: "",
			Action:    becdsakHonestThreshold,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "is-terminated",
			Usage:     "Calls the constant method isTerminated on the BondedECDSAKeep contract.",
			ArgsUsage: "",
			Action:    becdsakIsTerminated,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "public-key",
			Usage:     "Calls the constant method publicKey on the BondedECDSAKeep contract.",
			ArgsUsage: "",
			Action:    becdsakPublicKey,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "get-members",
			Usage:     "Calls the constant method getMembers on the BondedECDSAKeep contract.",
			ArgsUsage: "",
			Action:    becdsakGetMembers,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "owner",
			Usage:     "Calls the constant method owner on the BondedECDSAKeep contract.",
			ArgsUsage: "",
			Action:    becdsakOwner,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "digest",
			Usage:     "Calls the constant method digest on the BondedECDSAKeep contract.",
			ArgsUsage: "",
			Action:    becdsakDigest,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "get-opened-timestamp",
			Usage:     "Calls the constant method getOpenedTimestamp on the BondedECDSAKeep contract.",
			ArgsUsage: "",
			Action:    becdsakGetOpenedTimestamp,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "get-public-key",
			Usage:     "Calls the constant method getPublicKey on the BondedECDSAKeep contract.",
			ArgsUsage: "",
			Action:    becdsakGetPublicKey,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "is-closed",
			Usage:     "Calls the constant method isClosed on the BondedECDSAKeep contract.",
			ArgsUsage: "",
			Action:    becdsakIsClosed,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "member-stake",
			Usage:     "Calls the constant method memberStake on the BondedECDSAKeep contract.",
			ArgsUsage: "",
			Action:    becdsakMemberStake,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "submit-public-key",
			Usage:     "Calls the method submitPublicKey on the BondedECDSAKeep contract.",
			ArgsUsage: "[_publicKey] ",
			Action:    becdsakSubmitPublicKey,
			Before:    cli.BeforeFunc(cmd.NonConstArgsChecker.AndThen(cmd.ArgCountChecker(1))),
			Flags:     cmd.NonConstFlags,
		}, {
			Name:      "distribute-e-t-h-reward",
			Usage:     "Calls the payable method distributeETHReward on the BondedECDSAKeep contract.",
			ArgsUsage: "",
			Action:    becdsakDistributeETHReward,
			Before:    cli.BeforeFunc(cmd.PayableArgsChecker.AndThen(cmd.ArgCountChecker(0))),
			Flags:     cmd.PayableFlags,
		}, {
			Name:      "close-keep",
			Usage:     "Calls the method closeKeep on the BondedECDSAKeep contract.",
			ArgsUsage: "",
			Action:    becdsakCloseKeep,
			Before:    cli.BeforeFunc(cmd.NonConstArgsChecker.AndThen(cmd.ArgCountChecker(0))),
			Flags:     cmd.NonConstFlags,
		}, {
			Name:      "distribute-e-r-c20-reward",
			Usage:     "Calls the method distributeERC20Reward on the BondedECDSAKeep contract.",
			ArgsUsage: "[_tokenAddress] [_value] ",
			Action:    becdsakDistributeERC20Reward,
			Before:    cli.BeforeFunc(cmd.NonConstArgsChecker.AndThen(cmd.ArgCountChecker(2))),
			Flags:     cmd.NonConstFlags,
		}, {
			Name:      "return-partial-signer-bonds",
			Usage:     "Calls the payable method returnPartialSignerBonds on the BondedECDSAKeep contract.",
			ArgsUsage: "",
			Action:    becdsakReturnPartialSignerBonds,
			Before:    cli.BeforeFunc(cmd.PayableArgsChecker.AndThen(cmd.ArgCountChecker(0))),
			Flags:     cmd.PayableFlags,
		}, {
			Name:      "seize-signer-bonds",
			Usage:     "Calls the method seizeSignerBonds on the BondedECDSAKeep contract.",
			ArgsUsage: "",
			Action:    becdsakSeizeSignerBonds,
			Before:    cli.BeforeFunc(cmd.NonConstArgsChecker.AndThen(cmd.ArgCountChecker(0))),
			Flags:     cmd.NonConstFlags,
		}, {
			Name:      "withdraw",
			Usage:     "Calls the method withdraw on the BondedECDSAKeep contract.",
			ArgsUsage: "[_member] ",
			Action:    becdsakWithdraw,
			Before:    cli.BeforeFunc(cmd.NonConstArgsChecker.AndThen(cmd.ArgCountChecker(1))),
			Flags:     cmd.NonConstFlags,
		}},
	})
}

/// ------------------- Const methods -------------------

func becdsakIsActive(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	result, err := contract.IsActiveAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakCheckBondAmount(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	result, err := contract.CheckBondAmountAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakGetMemberETHBalance(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}
	_member, err := ethutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _member, a address, from passed value %v",
			c.Args()[0],
		)
	}

	result, err := contract.GetMemberETHBalanceAtBlock(
		_member,

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakMembers(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}
	arg0, err := hexutil.DecodeBig(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter arg0, a uint256, from passed value %v",
			c.Args()[0],
		)
	}

	result, err := contract.MembersAtBlock(
		arg0,

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakGetOwner(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	result, err := contract.GetOwnerAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakHonestThreshold(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	result, err := contract.HonestThresholdAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakIsTerminated(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	result, err := contract.IsTerminatedAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakPublicKey(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	result, err := contract.PublicKeyAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakGetMembers(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	result, err := contract.GetMembersAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakOwner(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	result, err := contract.OwnerAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakDigest(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	result, err := contract.DigestAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakGetOpenedTimestamp(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	result, err := contract.GetOpenedTimestampAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakGetPublicKey(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	result, err := contract.GetPublicKeyAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakIsClosed(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	result, err := contract.IsClosedAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakMemberStake(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	result, err := contract.MemberStakeAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

/// ------------------- Non-const methods -------------------

func becdsakSubmitPublicKey(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	_publicKey, err := hexutil.Decode(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _publicKey, a bytes, from passed value %v",
			c.Args()[0],
		)
	}

	var (
		transaction *types.Transaction
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.SubmitPublicKey(
			_publicKey,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		err = contract.CallSubmitPublicKey(
			_publicKey,
			cmd.BlockFlagValue.Uint,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(nil)
	}

	return nil
}

func becdsakDistributeETHReward(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	var (
		transaction *types.Transaction
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.DistributeETHReward(
			cmd.ValueFlagValue.Uint)
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		err = contract.CallDistributeETHReward(
			cmd.ValueFlagValue.Uint, cmd.BlockFlagValue.Uint,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(nil)
	}

	return nil
}

func becdsakCloseKeep(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	var (
		transaction *types.Transaction
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.CloseKeep()
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		err = contract.CallCloseKeep(
			cmd.BlockFlagValue.Uint,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(nil)
	}

	return nil
}

func becdsakDistributeERC20Reward(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	_tokenAddress, err := ethutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _tokenAddress, a address, from passed value %v",
			c.Args()[0],
		)
	}

	_value, err := hexutil.DecodeBig(c.Args()[1])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _value, a uint256, from passed value %v",
			c.Args()[1],
		)
	}

	var (
		transaction *types.Transaction
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.DistributeERC20Reward(
			_tokenAddress,
			_value,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		err = contract.CallDistributeERC20Reward(
			_tokenAddress,
			_value,
			cmd.BlockFlagValue.Uint,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(nil)
	}

	return nil
}

func becdsakReturnPartialSignerBonds(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	var (
		transaction *types.Transaction
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.ReturnPartialSignerBonds(
			cmd.ValueFlagValue.Uint)
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		err = contract.CallReturnPartialSignerBonds(
			cmd.ValueFlagValue.Uint, cmd.BlockFlagValue.Uint,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(nil)
	}

	return nil
}

func becdsakSeizeSignerBonds(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	var (
		transaction *types.Transaction
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.SeizeSignerBonds()
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		err = contract.CallSeizeSignerBonds(
			cmd.BlockFlagValue.Uint,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(nil)
	}

	return nil
}

func becdsakWithdraw(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeep(c)
	if err != nil {
		return err
	}

	_member, err := ethutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _member, a address, from passed value %v",
			c.Args()[0],
		)
	}

	var (
		transaction *types.Transaction
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.Withdraw(
			_member,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		err = contract.CallWithdraw(
			_member,
			cmd.BlockFlagValue.Uint,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(nil)
	}

	return nil
}

/// ------------------- Initialization -------------------

func initializeBondedECDSAKeep(c *cli.Context) (*contract.BondedECDSAKeep, error) {
	config, err := config.ReadEthereumConfig(c.GlobalString("config"))
	if err != nil {
		return nil, fmt.Errorf("error reading Ethereum config from file: [%v]", err)
	}

	client, _, _, err := ethutil.ConnectClients(config.URL, config.URLRPC)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Ethereum node: [%v]", err)
	}

	key, err := ethutil.DecryptKeyFile(
		config.Account.KeyFile,
		config.Account.KeyFilePassword,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read KeyFile: %s: [%v]",
			config.Account.KeyFile,
			err,
		)
	}

	checkInterval := cmd.DefaultMiningCheckInterval
	maxGasPrice := cmd.DefaultMaxGasPrice
	if config.MiningCheckInterval != 0 {
		checkInterval = time.Duration(config.MiningCheckInterval) * time.Second
	}
	if config.MaxGasPrice != 0 {
		maxGasPrice = new(big.Int).SetUint64(config.MaxGasPrice)
	}

	miningWaiter := ethutil.NewMiningWaiter(client, checkInterval, maxGasPrice)

	address := common.HexToAddress(config.ContractAddresses["BondedECDSAKeep"])

	return contract.NewBondedECDSAKeep(
		address,
		key,
		client,
		ethutil.NewNonceManager(key.Address, client),
		miningWaiter,
		&sync.Mutex{},
	)
}
