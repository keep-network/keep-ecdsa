package cmd

import (
	"fmt"

	"github.com/keep-network/keep-ecdsa/config"
	"github.com/keep-network/keep-ecdsa/pkg/chain/bitcoin"
	"github.com/keep-network/keep-ecdsa/pkg/extensions/tbtc/recovery"

	"github.com/urfave/cli"
)

// ResolveBitcoinBeneficiaryAddressCommand contains the definition of the
// resolve-bitcoin-address command-line subcommand.
var ResolveBitcoinBeneficiaryAddressCommand cli.Command

const resolveBitcoinAddressDescription = `Uses details provided in the configuration ` +
	`file to estimate next Bitcoin beneficiary address that will be used for liquidation ` +
	`recovery.`

func init() {
	ResolveBitcoinBeneficiaryAddressCommand =
		cli.Command{
			Name:        "resolve-bitcoin-address",
			Usage:       `Resolves next available bitcoin address`,
			Description: resolveBitcoinAddressDescription,
			Action:      ResolveBitcoinBeneficiaryAddress,
		}
}

// ResolveBitcoinBeneficiaryAddress resolves the next bitcoin address that would
// be used for liquidation recovery based on the nodes configuration.
func ResolveBitcoinBeneficiaryAddress(c *cli.Context) error {
	config, err := config.ReadConfig(c.GlobalString("config"))
	if err != nil {
		return fmt.Errorf("failed while reading config file: [%w]", err)
	}

	derivationIndexStorage, err := recovery.NewDerivationIndexStorage(config.Storage.DataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize new derivation index storage: [%w]", err)
	}

	err = config.Extensions.TBTC.Bitcoin.Validate()
	if err != nil {
		return fmt.Errorf("invalid bitcoin configuration: [%w]", err)
	}

	tbtcConfig := config.Extensions.TBTC

	chainParams, err := tbtcConfig.Bitcoin.ChainParams()
	if err != nil {
		return fmt.Errorf(
			"failed to parse the configured net params: [%w]",
			err,
		)
	}

	bitcoinHandle := bitcoin.Connect(tbtcConfig.Bitcoin.ElectrsURLWithDefault())

	beneficiaryAddress, err := recovery.ResolveAddress(
		tbtcConfig.Bitcoin.BeneficiaryAddress,
		derivationIndexStorage,
		chainParams,
		bitcoinHandle,
		true,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to resolve a btc address from [%s]: [%w]",
			tbtcConfig.Bitcoin.BeneficiaryAddress,
			err,
		)
	}

	logger.Infof("resolved bitcoin beneficiary address: %s", beneficiaryAddress)

	return nil
}
