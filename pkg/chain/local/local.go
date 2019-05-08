// Package local contains local stub implementation of the chain interface.
// This implementation is for development and testing purposes only.
package local

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/keep-network/keep-tecdsa/pkg/chain"
	"github.com/keep-network/keep-tecdsa/pkg/utils"
)

const initialTransaction = "0100000001f24d19b6980927dbe47c30fd13b1cc12e56a11cc019efed67a1b4d3937b74bab010000006a47304402201711a033c1b829719716c81419294214a7fce0f0f1f9f51b6821ca3a5beebbdd022059b7bdd0bf1fe08aa4b4654360732d2a1f97c602b2e198a41e7bc53d81376c9a0121028896955d043b5a43957b21901f2cce9f0bfb484531b03ad6cd3153e45e73ee2effffffff022823000000000000160014d849b1e1cede2ac7d7188cf8700e97d6975c91c4b2f9fd00000000001976a914d849b1e1cede2ac7d7188cf8700e97d6975c91c488ac00000000"

type localChain struct {
	transactions map[string]*wire.MsgTx
}

// PublishTransaction performs validation on a transaction encoded to hexadecimal
// bitcoin format and stores it on local chain transactions map. It returns
// transaction hash as an unique identifier of the transaction.
func (l *localChain) PublishTransaction(rawTx string) (string, error) {
	if rawTx == "" {
		return "", fmt.Errorf("empty transaction provided")
	}

	rawTxBytes, err := hex.DecodeString(rawTx)
	if err != nil {
		return "", fmt.Errorf("cannot decode string [%s]", err)
	}
	msgTx, err := utils.DeserializeTransaction(rawTxBytes)
	if err != nil {
		return "", fmt.Errorf("cannot deserialize transaction [%s]", err)
	}

	// Check for duplicate.
	txHash := msgTx.TxHash().String()
	if _, ok := l.transactions[txHash]; ok {
		return "", fmt.Errorf(
			"transaction already registered on chain for hash [%s]",
			txHash,
		)
	}

	// Validate transaction according to bitcoin rules.
	previousTxHash := msgTx.TxIn[0].PreviousOutPoint.Hash.String()
	previousTx, ok := l.transactions[previousTxHash]
	if !ok {
		return "", fmt.Errorf(
			"previous transaction not found for hash [%s]",
			previousTxHash,
		)
	}

	previousOutputIndex := msgTx.TxIn[0].PreviousOutPoint.Index
	previousOutputScript := hex.EncodeToString(previousTx.TxOut[previousOutputIndex].PkScript)
	previousOutputAmout := previousTx.TxOut[previousOutputIndex].Value

	if err := ValidateTransaction(
		previousOutputScript,
		previousOutputAmout,
		msgTx,
	); err != nil {
		return "", err
	}

	// Register transaction on chain.
	l.transactions[txHash] = msgTx

	return txHash, nil
}

// Connect returns a stub implementation of the chain interface.
func Connect() chain.Interface {
	initialTx := initialTx()
	return &localChain{
		transactions: map[string]*wire.MsgTx{
			initialTx.TxHash().String(): initialTx,
		},
	}
}

// ValidateTransaction verifies if a transaction fulfills all the requirements
// specified by bitcoin's StandardVerifyFlags.
func ValidateTransaction(
	previousOutputScript string,
	previousOutputAmout int64,
	transaction *wire.MsgTx) error {
	if len(transaction.TxIn) != 1 {
		return fmt.Errorf("only transactions with one input are supported")
	}
	inputIndex := 0

	subscript, err := hex.DecodeString(previousOutputScript)
	if err != nil {
		return fmt.Errorf("source subscript decoding failed [%s]", err)
	}

	validationEngine, err := txscript.NewEngine(
		subscript,
		transaction,
		inputIndex,
		txscript.StandardVerifyFlags,
		nil,
		nil,
		previousOutputAmout,
	)
	if err != nil {
		return fmt.Errorf(
			"cannot create validation engine [%s]",
			err,
		)
	}

	if err := validationEngine.Execute(); err != nil {
		return fmt.Errorf(
			"transaction validation failed [%s]",
			err,
		)
	}

	return nil
}

func initialTx() *wire.MsgTx {
	txString := initialTransaction
	txBytes, _ := hex.DecodeString(txString)
	tx, _ := utils.DeserializeTransaction(txBytes)
	return tx
}
