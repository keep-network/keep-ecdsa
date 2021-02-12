package cmd

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/keep-network/keep-common/pkg/chain/celo/celoutil"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/celo"

	"github.com/keep-network/keep-ecdsa/pkg/metrics"

	"github.com/keep-network/keep-core/pkg/diagnostics"

	"github.com/keep-network/keep-core/pkg/chain"
	coreMetrics "github.com/keep-network/keep-core/pkg/metrics"
	"github.com/keep-network/keep-core/pkg/net"

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/persistence"

	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-core/pkg/net/libp2p"
	"github.com/keep-network/keep-core/pkg/net/retransmission"
	"github.com/keep-network/keep-core/pkg/operator"

	"github.com/keep-network/keep-ecdsa/config"
	"github.com/keep-network/keep-ecdsa/pkg/chain/ethereum"
	"github.com/keep-network/keep-ecdsa/pkg/client"
	"github.com/keep-network/keep-ecdsa/pkg/extensions/tbtc"
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

// Values related with balance monitoring.
// defaultBalanceAlertThreshold determines the alert threshold below which
// the alert should be triggered.
var defaultBalanceAlertThreshold = big.NewInt(500000000000000000) // 0.5 ether
// defaultBalanceMonitoringTick determines how often the monitoring
// check should be triggered.
const defaultBalanceMonitoringTick = 10 * time.Minute

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

	chainHandle, privateKey, err := connectChain(ctx, config)
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

	operatorPrivateKey, operatorPublicKey := privateKeyToOperatorKeys(
		privateKey,
	)

	networkPrivateKey, _ := key.OperatorKeyToNetworkKey(
		operatorPrivateKey, operatorPublicKey,
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
		config.Ethereum.Account.KeyFilePassword,
	)

	sanctionedApplications, err := config.SanctionedApplications.Addresses()
	if err != nil {
		return fmt.Errorf("failed to get sanctioned applications addresses: [%v]", err)
	}

	clientHandle := client.Initialize(
		ctx,
		operatorPublicKey,
		chainHandle,
		networkProvider,
		persistence,
		sanctionedApplications,
		&config.Client,
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

func connectChain(
	ctx context.Context,
	config *config.Config,
) (eth.Handle, *ecdsa.PrivateKey, error) {
	if config.Ethereum.Enabled {
		return connectEthereumChain(ctx, config)
	}

	if config.Celo.Enabled {
		return connectCeloChain(ctx, config)
	}

	return nil, nil, fmt.Errorf(
		"config doesn't contain any enabled chain",
	)
}

func connectEthereumChain(
	ctx context.Context,
	config *config.Config,
) (eth.Handle, *ecdsa.PrivateKey, error) {
	ethereumKey, err := ethutil.DecryptKeyFile(
		config.Ethereum.Account.KeyFile,
		config.Ethereum.Account.KeyFilePassword,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to read key file [%s]: [%v]",
			config.Ethereum.Account.KeyFile,
			err,
		)
	}

	ethereumChain, err := ethereum.Connect(ethereumKey, &config.Ethereum)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to connect to ethereum node: [%v]",
			err,
		)
	}

	initializeEthereumExtensions(ctx, config.Extensions, ethereumChain)
	initializeBalanceMonitoring(
		ctx,
		ethereumChain,
		&config.Ethereum,
		ethereumKey.Address.Hex(),
	)

	return ethereumChain, ethereumKey.PrivateKey, nil
}

func initializeEthereumExtensions(
	ctx context.Context,
	config config.Extensions,
	ethereumChain *ethereum.EthereumChain,
) {
	if len(config.TBTC.TBTCSystem) > 0 {
		tbtcEthereumChain, err := ethereum.WithTBTCExtension(
			ethereumChain,
			config.TBTC.TBTCSystem,
		)
		if err != nil {
			logger.Errorf(
				"could not initialize tbtc chain extension: [%v]",
				err,
			)
			return
		}

		tbtc.Initialize(ctx, tbtcEthereumChain)
	}
}

func connectCeloChain(
	ctx context.Context,
	config *config.Config,
) (eth.Handle, *ecdsa.PrivateKey, error) {
	celoKey, err := celoutil.DecryptKeyFile(
		config.Celo.Account.KeyFile,
		config.Celo.Account.KeyFilePassword,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to read key file [%s]: [%v]",
			config.Celo.Account.KeyFile,
			err,
		)
	}

	celoChain, err := celo.Connect(celoKey, &config.Celo)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to connect to celo node: [%v]",
			err,
		)
	}

	// TODO: initialize Celo extensions
	initializeBalanceMonitoring(
		ctx,
		celoChain,
		&config.Celo,
		celoKey.Address.Hex(),
	)

	return celoChain, celoKey.PrivateKey, nil
}

func initializeMetrics(
	ctx context.Context,
	config *config.Config,
	netProvider net.Provider,
	stakeMonitor chain.StakeMonitor,
	ethereumAddres string,
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

	coreMetrics.ObserveEthConnectivity(
		ctx,
		registry,
		stakeMonitor,
		ethereumAddres,
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

type balanceMonitoringConfig interface {
	BalanceAlertThresholdValue() *big.Int
}

func initializeBalanceMonitoring(
	ctx context.Context,
	chainHandle eth.Handle,
	config balanceMonitoringConfig,
	address string,
) {
	balanceMonitor, err := chainHandle.BalanceMonitor()
	if err != nil {
		logger.Errorf("error obtaining balance monitor handle [%v]", err)
		return
	}

	alertThreshold := defaultBalanceAlertThreshold
	if value := config.BalanceAlertThresholdValue(); value != nil {
		alertThreshold = value
	}

	balanceMonitor.Observe(
		ctx,
		address,
		alertThreshold,
		defaultBalanceMonitoringTick,
	)

	logger.Infof(
		"started balance monitoring for address [%v] "+
			"with the alert threshold set to [%v]",
		address,
		alertThreshold,
	)
}

// TODO: Temporary helper function. Should be removed once the `operator`
//  package allows to convert keys of multiple chains.
func privateKeyToOperatorKeys(
	privateKey *ecdsa.PrivateKey,
) (*operator.PrivateKey, *operator.PublicKey) {
	return privateKey, &privateKey.PublicKey
}
