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

func (c *Chain) initializeBalanceMonitoring(
	ctx context.Context,
) {
	balanceMonitor, err := c.balanceMonitor()
	if err != nil {
		logger.Errorf("error obtaining balance monitor handle [%v]", err)
		return
	}

	alertThreshold := defaultBalanceAlertThreshold
	if value := c.config.BalanceAlertThreshold; value != nil {
		alertThreshold = value
	}

	balanceMonitor.Observe(
		ctx,
		fromExternalAddress(ExternalAddress(c.Address())),
		alertThreshold,
		defaultBalanceMonitoringTick,
	)

	logger.Infof(
		"started balance monitoring for address [%v] "+
			"with the alert threshold set to [%v]",
		c.Address().Hex(),
		alertThreshold,
	)
}

// BalanceMonitor returns a balance monitor.
func (c *Chain) balanceMonitor() (*celoutil.BalanceMonitor, error) {
	weiBalanceOf := func(address InternalAddress) (*celo.Wei, error) {
		return c.WeiBalanceOf(toExternalAddress(address))
	}

	return celoutil.NewBalanceMonitor(weiBalanceOf), nil
}
