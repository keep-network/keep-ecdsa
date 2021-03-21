//+build !celo

package ethereum

import (
	"context"
	"math/big"
	"time"

	"github.com/keep-network/keep-common/pkg/chain/ethereum"
)

// Values related with balance monitoring.
//
// defaultBalanceAlertThreshold determines the alert threshold below which
// the alert should be triggered.
var defaultBalanceAlertThreshold = ethereum.WrapWei(
	big.NewInt(500000000000000000),
)

// defaultBalanceMonitoringTick determines how often the monitoring
// check should be triggered.
const defaultBalanceMonitoringTick = 10 * time.Minute

func (ec *ethereumChain) initializeBalanceMonitoring(
	ctx context.Context,
) {
	balanceMonitor, err := ec.BalanceMonitor()
	if err != nil {
		logger.Errorf("error obtaining balance monitor handle [%v]", err)
		return
	}

	alertThreshold := defaultBalanceAlertThreshold
	if value := ec.config.BalanceAlertThreshold; value != nil {
		alertThreshold = value
	}

	balanceMonitor.Observe(
		ctx,
		ec.operatorAddress(),
		alertThreshold,
		defaultBalanceMonitoringTick,
	)

	logger.Infof(
		"started balance monitoring for address [%v] "+
			"with the alert threshold set to [%v]",
		ec.operatorAddress().Hex(),
		alertThreshold,
	)
}
