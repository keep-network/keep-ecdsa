package client

import (
	"time"

	configtime "github.com/keep-network/keep-ecdsa/config/time"
)

const (
	// The default look-back period to check if existing, active keeps are awaiting
	// signer generation. When the client starts, it goes through all keeps
	// registered on-chain to check whether it's a member of one of them and to
	// generate the signing key if needed. The client does not check keeps older
	// than the awaiting key generation lookback value allows to minimize the
	// number of calls to the chain calls.
	defaultAwaitingKeyGenerationLookback = 24 * time.Hour

	// The default value of a timeout for a keep key generation.
	defaultKeyGenerationTimeout = 3 * time.Hour

	// The default value of a timeout for a signature calculation.
	defaultSigningTimeout = 2 * time.Hour
)

// Config contains configuration for tss protocol execution.
type Config struct {
	// Defines the look-back period to check if existing, active keeps are awaiting
	// signer generation on the client start. The client does not check keeps older
	// than the look-back value.
	AwaitingKeyGenerationLookback configtime.Duration

	// Timeout for key generation and signature calculation.
	KeyGenerationTimeout configtime.Duration
	SigningTimeout       configtime.Duration
}

// GetAwaitingKeyGenerationLookback returns a look-back period to check if
// existing, active keeps are awaiting signer generation. If a value is not set
// it returns a default value.
func (c *Config) GetAwaitingKeyGenerationLookback() time.Duration {
	lookbackPeriod := c.AwaitingKeyGenerationLookback.ToDuration()
	if lookbackPeriod == 0 {
		lookbackPeriod = defaultAwaitingKeyGenerationLookback
	}

	return lookbackPeriod
}

// GetKeyGenerationTimeout returns key generation timeout. If a value is not set
// it returns a default value.
func (c *Config) GetKeyGenerationTimeout() time.Duration {
	timeout := c.KeyGenerationTimeout.ToDuration()
	if timeout == 0 {
		timeout = defaultKeyGenerationTimeout
	}

	return timeout
}

// GetSigningTimeout returns signature calculation timeout. If a value is not set
// it returns a default value.
func (c *Config) GetSigningTimeout() time.Duration {
	timeout := c.SigningTimeout.ToDuration()
	if timeout == 0 {
		timeout = defaultSigningTimeout
	}

	return timeout
}
