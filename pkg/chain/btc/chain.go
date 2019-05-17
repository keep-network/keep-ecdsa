// Package btc contains interface for interaction with a blockchain.
package btc

// ChainInterface is an interface that provides ability to interact with a blockchain.
type ChainInterface interface {
	// PublishTransaction publishes a transaction to a chain. It requires raw
	// transaction to be provided in a format specific to a chain. It returns
	// an unique identifier of the transaction.
	PublishTransaction(rawTx string) (string, error)
}
