package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-ecdsa/internal/config"
	"github.com/keep-network/keep-ecdsa/pkg/registry"
	"github.com/urfave/cli"
)

var SigningCommand cli.Command

func init() {
	SigningCommand = cli.Command{
		Name:  "signing",
		Usage: "Provides several tools useful for out-of-band signing",
		Subcommands: []cli.Command{
			{
				Name: "decrypt-key-share",
				Usage: "Decrypts the key share of the operator for the given " +
					"keep and stores it in a file",
				Action: DecryptKeyShare,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "keep-address,k",
						Usage: "address of the keep",
					},
				},
			},
		},
	}
}

func DecryptKeyShare(c *cli.Context) error {
	config, err := config.ReadConfig(c.GlobalString("config"))
	if err != nil {
		return fmt.Errorf("failed while reading config file: [%v]", err)
	}

	keyFile, err := ethutil.DecryptKeyFile(
		config.Ethereum.Account.KeyFile,
		config.Ethereum.Account.KeyFilePassword,
	)
	if err != nil {
		return fmt.Errorf("failed to decrypt key file: [%v]", err)
	}

	keepAddressHex := c.String("keep-address")
	if !common.IsHexAddress(keepAddressHex) {
		return fmt.Errorf("invalid keep address")
	}

	keepAddress := common.HexToAddress(keepAddressHex)

	handle, err := persistence.NewDiskHandle(config.Storage.DataDir)
	if err != nil {
		return fmt.Errorf(
			"failed while creating a storage disk handler: [%v]",
			err,
		)
	}

	persistence := persistence.NewEncryptedPersistence(
		handle,
		config.Ethereum.Account.KeyFilePassword,
	)

	keepRegistry := registry.NewKeepsRegistry(persistence)

	keepRegistry.LoadExistingKeeps()

	signers, err := keepRegistry.GetSigners(keepAddress)
	if err != nil {
		return fmt.Errorf(
			"no signers for keep [%s]: [%v]",
			keepAddress.String(),
			err,
		)
	}

	signer := signers[0]

	signerBytes, err := signer.Marshal()
	if err != nil {
		return fmt.Errorf(
			"failed to marshall signer for keep [%s]: [%v]",
			keepAddress.String(),
			err,
		)
	}

	targetFilePath := fmt.Sprintf(
		"key_share_%.10s_%.10s",
		keepAddress.String(),
		keyFile.Address.String(),
	)

	if _, err := os.Stat(targetFilePath); !os.IsNotExist(err) {
		return fmt.Errorf(
			"could not write shares to file; file [%s] already exists",
			targetFilePath,
		)
	}

	err = ioutil.WriteFile(targetFilePath, signerBytes, 0444) // read-only
	if err != nil {
		return fmt.Errorf(
			"failed to write to file [%s]: [%v]",
			targetFilePath,
			err,
		)
	}

	logger.Infof(
		"key share has been decrypted successfully and written to file [%s]",
		targetFilePath,
	)

	return nil
}
