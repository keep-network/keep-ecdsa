package tss

import (
	configtime "github.com/keep-network/keep-ecdsa/internal/config/time"
)

// Config contains configuration for tss protocol execution.
type Config struct {
	// Timeout for pre-parameters generation in tss-lib.
	PreParamsGenerationTimeout configtime.Duration
}
