package tss

import (
	"time"

	configtime "github.com/keep-network/keep-ecdsa/internal/config/time"
)

const (
	defaultPreParamsGenerationTimeout = 2 * time.Minute
)

// Config contains configuration for tss protocol execution.
type Config struct {
	// Timeout for pre-parameters generation in tss-lib.
	PreParamsGenerationTimeout configtime.Duration
}

// GetKeyGenerationTimeout return key generation timeout as `time.Duration`.
func (c *Config) GetPreParamsGenerationTimeout() time.Duration {
	timeout := c.PreParamsGenerationTimeout.ToDuration()
	if timeout == 0 {
		timeout = defaultPreParamsGenerationTimeout
	}

	return timeout
}
