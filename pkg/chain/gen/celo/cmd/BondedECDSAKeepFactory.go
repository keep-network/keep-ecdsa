// Code generated - DO NOT EDIT.
// This file is a generated command and any manual changes will be lost.

package cmd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/common/hexutil"
	"github.com/celo-org/celo-blockchain/core/types"

	chainutil "github.com/keep-network/keep-common/pkg/chain/celo/celoutil"
	"github.com/keep-network/keep-common/pkg/cmd"
	"github.com/keep-network/keep-ecdsa/config"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/celo/contract"

	"github.com/urfave/cli"
)

var BondedECDSAKeepFactoryCommand cli.Command

var bondedECDSAKeepFactoryDescription = `The bonded-e-c-d-s-a-keep-factory command allows calling the BondedECDSAKeepFactory contract on an
	ETH-like network. It has subcommands corresponding to each contract method,
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
		Name:        "bonded-e-c-d-s-a-keep-factory",
		Usage:       `Provides access to the BondedECDSAKeepFactory contract.`,
		Description: bondedECDSAKeepFactoryDescription,
		Subcommands: []cli.Command{{
			Name:      "balance-of",
			Usage:     "Calls the constant method balanceOf on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[_operator] ",
			Action:    becdsakfBalanceOf,
			Before:    cmd.ArgCountChecker(1),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "callback-gas",
			Usage:     "Calls the constant method callbackGas on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "",
			Action:    becdsakfCallbackGas,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "get-keep-at-index",
			Usage:     "Calls the constant method getKeepAtIndex on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[index] ",
			Action:    becdsakfGetKeepAtIndex,
			Before:    cmd.ArgCountChecker(1),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "get-keep-count",
			Usage:     "Calls the constant method getKeepCount on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "",
			Action:    becdsakfGetKeepCount,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "get-keep-opened-timestamp",
			Usage:     "Calls the constant method getKeepOpenedTimestamp on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[_keep] ",
			Action:    becdsakfGetKeepOpenedTimestamp,
			Before:    cmd.ArgCountChecker(1),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "get-sortition-pool",
			Usage:     "Calls the constant method getSortitionPool on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[_application] ",
			Action:    becdsakfGetSortitionPool,
			Before:    cmd.ArgCountChecker(1),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "get-sortition-pool-weight",
			Usage:     "Calls the constant method getSortitionPoolWeight on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[_application] ",
			Action:    becdsakfGetSortitionPoolWeight,
			Before:    cmd.ArgCountChecker(1),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "group-selection-seed",
			Usage:     "Calls the constant method groupSelectionSeed on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "",
			Action:    becdsakfGroupSelectionSeed,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "has-minimum-stake",
			Usage:     "Calls the constant method hasMinimumStake on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[_operator] ",
			Action:    becdsakfHasMinimumStake,
			Before:    cmd.ArgCountChecker(1),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "is-operator-authorized",
			Usage:     "Calls the constant method isOperatorAuthorized on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[_operator] ",
			Action:    becdsakfIsOperatorAuthorized,
			Before:    cmd.ArgCountChecker(1),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "is-operator-eligible",
			Usage:     "Calls the constant method isOperatorEligible on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[_operator] [_application] ",
			Action:    becdsakfIsOperatorEligible,
			Before:    cmd.ArgCountChecker(2),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "is-operator-registered",
			Usage:     "Calls the constant method isOperatorRegistered on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[_operator] [_application] ",
			Action:    becdsakfIsOperatorRegistered,
			Before:    cmd.ArgCountChecker(2),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "is-operator-up-to-date",
			Usage:     "Calls the constant method isOperatorUpToDate on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[_operator] [_application] ",
			Action:    becdsakfIsOperatorUpToDate,
			Before:    cmd.ArgCountChecker(2),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "keeps",
			Usage:     "Calls the constant method keeps on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[arg0] ",
			Action:    becdsakfKeeps,
			Before:    cmd.ArgCountChecker(1),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "master-keep-address",
			Usage:     "Calls the constant method masterKeepAddress on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "",
			Action:    becdsakfMasterKeepAddress,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "minimum-bond",
			Usage:     "Calls the constant method minimumBond on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "",
			Action:    becdsakfMinimumBond,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "new-entry-fee-estimate",
			Usage:     "Calls the constant method newEntryFeeEstimate on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "",
			Action:    becdsakfNewEntryFeeEstimate,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "new-group-selection-seed-fee",
			Usage:     "Calls the constant method newGroupSelectionSeedFee on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "",
			Action:    becdsakfNewGroupSelectionSeedFee,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "open-keep-fee-estimate",
			Usage:     "Calls the constant method openKeepFeeEstimate on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "",
			Action:    becdsakfOpenKeepFeeEstimate,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "pool-stake-weight-divisor",
			Usage:     "Calls the constant method poolStakeWeightDivisor on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "",
			Action:    becdsakfPoolStakeWeightDivisor,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "reseed-pool",
			Usage:     "Calls the constant method reseedPool on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "",
			Action:    becdsakfReseedPool,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "beacon-callback",
			Usage:     "Calls the method beaconCallback on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[_relayEntry] ",
			Action:    becdsakfBeaconCallback,
			Before:    cli.BeforeFunc(cmd.NonConstArgsChecker.AndThen(cmd.ArgCountChecker(1))),
			Flags:     cmd.NonConstFlags,
		}, {
			Name:      "create-sortition-pool",
			Usage:     "Calls the method createSortitionPool on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[_application] ",
			Action:    becdsakfCreateSortitionPool,
			Before:    cli.BeforeFunc(cmd.NonConstArgsChecker.AndThen(cmd.ArgCountChecker(1))),
			Flags:     cmd.NonConstFlags,
		}, {
			Name:      "is-recognized",
			Usage:     "Calls the method isRecognized on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[_delegatedAuthorityRecipient] ",
			Action:    becdsakfIsRecognized,
			Before:    cli.BeforeFunc(cmd.NonConstArgsChecker.AndThen(cmd.ArgCountChecker(1))),
			Flags:     cmd.NonConstFlags,
		}, {
			Name:      "open-keep",
			Usage:     "Calls the payable method openKeep on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[_groupSize] [_honestThreshold] [_owner] [_bond] [_stakeLockDuration] ",
			Action:    becdsakfOpenKeep,
			Before:    cli.BeforeFunc(cmd.PayableArgsChecker.AndThen(cmd.ArgCountChecker(5))),
			Flags:     cmd.PayableFlags,
		}, {
			Name:      "register-member-candidate",
			Usage:     "Calls the method registerMemberCandidate on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[_application] ",
			Action:    becdsakfRegisterMemberCandidate,
			Before:    cli.BeforeFunc(cmd.NonConstArgsChecker.AndThen(cmd.ArgCountChecker(1))),
			Flags:     cmd.NonConstFlags,
		}, {
			Name:      "request-new-group-selection-seed",
			Usage:     "Calls the payable method requestNewGroupSelectionSeed on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "",
			Action:    becdsakfRequestNewGroupSelectionSeed,
			Before:    cli.BeforeFunc(cmd.PayableArgsChecker.AndThen(cmd.ArgCountChecker(0))),
			Flags:     cmd.PayableFlags,
		}, {
			Name:      "set-minimum-bondable-value",
			Usage:     "Calls the method setMinimumBondableValue on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[_minimumBondableValue] [_groupSize] [_honestThreshold] ",
			Action:    becdsakfSetMinimumBondableValue,
			Before:    cli.BeforeFunc(cmd.NonConstArgsChecker.AndThen(cmd.ArgCountChecker(3))),
			Flags:     cmd.NonConstFlags,
		}, {
			Name:      "update-operator-status",
			Usage:     "Calls the method updateOperatorStatus on the BondedECDSAKeepFactory contract.",
			ArgsUsage: "[_operator] [_application] ",
			Action:    becdsakfUpdateOperatorStatus,
			Before:    cli.BeforeFunc(cmd.NonConstArgsChecker.AndThen(cmd.ArgCountChecker(2))),
			Flags:     cmd.NonConstFlags,
		}},
	})
}

/// ------------------- Const methods -------------------

func becdsakfBalanceOf(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}
	_operator, err := chainutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _operator, a address, from passed value %v",
			c.Args()[0],
		)
	}

	result, err := contract.BalanceOfAtBlock(
		_operator,

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfCallbackGas(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	result, err := contract.CallbackGasAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfGetKeepAtIndex(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}
	index, err := hexutil.DecodeBig(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter index, a uint256, from passed value %v",
			c.Args()[0],
		)
	}

	result, err := contract.GetKeepAtIndexAtBlock(
		index,

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfGetKeepCount(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	result, err := contract.GetKeepCountAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfGetKeepOpenedTimestamp(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}
	_keep, err := chainutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _keep, a address, from passed value %v",
			c.Args()[0],
		)
	}

	result, err := contract.GetKeepOpenedTimestampAtBlock(
		_keep,

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfGetSortitionPool(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}
	_application, err := chainutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _application, a address, from passed value %v",
			c.Args()[0],
		)
	}

	result, err := contract.GetSortitionPoolAtBlock(
		_application,

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfGetSortitionPoolWeight(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}
	_application, err := chainutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _application, a address, from passed value %v",
			c.Args()[0],
		)
	}

	result, err := contract.GetSortitionPoolWeightAtBlock(
		_application,

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfGroupSelectionSeed(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	result, err := contract.GroupSelectionSeedAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfHasMinimumStake(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}
	_operator, err := chainutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _operator, a address, from passed value %v",
			c.Args()[0],
		)
	}

	result, err := contract.HasMinimumStakeAtBlock(
		_operator,

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfIsOperatorAuthorized(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}
	_operator, err := chainutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _operator, a address, from passed value %v",
			c.Args()[0],
		)
	}

	result, err := contract.IsOperatorAuthorizedAtBlock(
		_operator,

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfIsOperatorEligible(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}
	_operator, err := chainutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _operator, a address, from passed value %v",
			c.Args()[0],
		)
	}

	_application, err := chainutil.AddressFromHex(c.Args()[1])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _application, a address, from passed value %v",
			c.Args()[1],
		)
	}

	result, err := contract.IsOperatorEligibleAtBlock(
		_operator,
		_application,

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfIsOperatorRegistered(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}
	_operator, err := chainutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _operator, a address, from passed value %v",
			c.Args()[0],
		)
	}

	_application, err := chainutil.AddressFromHex(c.Args()[1])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _application, a address, from passed value %v",
			c.Args()[1],
		)
	}

	result, err := contract.IsOperatorRegisteredAtBlock(
		_operator,
		_application,

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfIsOperatorUpToDate(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}
	_operator, err := chainutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _operator, a address, from passed value %v",
			c.Args()[0],
		)
	}

	_application, err := chainutil.AddressFromHex(c.Args()[1])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _application, a address, from passed value %v",
			c.Args()[1],
		)
	}

	result, err := contract.IsOperatorUpToDateAtBlock(
		_operator,
		_application,

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfKeeps(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
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

	result, err := contract.KeepsAtBlock(
		arg0,

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfMasterKeepAddress(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	result, err := contract.MasterKeepAddressAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfMinimumBond(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	result, err := contract.MinimumBondAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfNewEntryFeeEstimate(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	result, err := contract.NewEntryFeeEstimateAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfNewGroupSelectionSeedFee(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	result, err := contract.NewGroupSelectionSeedFeeAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfOpenKeepFeeEstimate(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	result, err := contract.OpenKeepFeeEstimateAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfPoolStakeWeightDivisor(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	result, err := contract.PoolStakeWeightDivisorAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakfReseedPool(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	result, err := contract.ReseedPoolAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

/// ------------------- Non-const methods -------------------

func becdsakfBeaconCallback(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	_relayEntry, err := hexutil.DecodeBig(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _relayEntry, a uint256, from passed value %v",
			c.Args()[0],
		)
	}

	var (
		transaction *types.Transaction
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.BeaconCallback(
			_relayEntry,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		err = contract.CallBeaconCallback(
			_relayEntry,
			cmd.BlockFlagValue.Uint,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(nil)
	}

	return nil
}

func becdsakfCreateSortitionPool(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	_application, err := chainutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _application, a address, from passed value %v",
			c.Args()[0],
		)
	}

	var (
		transaction *types.Transaction
		result      common.Address
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.CreateSortitionPool(
			_application,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		result, err = contract.CallCreateSortitionPool(
			_application,
			cmd.BlockFlagValue.Uint,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(result)
	}

	return nil
}

func becdsakfIsRecognized(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	_delegatedAuthorityRecipient, err := chainutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _delegatedAuthorityRecipient, a address, from passed value %v",
			c.Args()[0],
		)
	}

	var (
		transaction *types.Transaction
		result      bool
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.IsRecognized(
			_delegatedAuthorityRecipient,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		result, err = contract.CallIsRecognized(
			_delegatedAuthorityRecipient,
			cmd.BlockFlagValue.Uint,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(result)
	}

	return nil
}

func becdsakfOpenKeep(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	_groupSize, err := hexutil.DecodeBig(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _groupSize, a uint256, from passed value %v",
			c.Args()[0],
		)
	}

	_honestThreshold, err := hexutil.DecodeBig(c.Args()[1])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _honestThreshold, a uint256, from passed value %v",
			c.Args()[1],
		)
	}

	_owner, err := chainutil.AddressFromHex(c.Args()[2])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _owner, a address, from passed value %v",
			c.Args()[2],
		)
	}

	_bond, err := hexutil.DecodeBig(c.Args()[3])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _bond, a uint256, from passed value %v",
			c.Args()[3],
		)
	}

	_stakeLockDuration, err := hexutil.DecodeBig(c.Args()[4])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _stakeLockDuration, a uint256, from passed value %v",
			c.Args()[4],
		)
	}

	var (
		transaction *types.Transaction
		result      common.Address
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.OpenKeep(
			_groupSize,
			_honestThreshold,
			_owner,
			_bond,
			_stakeLockDuration,
			cmd.ValueFlagValue.Uint)
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		result, err = contract.CallOpenKeep(
			_groupSize,
			_honestThreshold,
			_owner,
			_bond,
			_stakeLockDuration,
			cmd.ValueFlagValue.Uint, cmd.BlockFlagValue.Uint,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(result)
	}

	return nil
}

func becdsakfRegisterMemberCandidate(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	_application, err := chainutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _application, a address, from passed value %v",
			c.Args()[0],
		)
	}

	var (
		transaction *types.Transaction
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.RegisterMemberCandidate(
			_application,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		err = contract.CallRegisterMemberCandidate(
			_application,
			cmd.BlockFlagValue.Uint,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(nil)
	}

	return nil
}

func becdsakfRequestNewGroupSelectionSeed(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	var (
		transaction *types.Transaction
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.RequestNewGroupSelectionSeed(
			cmd.ValueFlagValue.Uint)
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		err = contract.CallRequestNewGroupSelectionSeed(
			cmd.ValueFlagValue.Uint, cmd.BlockFlagValue.Uint,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(nil)
	}

	return nil
}

func becdsakfSetMinimumBondableValue(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	_minimumBondableValue, err := hexutil.DecodeBig(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _minimumBondableValue, a uint256, from passed value %v",
			c.Args()[0],
		)
	}

	_groupSize, err := hexutil.DecodeBig(c.Args()[1])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _groupSize, a uint256, from passed value %v",
			c.Args()[1],
		)
	}

	_honestThreshold, err := hexutil.DecodeBig(c.Args()[2])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _honestThreshold, a uint256, from passed value %v",
			c.Args()[2],
		)
	}

	var (
		transaction *types.Transaction
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.SetMinimumBondableValue(
			_minimumBondableValue,
			_groupSize,
			_honestThreshold,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		err = contract.CallSetMinimumBondableValue(
			_minimumBondableValue,
			_groupSize,
			_honestThreshold,
			cmd.BlockFlagValue.Uint,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(nil)
	}

	return nil
}

func becdsakfUpdateOperatorStatus(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepFactory(c)
	if err != nil {
		return err
	}

	_operator, err := chainutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _operator, a address, from passed value %v",
			c.Args()[0],
		)
	}

	_application, err := chainutil.AddressFromHex(c.Args()[1])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _application, a address, from passed value %v",
			c.Args()[1],
		)
	}

	var (
		transaction *types.Transaction
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.UpdateOperatorStatus(
			_operator,
			_application,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		err = contract.CallUpdateOperatorStatus(
			_operator,
			_application,
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

func initializeBondedECDSAKeepFactory(c *cli.Context) (*contract.BondedECDSAKeepFactory, error) {
	config, err := config.ReadCeloConfig(c.GlobalString("config"))
	if err != nil {
		return nil, fmt.Errorf("error reading config from file: [%v]", err)
	}

	client, _, _, err := chainutil.ConnectClients(config.URL, config.URLRPC)
	if err != nil {
		return nil, fmt.Errorf("error connecting to host chain node: [%v]", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf(
			"failed to resolve host chain id: [%v]",
			err,
		)
	}

	key, err := chainutil.DecryptKeyFile(
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
	if config.MaxGasPrice != nil {
		maxGasPrice = config.MaxGasPrice.Int
	}

	miningWaiter := chainutil.NewMiningWaiter(
		client,
		checkInterval,
		maxGasPrice,
	)

	blockCounter, err := chainutil.NewBlockCounter(client)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create block counter: [%v]",
			err,
		)
	}

	address := common.HexToAddress(config.ContractAddresses["BondedECDSAKeepFactory"])

	return contract.NewBondedECDSAKeepFactory(
		address,
		chainID,
		key,
		client,
		chainutil.NewNonceManager(client, key.Address),
		miningWaiter,
		blockCounter,
		&sync.Mutex{},
	)
}
