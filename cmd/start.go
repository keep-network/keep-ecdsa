package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/persistence"

	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-core/pkg/net/libp2p"
	"github.com/keep-network/keep-core/pkg/net/retransmission"
	"github.com/keep-network/keep-core/pkg/operator"

	"github.com/keep-network/keep-tecdsa/internal/config"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/ethereum"
	"github.com/keep-network/keep-tecdsa/pkg/client"

	"github.com/urfave/cli"
)

var logger = log.Logger("keep-cmd")

// StartCommand contains the definition of the start command-line subcommand.
var StartCommand cli.Command

const startDescription = `Starts the Keep tECDSA client in the foreground.`

// Constants related with network.
//
// In order to communicate, nodes in the network should have a connection
// between them. Basically a node can:
// - receive a connection from another peer
// - automatically open a connection to another peer
//   during core bootstrap round
// - automatically open a connection to another peer
// 	 after routing table refresh (DHT bootstrap round)
//
// Ideally, each node in the network should have a connection with all
// other nodes or at least be aware of their existence. This strongly depends
// on the actual network topology but some parameters can be fine-tuned
// in order to improve the behavior of the network.
const (
	// routingTableRefreshPeriod determines the frequency of routing table
	// refreshes. Routing table is actually a structure which contains
	// transport identifiers of other network peers along with their
	// addresses. A refresh of the routing table is basically a query
	// sent to connected peers asking about new entries from their routing
	// tables. If the node receives an information about new peers it will
	// try to connect them automatically.
	//
	// The refresh period should be set to a value which will
	// allow to keep the routing table up to date with the actual
	// network state. Bear in mind a smaller value may not have sense
	// as changes in the network need some time to propagate and frequent
	// refreshes can increase resource consumption and network congestion.
	routingTableRefreshPeriod = 5 * time.Minute

	// bootstrapMinPeerThreshold determines the minimum number of peers
	// that the node will try to keep connections with. Number of active
	// connections is checked during core bootstrap rounds and if this number
	// is less than the minimum, new connection attempts will be performed
	// against peers listed in config (`LibP2P.Peers`).
	bootstrapMinPeerThreshold = 10
)

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

	operatorPrivateKey, operatorPublicKey := operator.EthereumKeyToOperatorKey(ethereumKey)

	networkPrivateKey, _ := key.OperatorKeyToNetworkKey(
		operatorPrivateKey, operatorPublicKey,
	)

	networkProvider, err := libp2p.Connect(
		ctx,
		config.LibP2P,
		networkPrivateKey,
		ethereum.NewEthereumStakeMonitor(),
		retransmission.NewTicker(make(chan uint64)),
		libp2p.WithRoutingTableRefreshPeriod(routingTableRefreshPeriod),
		libp2p.WithBootstrapMinPeerThreshold(bootstrapMinPeerThreshold),
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
