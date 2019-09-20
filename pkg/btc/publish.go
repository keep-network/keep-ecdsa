package btc

import (
	"encoding/hex"
	"fmt"

	"github.com/keep-network/keep-tecdsa/pkg/chain/btc"
)

// Publish submits a transaction to the block chain. It requires chain implementation
// and a raw transaction to be provided. It returns unique transaction identifier.
func Publish(chain btc.Interface, rawTx []byte) (string, error) {
	result, err := chain.PublishTransaction(hex.EncodeToString(rawTx))
	if err != nil {
		return "", fmt.Errorf("failed to publish transaction: [%v]", err)
	}
	return result, nil
}
