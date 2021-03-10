// Code generated - DO NOT EDIT.
// This file is a generated command and any manual changes will be lost.

package cmd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"

	chainutil "github.com/keep-network/keep-common/pkg/chain/celo/celoutil"
	"github.com/keep-network/keep-common/pkg/cmd"
	"github.com/keep-network/keep-ecdsa/config"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/celo/contract"

	"github.com/urfave/cli"
)

var BondedECDSAKeepVendorCommand cli.Command

var bondedECDSAKeepVendorDescription = `The bonded-e-c-d-s-a-keep-vendor command allows calling the BondedECDSAKeepVendor contract on an
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
		Name:        "bonded-e-c-d-s-a-keep-vendor",
		Usage:       `Provides access to the BondedECDSAKeepVendor contract.`,
		Description: bondedECDSAKeepVendorDescription,
		Subcommands: []cli.Command{{
			Name:      "factory-upgrade-time-delay",
			Usage:     "Calls the constant method factoryUpgradeTimeDelay on the BondedECDSAKeepVendor contract.",
			ArgsUsage: "",
			Action:    becdsakvFactoryUpgradeTimeDelay,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "initialized",
			Usage:     "Calls the constant method initialized on the BondedECDSAKeepVendor contract.",
			ArgsUsage: "",
			Action:    becdsakvInitialized,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "select-factory",
			Usage:     "Calls the constant method selectFactory on the BondedECDSAKeepVendor contract.",
			ArgsUsage: "",
			Action:    becdsakvSelectFactory,
			Before:    cmd.ArgCountChecker(0),
			Flags:     cmd.ConstFlags,
		}, {
			Name:      "complete-factory-upgrade",
			Usage:     "Calls the method completeFactoryUpgrade on the BondedECDSAKeepVendor contract.",
			ArgsUsage: "",
			Action:    becdsakvCompleteFactoryUpgrade,
			Before:    cli.BeforeFunc(cmd.NonConstArgsChecker.AndThen(cmd.ArgCountChecker(0))),
			Flags:     cmd.NonConstFlags,
		}, {
			Name:      "initialize",
			Usage:     "Calls the method initialize on the BondedECDSAKeepVendor contract.",
			ArgsUsage: "[registryAddress] [factory] ",
			Action:    becdsakvInitialize,
			Before:    cli.BeforeFunc(cmd.NonConstArgsChecker.AndThen(cmd.ArgCountChecker(2))),
			Flags:     cmd.NonConstFlags,
		}, {
			Name:      "upgrade-factory",
			Usage:     "Calls the method upgradeFactory on the BondedECDSAKeepVendor contract.",
			ArgsUsage: "[_factory] ",
			Action:    becdsakvUpgradeFactory,
			Before:    cli.BeforeFunc(cmd.NonConstArgsChecker.AndThen(cmd.ArgCountChecker(1))),
			Flags:     cmd.NonConstFlags,
		}},
	})
}

/// ------------------- Const methods -------------------

func becdsakvFactoryUpgradeTimeDelay(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepVendor(c)
	if err != nil {
		return err
	}

	result, err := contract.FactoryUpgradeTimeDelayAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakvInitialized(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepVendor(c)
	if err != nil {
		return err
	}

	result, err := contract.InitializedAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

func becdsakvSelectFactory(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepVendor(c)
	if err != nil {
		return err
	}

	result, err := contract.SelectFactoryAtBlock(

		cmd.BlockFlagValue.Uint,
	)

	if err != nil {
		return err
	}

	cmd.PrintOutput(result)

	return nil
}

/// ------------------- Non-const methods -------------------

func becdsakvCompleteFactoryUpgrade(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepVendor(c)
	if err != nil {
		return err
	}

	var (
		transaction *types.Transaction
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.CompleteFactoryUpgrade()
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		err = contract.CallCompleteFactoryUpgrade(
			cmd.BlockFlagValue.Uint,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(nil)
	}

	return nil
}

func becdsakvInitialize(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepVendor(c)
	if err != nil {
		return err
	}

	registryAddress, err := chainutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter registryAddress, a address, from passed value %v",
			c.Args()[0],
		)
	}

	factory, err := chainutil.AddressFromHex(c.Args()[1])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter factory, a address, from passed value %v",
			c.Args()[1],
		)
	}

	var (
		transaction *types.Transaction
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.Initialize(
			registryAddress,
			factory,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		err = contract.CallInitialize(
			registryAddress,
			factory,
			cmd.BlockFlagValue.Uint,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(nil)
	}

	return nil
}

func becdsakvUpgradeFactory(c *cli.Context) error {
	contract, err := initializeBondedECDSAKeepVendor(c)
	if err != nil {
		return err
	}

	_factory, err := chainutil.AddressFromHex(c.Args()[0])
	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter _factory, a address, from passed value %v",
			c.Args()[0],
		)
	}

	var (
		transaction *types.Transaction
	)

	if c.Bool(cmd.SubmitFlag) {
		// Do a regular submission. Take payable into account.
		transaction, err = contract.UpgradeFactory(
			_factory,
		)
		if err != nil {
			return err
		}

		cmd.PrintOutput(transaction.Hash)
	} else {
		// Do a call.
		err = contract.CallUpgradeFactory(
			_factory,
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

func initializeBondedECDSAKeepVendor(c *cli.Context) (*contract.BondedECDSAKeepVendor, error) {
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

	address := common.HexToAddress(config.ContractAddresses["BondedECDSAKeepVendor"])

	return contract.NewBondedECDSAKeepVendor(
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
