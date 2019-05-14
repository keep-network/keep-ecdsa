package btc

import (
	"encoding/hex"
	"fmt"

	"github.com/keep-network/keep-tecdsa/pkg/btc/chain"
)

// Publish submits a transaction to the block chain. It requires chain implementation
// and a raw transaction to be provided. It returns unique transaction identifier.
func Publish(chain chain.Interface, rawTx []byte) (string, error) {
	result, err := chain.PublishTransaction(hex.EncodeToString(rawTx))
	if err != nil {
		return "", fmt.Errorf("transaction broadcast failed [%s]", err)
	}
	return result, nil
}
