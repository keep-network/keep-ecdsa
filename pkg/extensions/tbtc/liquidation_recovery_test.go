package tbtc

import (
	"bytes"
	cecdsa "crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
)

func TestPublicKeyToP2WPKHScriptCode(t *testing.T) {
	curve := elliptic.P224()
	privateKey, _ := cecdsa.GenerateKey(curve, rand.Reader)
	scriptCodeBytes, err := PublicKeyToP2WPKHScriptCode(&privateKey.PublicKey, &chaincfg.TestNet3Params)

	if err != nil {
		t.Error(err)
	}

	if len(scriptCodeBytes) != 26 {
		t.Errorf("The script code must be exactly 26 bytes long. Instead, it was %v", len(scriptCodeBytes))
	}
}

func TestConstructUnsignedTransaction(t *testing.T) {
	recipientAddresses := []string{
		"bcrt1q5sz7jly79m76a5e8py6kv402q07p725vm4s0zl",
		"bcrt1qlxt5a04pefwkl90mna2sn79nu7asq3excx60h0",
		"bcrt1qjhpgmmhaxfwj6t7zf3dvs2fhdhx02g8qn3xwsf",
	}

	previousOutputValue := int64(100000000)
	expectedShare := int64(45)

	messageTransaction, err := ConstructUnsignedTransaction(
		"0b99dea9655f219991001e9296cfe2103dd918a21ef477a14121d1a0ba9491f1",
		uint32(0),
		previousOutputValue,
		int64(700),
		recipientAddresses,
		&chaincfg.TestNet3Params,
	)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(messageTransaction.TxIn) != 1 {
		t.Errorf("There should be 1 input transaction. Got %v instead.", len(messageTransaction.TxIn))
	}
	if len(messageTransaction.TxOut) != len(recipientAddresses) {
		t.Errorf(
			"there should be an output transaction for each recovery address\nexpected: %v\nactual:   %v",
			len(recipientAddresses),
			len(messageTransaction.TxOut),
		)
	}
	for _, transactionOut := range messageTransaction.TxOut {
		if transactionOut.Value != expectedShare {
			t.Errorf(
				"incorrect transaction output value\nexpected: %v\nactual:   %v",
				expectedShare,
				transactionOut.Value,
			)
		}
	}
}

func Test_BuildSignedTransactionHexString(t *testing.T) {
	unsignedTransaction := wire.NewMsgTx(wire.TxVersion)
	previousOutputTransactionHash, err := chainhash.NewHashFromStr(
		"0b99dea9655f219991001e9296cfe2103dd918a21ef477a14121d1a0ba9491f1",
	)
	if err != nil {
		t.Errorf("Something went wrong generating the previousOutputTransactionHash: %v", err)
	}
	unsignedTransaction.AddTxIn(wire.NewTxIn(
		wire.NewOutPoint(previousOutputTransactionHash, 0),
		nil,
		nil,
	))

	for _, txValue := range []int{100, 100, 100} {
		unsignedTransaction.AddTxOut(wire.NewTxOut(
			int64(txValue),
			nil,
		))
	}
	curve := elliptic.P224()
	privateKey, _ := cecdsa.GenerateKey(curve, rand.Reader)
	signedTransactionHex, err := BuildSignedTransactionHexString(
		unsignedTransaction,
		&ecdsa.Signature{
			R:          big.NewInt(int64(3)),
			S:          big.NewInt(int64(7)),
			RecoveryID: 1,
		},
		&privateKey.PublicKey,
	)
	if err != nil {
		t.Errorf("Something went wrong building the signed transaction string: %v", err)
	}

	signedTransaction := wire.NewMsgTx(wire.TxVersion)
	signedTransactionBytes, _ := hex.DecodeString(signedTransactionHex)
	signedTransaction.BtcDecode(bytes.NewReader(signedTransactionBytes), 1, wire.WitnessEncoding)

	if len(signedTransaction.TxIn) != len(unsignedTransaction.TxIn) {
		t.Errorf(
			"the original and signed transactions must have the same number of input transactions\n"+
				"expected: %v\nactual:   %v",
			len(unsignedTransaction.TxIn),
			len(signedTransaction.TxIn),
		)
	}
	for i, signedTransactionIn := range signedTransaction.TxIn {
		originalTransactionIn := unsignedTransaction.TxIn[i]
		if signedTransactionIn.PreviousOutPoint.Hash != originalTransactionIn.PreviousOutPoint.Hash {
			t.Errorf(
				"TxIn hashes don't match\nexpected: %v\nactual:   %v",
				originalTransactionIn.PreviousOutPoint.Hash,
				signedTransactionIn.PreviousOutPoint.Hash,
			)
		}
		if signedTransactionIn.Witness == nil {
			t.Errorf("TxIn does not have a witness.")
		}
	}
	if len(signedTransaction.TxOut) != len(unsignedTransaction.TxOut) {
		t.Errorf(
			"the original and signed transactions must have the same number of output transactions\n"+
				"expected: %v\nactual:   %v",
			len(unsignedTransaction.TxOut),
			len(signedTransaction.TxOut),
		)
	}
	for i, signedTransactionOut := range signedTransaction.TxOut {
		originalTransactionOut := unsignedTransaction.TxOut[i]
		if signedTransactionOut.Value != originalTransactionOut.Value {
			t.Errorf(
				"TxOut values don't match.\nexpected: %v\nactual:   %v",
				originalTransactionOut.Value,
				signedTransactionOut.Value,
			)
		}
	}
}
