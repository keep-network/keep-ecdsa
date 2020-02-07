package cmd

import (
	"context"
	"fmt"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-core/pkg/net/libp2p"
	"github.com/keep-network/keep-core/pkg/net/retransmission"
	"github.com/urfave/cli"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-core/pkg/operator"

	"github.com/keep-network/keep-tecdsa/internal/config"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/ethereum"
	"github.com/keep-network/keep-tecdsa/pkg/client"
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

	ethereumKey, err := ethutil.DecryptKeyFile(
		config.Ethereum.Account.KeyFile,
		config.Ethereum.Account.KeyFilePassword,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to read key file [%s]: [%v]",
			config.Ethereum.Account.KeyFile,
			err,
		)
	}

	ethereumChain, err := ethereum.Connect(ethereumKey.PrivateKey, &config.Ethereum)
	if err != nil {
		return fmt.Errorf("failed to connect to ethereum node: [%v]", err)
	}

	stakeMonitor, err := ethereumChain.StakeMonitor()
	if err != nil {
		return fmt.Errorf("error obtaining stake monitor handle [%v]", err)
	}
	hasMinimumStake, err := stakeMonitor.HasMinimumStake(
		ethereumKey.Address.Hex(),
	)
	if err != nil {
		return fmt.Errorf("could not check the stake [%v]", err)
	}
	if !hasMinimumStake {
		return fmt.Errorf("stake is below the required minimum")
	}

	operatorPrivateKey, operatorPublicKey := operator.EthereumKeyToOperatorKey(ethereumKey)

	networkPrivateKey, _ := key.OperatorKeyToNetworkKey(
		operatorPrivateKey, operatorPublicKey,
	)

	networkProvider, err := libp2p.Connect(
		ctx,
		config.LibP2P,
		networkPrivateKey,
		stakeMonitor,
		retransmission.NewTicker(make(chan uint64)),
	)
	if err != nil {
		return err
	}

	persistence := persistence.NewEncryptedPersistence(
		persistence.NewDiskHandle(config.Storage.DataDir),
		config.Ethereum.Account.KeyFilePassword,
	)

	sanctionedApplications, err := config.SanctionedApplications.Addresses()
	if err != nil {
		return fmt.Errorf("failed to get sanctioned applications addresses: [%v]", err)
	}

	client.Initialize(
		ethereumChain,
		networkProvider,
		persistence,
		sanctionedApplications,
	)
	logger.Debugf("initialized operator with address: [%s]", ethereumKey.Address.String())

	logger.Info("client started")

	select {
	case <-ctx.Done():
		if err != nil {
			return err
		}

		return fmt.Errorf("unexpected context cancellation")
	}
}
