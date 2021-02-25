//+build celo

package cmd

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/celo-org/celo-blockchain/accounts/keystore"

	commoncelo "github.com/keep-network/keep-common/pkg/chain/celo"
	"github.com/keep-network/keep-common/pkg/chain/celo/celoutil"
	"github.com/keep-network/keep-ecdsa/config"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/celo"
)

// Values related with balance monitoring.
//
// defaultBalanceAlertThreshold determines the alert threshold below which
// the alert should be triggered.
var defaultBalanceAlertThreshold = commoncelo.WrapWei(
	big.NewInt(500000000000000000),
)

// defaultBalanceMonitoringTick determines how often the monitoring
// check should be triggered.
const defaultBalanceMonitoringTick = 10 * time.Minute

func connectChain(
	ctx context.Context,
	config *config.Config,
) (eth.Handle, *operatorKeys, error) {
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
	initializeBalanceMonitoring(ctx, config, celoChain, celoKey)

	operatorKeys := &operatorKeys{
		public:  &celoKey.PrivateKey.PublicKey,
		private: celoKey.PrivateKey,
	}

	return celoChain, operatorKeys, nil
}

func initializeBalanceMonitoring(
	ctx context.Context,
	config *config.Config,
	celoChain *celo.Chain,
	celokey *keystore.Key,
) {
	balanceMonitor, err := celoChain.BalanceMonitor()
	if err != nil {
		logger.Errorf("error obtaining balance monitor handle [%v]", err)
		return
	}

	alertThreshold := defaultBalanceAlertThreshold
	if value := config.Celo.BalanceAlertThreshold; value != nil {
		alertThreshold = value
	}

	balanceMonitor.Observe(
		ctx,
		celokey.Address,
		alertThreshold,
		defaultBalanceMonitoringTick,
	)

	logger.Infof(
		"started balance monitoring for address [%v] "+
			"with the alert threshold set to [%v]",
		celokey.Address.Hex(),
		alertThreshold,
	)
}

func extractKeyFilePassword(config *config.Config) string {
	return config.Celo.Account.KeyFilePassword
}
