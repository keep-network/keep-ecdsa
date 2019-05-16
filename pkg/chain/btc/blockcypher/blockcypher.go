// Package blockcypher contains implementation of the chain interface communicating
// with [Block Cypher API](https://www.blockcypher.com/dev/bitcoin/).
package blockcypher

import (
	"github.com/blockcypher/gobcy"
	"github.com/keep-network/keep-tecdsa/pkg/btc/chain"
)

type blockcypher struct {
	api gobcy.API
}

// Config contains configuration for Block Cypher API.
type Config struct {
	// Token is Block Cypher's user token required for access to POST and DELETE
	// calls on the API.
	Token string
	Coin  string // Options: "btc", "bcy", "ltc", "doge"
	Chain string // Options: "main", "test3", "test"
}

// PublishTransaction sends a raw transaction provided as a hexadecimal string
// to Block Cypher's API. It returns a transaction hash as a hexadecimal string.
func (bc *blockcypher) PublishTransaction(rawTx string) (string, error) {
	tx, err := bc.api.PushTX(rawTx)
	if err != nil {
		return "", err
	}

	return tx.Trans.Hash, nil
}

// Connect performs initialization for communication with Block Cypher based on
// provided config.
func Connect(config *Config) (chain.Interface, error) {
	blockCypherAPI := gobcy.API{
		Token: config.Token,
		Coin:  config.Coin,
		Chain: config.Chain,
	}

	return &blockcypher{api: blockCypherAPI}, nil
}
