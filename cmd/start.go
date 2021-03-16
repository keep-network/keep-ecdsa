package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/keep-network/keep-ecdsa/pkg/metrics"

	"github.com/keep-network/keep-core/pkg/diagnostics"

	"github.com/keep-network/keep-core/pkg/chain"
	coreMetrics "github.com/keep-network/keep-core/pkg/metrics"
	"github.com/keep-network/keep-core/pkg/net"

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/persistence"

	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-core/pkg/net/libp2p"
	"github.com/keep-network/keep-core/pkg/net/retransmission"
	"github.com/keep-network/keep-ecdsa/config"
	"github.com/keep-network/keep-ecdsa/pkg/client"
	"github.com/keep-network/keep-ecdsa/pkg/firewall"

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

	chainHandle, operatorKeys, err := connectChain(ctx, config)
	if err != nil {
		return err
	}

	stakeMonitor, err := chainHandle.StakeMonitor()
	if err != nil {
		return fmt.Errorf("error obtaining stake monitor handle: [%v]", err)
	}
	hasMinimumStake, err := stakeMonitor.HasMinimumStake(
		chainHandle.Address().Hex(),
	)
	if err != nil {
		return fmt.Errorf("could not check the stake: [%v]", err)
	}
	if !hasMinimumStake {
		logger.Errorf(
			"no minimum KEEP stake or operator is not authorized to use it; " +
				"please make sure the operator address in the configuration " +
				"is correct and it has KEEP tokens delegated and the operator " +
				"contract has been authorized to operate on the stake",
		)
	}

	networkPrivateKey, _ := key.OperatorKeyToNetworkKey(
		operatorKeys.private, operatorKeys.public,
	)

	networkProvider, err := libp2p.Connect(
		ctx,
		config.LibP2P,
		networkPrivateKey,
		libp2p.ProtocolECDSA,
		firewall.NewStakeOrActiveKeepPolicy(chainHandle, stakeMonitor),
		retransmission.NewTimeTicker(ctx, 1*time.Second),
		libp2p.WithRoutingTableRefreshPeriod(routingTableRefreshPeriod),
	)
	if err != nil {
		return err
	}

	nodeHeader(networkProvider.ConnectionManager().AddrStrings(), config.LibP2P.Port)

	handle, err := persistence.NewDiskHandle(config.Storage.DataDir)
	if err != nil {
		return fmt.Errorf("failed while creating a storage disk handler: [%v]", err)
	}
	persistence := persistence.NewEncryptedPersistence(
		handle,
		extractKeyFilePassword(config),
	)

	clientHandle := client.Initialize(
		ctx,
		operatorKeys.public,
		chainHandle,
		networkProvider,
		persistence,
		&config.Client,
		&config.Extensions.TBTC.BTCRefunds,
		&config.TSS,
	)
	logger.Debugf("initialized operator with address: [%s]", chainHandle.Address().String())

	initializeMetrics(ctx, config, networkProvider, stakeMonitor, chainHandle.Address().Hex(), clientHandle)
	initializeDiagnostics(config, networkProvider)

	logger.Info("client started")

	select {
	case <-ctx.Done():
		if err != nil {
			return err
		}

		return fmt.Errorf("unexpected context cancellation")
	}
}

func initializeMetrics(
	ctx context.Context,
	config *config.Config,
	netProvider net.Provider,
	stakeMonitor chain.StakeMonitor,
	address string,
	clientHandle *client.Handle,
) {
	registry, isConfigured := coreMetrics.Initialize(
		config.Metrics.Port,
	)
	if !isConfigured {
		logger.Infof("metrics are not configured")
		return
	}

	logger.Infof(
		"enabled metrics on port [%v]",
		config.Metrics.Port,
	)

	coreMetrics.ObserveConnectedPeersCount(
		ctx,
		registry,
		netProvider,
		time.Duration(config.Metrics.NetworkMetricsTick)*time.Second,
	)

	coreMetrics.ObserveConnectedBootstrapCount(
		ctx,
		registry,
		netProvider,
		config.LibP2P.Peers,
		time.Duration(config.Metrics.NetworkMetricsTick)*time.Second,
	)

	// TODO: make this metric chain-agnostic
	coreMetrics.ObserveEthConnectivity(
		ctx,
		registry,
		stakeMonitor,
		address,
		time.Duration(config.Metrics.EthereumMetricsTick)*time.Second,
	)

	metrics.ObserveTSSPreParamsPoolSize(
		ctx,
		registry,
		clientHandle,
		time.Duration(config.Metrics.ClientMetricsTick)*time.Second,
	)
}

func initializeDiagnostics(
	config *config.Config,
	netProvider net.Provider,
) {
	registry, isConfigured := diagnostics.Initialize(
		config.Diagnostics.Port,
	)
	if !isConfigured {
		logger.Infof("diagnostics are not configured")
		return
	}

	logger.Infof(
		"enabled diagnostics on port [%v]",
		config.Diagnostics.Port,
	)

	diagnostics.RegisterConnectedPeersSource(registry, netProvider)
	diagnostics.RegisterClientInfoSource(registry, netProvider)
}
