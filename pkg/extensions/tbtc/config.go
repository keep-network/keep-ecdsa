package tbtc

import (
	"time"

	"github.com/keep-network/keep-ecdsa/pkg/chain/bitcoin"

	configtime "github.com/keep-network/keep-ecdsa/config/time"
)

const (
	// The default value of a timeout for liquidation recovery.
	defaultLiquidationRecoveryTimeout = 48 * time.Hour
)

// Config stores configuration of application extensions responsible for
// executing signer actions specific for TBTC application.
type Config struct {
	TBTCSystem                 string
	Bitcoin                    bitcoin.Config
	LiquidationRecoveryTimeout configtime.Duration
}

// GetLiquidationRecoveryTimeout returns the liquidation recovery timeout. If a
// value is not set it returns a default value.
func (c *Config) GetLiquidationRecoveryTimeout() time.Duration {
	timeout := c.LiquidationRecoveryTimeout.ToDuration()
	if timeout == 0 {
		timeout = defaultLiquidationRecoveryTimeout
	}

	return timeout
}
