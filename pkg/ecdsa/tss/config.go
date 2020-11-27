package tss

import (
	"time"

	configtime "github.com/keep-network/keep-ecdsa/config/time"
)

const (
	defaultPreParamsGenerationTimeout = 2 * time.Minute
	defaultPreParamsTargetPoolSize    = 20
)

// Config contains configuration for tss protocol execution.
type Config struct {
	// Timeout for pre-parameters generation in tss-lib.
	PreParamsGenerationTimeout configtime.Duration

	// Target size of the TSS pre params pool.
	PreParamsTargetPoolSize int
}

// GetPreParamsGenerationTimeout returns pre-parameters generation timeout. If
// a value is not set it returns a default value.
func (c *Config) GetPreParamsGenerationTimeout() time.Duration {
	timeout := c.PreParamsGenerationTimeout.ToDuration()
	if timeout == 0 {
		timeout = defaultPreParamsGenerationTimeout
	}

	return timeout
}

// GetPreParamsTargetPoolSize returns the pre-parameters target pool size. If
// a value is not set it returns a default value.
func (c *Config) GetPreParamsTargetPoolSize() int {
	poolSize := c.PreParamsTargetPoolSize
	if poolSize == 0 {
		poolSize = defaultPreParamsTargetPoolSize
	}

	return poolSize
}
