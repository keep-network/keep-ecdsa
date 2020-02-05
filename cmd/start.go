package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ipfs/go-log"
	"github.com/urfave/cli"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-core/pkg/operator"

	"github.com/keep-network/keep-tecdsa/internal/config"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/ethereum"
	"github.com/keep-network/keep-tecdsa/pkg/client"
	"github.com/keep-network/keep-tecdsa/pkg/net/local"
)

var logger = log.Logger("keep-cmd")

// StartCommand contains the definition of the start command-line subcommand.
var StartCommand cli.Command

const startDescription = `Starts the Keep tECDSA client in the foreground.`

func init() {
	StartCommand =
		cli.Command{
			Name:        "start",
			Usage:       `Starts the Keep tECDSA client in the foreground`,
			Description: startDescription,
			Action:      Start,
		}
}

// Start starts a client.
func Start(c *cli.Context) error {
	config, err := config.ReadConfig(c.GlobalString("config"))
	if err != nil {
		return fmt.Errorf("failed while reading config file: [%v]", err)
	}

	ctx := context.Background()

	// TODO: This should be reverted when we've got network unicast implemented.
	// As a temporary solution to support multiple operators on one client we loop
	// over configured keys.
	for i, keyFile := range config.Ethereum.Account.KeyFile {
		ethereumKey, err := ethutil.DecryptKeyFile(
			keyFile,
			config.Ethereum.Account.KeyFilePassword,
		)
		if err != nil {
			return fmt.Errorf(
				"failed to read key file [%s]: [%v]", keyFile, err,
			)
		}

		ethereumChain, err := ethereum.Connect(ethereumKey.PrivateKey, &config.Ethereum)
		if err != nil {
			return fmt.Errorf("failed to connect to ethereum node: [%v]", err)
		}

		operatorPrivateKey, operatorPublicKey := operator.EthereumKeyToOperatorKey(ethereumKey)

		_, networkPublicKey := key.OperatorKeyToNetworkKey(
			operatorPrivateKey, operatorPublicKey,
		)

		networkProvider := local.LocalProvider(
			networkPublicKey,
		)

		dirPath := fmt.Sprintf("%s/membership_%d", config.Storage.DataDir, i)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			err = os.Mkdir(dirPath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("error occurred while creating a dir: [%v]", err)
			}
		}

		persistence := persistence.NewEncryptedPersistence(
			persistence.NewDiskHandle(dirPath),
			config.Ethereum.Account.KeyFilePassword,
		)

		sanctionedApplications, err := config.SanctionedApplications.Addresses()
		if err != nil {
			return fmt.Errorf("failed to get sanctioned applications addresses: [%v]", err)
		}

		registrationTicker, err := config.RegistrationTicker()
		if err != nil {
			return fmt.Errorf("failed to get registration ticker: [%v]", err)
		}

		client.Initialize(
			ethereumChain,
			networkProvider,
			persistence,
			sanctionedApplications,
		)
		logger.Debugf("initialized operator with address: [%s]", ethereumKey.Address.String())
	}

	logger.Info("client started")

	select {
	case <-ctx.Done():
		if err != nil {
			return err
		}

		return fmt.Errorf("unexpected context cancellation")
	}
}
