// Package btc contains interface for interaction with a bitcoin blockchain.
package btc

// Interface is an interface that provides ability to interact with a blockchain.
type Interface interface {
	// PublishTransaction publishes a transaction to a chain. It requires raw
	// transaction to be provided in a format specific to a chain. It returns
	// an unique identifier of the transaction.
	PublishTransaction(rawTx string) (string, error)

	// GetCurrentTargetBits gets difficulty target bits from latest block published
	// to the blockchain.
	GetCurrentTargetBits() (int, error)
}
