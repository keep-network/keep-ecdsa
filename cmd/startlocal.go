package cmd

import (
	"context"
	"fmt"
	"math/big"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-core/pkg/operator"
	"github.com/keep-network/keep-tecdsa/internal/config"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/ethereum"
	"github.com/keep-network/keep-tecdsa/pkg/client"
	"github.com/keep-network/keep-tecdsa/pkg/net/local"
	"github.com/urfave/cli"
)

// StartLocalCommand contains the definition of the local-signer command-line
// subcommand.
var StartLocalCommand cli.Command

const startLocalDescription = `Starts three Keep tECDSA client in the foreground in 
local signer mode. It requires three config files named 'config_1.toml', 'config_2.toml', 
'config_3.toml' for each client to be provided in 'configs' directory.`

func init() {
	StartLocalCommand = cli.Command{
		Name:        "start-local",
		Usage:       "Starts local signer version of Keep ECDSA client",
		Description: startLocalDescription,
		Action:      StartLocal,
	}
}

func StartLocal(c *cli.Context) error {
	groupSize := 3

	ctx := context.Background()
	var err error

	for index := 1; index <= groupSize; index++ {
		configPath := fmt.Sprintf("./configs/config_%d.toml", index)

		config, err := config.ReadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed while reading config file: [%v]", err)
		}

		ethereumKey, err := ethutil.DecryptKeyFile(
			config.Ethereum.Account.KeyFile,
			config.Ethereum.Account.KeyFilePassword,
		)
		if err != nil {
			return fmt.Errorf(
				"failed to read key file [%s]: [%v]", config.Ethereum.Account.KeyFile, err,
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

		transportID := new(big.Int).SetBytes(
			[]byte(ethereumChain.Address().String()),
		).String()

		netErrChan := make(chan error)
		go func() {
			for {
				logger.Errorf("network error ocurred: [%v]", <-netErrChan)
			}
		}()

		networkProvider := local.LocalProvider(
			transportID,
			networkPublicKey,
			netErrChan,
		)

		persistence := persistence.NewEncryptedPersistence(
			persistence.NewDiskHandle(config.Storage.DataDir),
			config.Ethereum.Account.KeyFilePassword,
		)

		client.Initialize(
			ethereumChain,
			networkProvider,
			persistence,
		)

		logger.Infof("client %d started", index)

	}

	select {
	case <-ctx.Done():
		if err != nil {
			return err
		}

		return fmt.Errorf("unexpected context cancellation")
	}

	return nil
}
