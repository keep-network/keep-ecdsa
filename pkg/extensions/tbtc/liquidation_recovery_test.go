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

func Test_PublicKeyToP2WPKHScriptCode_Works(t *testing.T) {
	curve := elliptic.P224()
	privateKey, _ := cecdsa.GenerateKey(curve, rand.Reader)
	scriptCodeBytes, err := PublicKeyToP2WPKHScriptCode(&privateKey.PublicKey, &chaincfg.TestNet3Params)

	if err != nil {
		t.Errorf("%v", err)
	}

	if len(scriptCodeBytes) != 26 {
		t.Errorf("The script code must be exactly 26 bytes long. Instead, it was %v", len(scriptCodeBytes))
	}
}

func Test_ConstructUnsignedTransaction(t *testing.T) {
	recoveryAddresses := []string{
		"bcrt1q5sz7jly79m76a5e8py6kv402q07p725vm4s0zl",
		"bcrt1qlxt5a04pefwkl90mna2sn79nu7asq3excx60h0",
		"bcrt1qjhpgmmhaxfwj6t7zf3dvs2fhdhx02g8qn3xwsf",
	}
	transactionValue := int64(100000000)
	msgTx, err := ConstructUnsignedTransaction(
		"0b99dea9655f219991001e9296cfe2103dd918a21ef477a14121d1a0ba9491f1",
		uint32(0),
		transactionValue,
		int64(700),
		recoveryAddresses,
		&chaincfg.TestNet3Params,
	)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(msgTx.TxIn) != 1 {
		t.Errorf("There should be 1 input transaction. Got %v instead.", len(msgTx.TxIn))
	}
	if len(msgTx.TxOut) != len(recoveryAddresses) {
		t.Errorf(
			"There should be an output transaction for each recovery address. Got %v but expected %v",
			len(msgTx.TxOut),
			len(recoveryAddresses),
		)
	}
	share := msgTx.TxOut[0].Value
	for _, txOut := range msgTx.TxOut {
		feelessShare := transactionValue / int64(len(recoveryAddresses))
		if txOut.Value >= feelessShare {
			t.Errorf(
				"Each output transaction should not be more than a signer's feeless share of deposit value. %v >= %v",
				txOut.Value,
				feelessShare,
			)
		}
		if txOut.Value != share {
			t.Errorf(
				"Each output transaction should represent an equal share. Got %v but expected %v",
				txOut.Value,
				share,
			)
		}
	}
}

func Test_BuildSignedTransactionHexString(t *testing.T) {
	tx := wire.NewMsgTx(wire.TxVersion)
	previousOutputTransactionHash, err := chainhash.NewHashFromStr(
		"0b99dea9655f219991001e9296cfe2103dd918a21ef477a14121d1a0ba9491f1",
	)
	if err != nil {
		t.Errorf("Something went wrong generating the previousOutputTransactionHash: %v", err)
	}
	tx.AddTxIn(wire.NewTxIn(
		wire.NewOutPoint(previousOutputTransactionHash, 0),
		nil,
		nil,
	))

	for _, txValue := range []int{100, 100, 100} {
		tx.AddTxOut(wire.NewTxOut(
			int64(txValue),
			nil,
		))
	}
	curve := elliptic.P224()
	privateKey, _ := cecdsa.GenerateKey(curve, rand.Reader)
	signedHexString, err := BuildSignedTransactionHexString(
		tx,
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

	decodedTx := wire.NewMsgTx(wire.TxVersion)
	b, _ := hex.DecodeString(signedHexString)
	decodedTx.BtcDecode(bytes.NewReader(b), 1, wire.WitnessEncoding)

	if len(decodedTx.TxIn) != len(tx.TxIn) {
		t.Errorf(
			"The original and signed transactions must have the same number of input transactions.\n"+
				"Actual:   %v\nExpected: %v",
			len(decodedTx.TxIn),
			len(tx.TxIn),
		)
	}
	for i, decodedTxIn := range decodedTx.TxIn {
		originalTxIn := tx.TxIn[i]
		if decodedTxIn.PreviousOutPoint.Hash != originalTxIn.PreviousOutPoint.Hash {
			t.Errorf(
				"TxIn hashes don't match.\nActual:   %v\nExpected: %v",
				decodedTxIn.PreviousOutPoint.Hash,
				originalTxIn.PreviousOutPoint.Hash,
			)
		}
		if decodedTxIn.Witness == nil {
			t.Errorf("TxIn does not have a witness.")
		}
	}
	if len(decodedTx.TxOut) != len(tx.TxOut) {
		t.Errorf(
			"The original and signed transactions must have the same number of output transactions.\n"+
				"Actual:   %v\nExpected: %v",
			len(decodedTx.TxOut),
			len(tx.TxOut),
		)
	}
	for i, decodedTxOut := range decodedTx.TxOut {
		originalTxOut := tx.TxOut[i]
		if decodedTxOut.Value != originalTxOut.Value {
			t.Errorf(
				"TxOut values don't match.\nActual:   %v\nExpected: %v",
				decodedTxOut.Value,
				originalTxOut.Value,
			)
		}
	}
}
