// Package chain contains interface for interaction with a blockchain.
package chain

// Interface is an interface that provides ability to interact with a blockchain.
type Interface interface {
	// SubmitTransaction submits a transaction to a chain. It requires raw
	// transaction to be provided in a format specific to a chain. It returns
	// an unique identifier of the transaction.
	SubmitTransaction(rawTx string) (string, error)
}
