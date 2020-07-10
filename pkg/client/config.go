package client

import (
	configtime "github.com/keep-network/keep-ecdsa/internal/config/time"
)

// Config contains configuration for tss protocol execution.
type Config struct {
	// Period to check keeps for awaiting key generation on client start.
	AwaitingKeyGenerationLookback configtime.Duration

	// Timeout for key generation and signature calculation.
	KeyGenerationTimeout configtime.Duration
	SigningTimeout       configtime.Duration
}
