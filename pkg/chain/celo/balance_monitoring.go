//+build celo

package celo

import (
	"context"
	"math/big"
	"time"

	"github.com/keep-network/keep-common/pkg/chain/celo"
	"github.com/keep-network/keep-common/pkg/chain/celo/celoutil"
)

// Values related with balance monitoring.
//
// defaultBalanceAlertThreshold determines the alert threshold below which
// the alert should be triggered.
var defaultBalanceAlertThreshold = celo.WrapWei(
	big.NewInt(500000000000000000),
)

// defaultBalanceMonitoringTick determines how often the monitoring
// check should be triggered.
const defaultBalanceMonitoringTick = 10 * time.Minute

func (cc *celoChain) initializeBalanceMonitoring(
	ctx context.Context,
) {
	balanceMonitor, err := cc.balanceMonitor()
	if err != nil {
		logger.Errorf("error obtaining balance monitor handle [%v]", err)
		return
	}

	alertThreshold := defaultBalanceAlertThreshold
	if value := cc.config.BalanceAlertThreshold; value != nil {
		alertThreshold = value
	}

	balanceMonitor.Observe(
		ctx,
		fromExternalAddress(ExternalAddress(cc.Address())),
		alertThreshold,
		defaultBalanceMonitoringTick,
	)

	logger.Infof(
		"started balance monitoring for address [%v] "+
			"with the alert threshold set to [%v]",
		cc.Address().Hex(),
		alertThreshold,
	)
}

// BalanceMonitor returns a balance monitor.
func (cc *celoChain) balanceMonitor() (*celoutil.BalanceMonitor, error) {
	weiBalanceOf := func(address InternalAddress) (*celo.Wei, error) {
		return cc.WeiBalanceOf(toExternalAddress(address))
	}

	return celoutil.NewBalanceMonitor(weiBalanceOf), nil
}
