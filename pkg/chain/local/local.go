// Package local contains local stub implementation of the chain interface.
// This implementation is for development and testing purposes only.
package local

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/keep-network/keep-tecdsa/pkg/chain"
)

type localChain struct{}

// BroadcastTransaction is a stub implementation which calculates a SHA-256 hash
// for given string and returns it encoded to a hexadecimal string.
func (l *localChain) BroadcastTransaction(rawTx string) (string, error) {
	if rawTx == "" {
		return "", fmt.Errorf("empty transaction provided")
	}
	hash := sha256.Sum256([]byte(rawTx))
	return hex.EncodeToString(hash[:]), nil
}

// Connect returns a stub implementation of the chain interface.
func Connect() chain.Interface {
	return &localChain{}
}
