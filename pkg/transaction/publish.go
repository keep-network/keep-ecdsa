// Package transaction handles transaction publishing to a block chain.
package transaction

import (
	"fmt"

	"github.com/keep-network/keep-tecdsa/pkg/chain"
)

// Publish submits a transaction to the block chain. It requires chain implementation
// and a raw transaction to be provided. It returns unique transaction identifier.
func Publish(chain chain.Interface, rawTx string) (string, error) {
	result, err := chain.SubmitTransaction(rawTx)
	if err != nil {
		return "", fmt.Errorf("transaction broadcast failed [%s]", err)
	}
	return result, nil
}
