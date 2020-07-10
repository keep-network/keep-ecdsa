package client

import (
	"time"

	configtime "github.com/keep-network/keep-ecdsa/internal/config/time"
)

const (
	defaultAwaitingKeyGenerationLookback = 24 * time.Hour
	defaultKeyGenerationTimeout          = 3 * time.Hour
	defaultSigningTimeout                = 2 * time.Hour
)

// Config contains configuration for tss protocol execution.
type Config struct {
	// Period to check keeps for awaiting key generation on client start.
	AwaitingKeyGenerationLookback configtime.Duration

	// Timeout for key generation and signature calculation.
	KeyGenerationTimeout configtime.Duration
	SigningTimeout       configtime.Duration
}

// GetAwaitingKeyGenerationLookback return awaiting key generation lookback as
// `time.Duration`.
func (c *Config) GetAwaitingKeyGenerationLookback() time.Duration {
	lookbackPeriod := c.AwaitingKeyGenerationLookback.ToDuration()
	if lookbackPeriod == 0 {
		lookbackPeriod = defaultAwaitingKeyGenerationLookback
	}

	return lookbackPeriod
}

// GetKeyGenerationTimeout return key generation timeout as `time.Duration`.
func (c *Config) GetKeyGenerationTimeout() time.Duration {
	timeout := c.KeyGenerationTimeout.ToDuration()
	if timeout == 0 {
		timeout = defaultKeyGenerationTimeout
	}

	return timeout
}

// GetSigningTimeout return key signing timeout as `time.Duration`.
func (c *Config) GetSigningTimeout() time.Duration {
	timeout := c.SigningTimeout.ToDuration()
	if timeout == 0 {
		timeout = defaultSigningTimeout
	}

	return timeout
}
