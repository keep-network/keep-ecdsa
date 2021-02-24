//+build !celo

package cmd

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-ecdsa/config"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/ethereum"
	"github.com/keep-network/keep-ecdsa/pkg/extensions/tbtc"
)

// Values related with balance monitoring.
//
// defaultBalanceAlertThreshold determines the alert threshold below which
// the alert should be triggered.
var defaultBalanceAlertThreshold = big.NewInt(500000000000000000) // 0.5 ETH

// defaultBalanceMonitoringTick determines how often the monitoring
// check should be triggered.
const defaultBalanceMonitoringTick = 10 * time.Minute

func connectChain(
	ctx context.Context,
	config *config.Config,
) (eth.Handle, *operatorKeys, error) {
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

	initializeExtensions(ctx, config.Extensions, ethereumChain)
	initializeBalanceMonitoring(ctx, config, ethereumChain, ethereumKey)

	operatorKeys := &operatorKeys{
		public:  &ethereumKey.PrivateKey.PublicKey,
		private: ethereumKey.PrivateKey,
	}

	return ethereumChain, operatorKeys, nil
}

func initializeExtensions(
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

func initializeBalanceMonitoring(
	ctx context.Context,
	config *config.Config,
	ethereumChain *ethereum.EthereumChain,
	ethereumKey *keystore.Key,
) {
	balanceMonitor, err := ethereumChain.BalanceMonitor()
	if err != nil {
		logger.Errorf("error obtaining balance monitor handle [%v]", err)
		return
	}

	alertThreshold := defaultBalanceAlertThreshold
	if value := config.Ethereum.BalanceAlertThreshold; value != nil {
		alertThreshold = value.Int
	}

	address := ethereumKey.Address.Hex()

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

func extractKeyFilePassword(config *config.Config) string {
	return config.Ethereum.Account.KeyFilePassword
}
