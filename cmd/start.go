package cmd

import (
	"context"
	"fmt"
	"math/big"

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

	ctx := context.Background()

	client.Initialize(
		ethereumChain,
		networkProvider,
		persistence,
	)

	logger.Info("client started")

	select {
	case <-ctx.Done():
		if err != nil {
			return err
		}

		return fmt.Errorf("unexpected context cancellation")
	}
}
