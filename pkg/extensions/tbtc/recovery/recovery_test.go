package recovery

import (
	"bytes"
	cecdsa "crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/google/go-cmp/cmp"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
	"gotest.tools/v3/assert"
)

func TestPublicKeyToP2WPKHScriptCode(t *testing.T) {
	// Test based on test values from BIP143:
	// https://github.com/bitcoin/bips/blob/master/bip-0143.mediawiki#native-p2wpkh
	serializedPublicKey, _ := hex.DecodeString("025476c2e83188368da1ff3e292e7acafcdb3566bb0ad253f62fc70f07aeee6357")
	expectedScriptCode, _ := hex.DecodeString("1976a9141d0f172a0ecb48aee1be1f2687d2963ae33f71a188ac")

	publicKey, _ := btcec.ParsePubKey(serializedPublicKey, btcec.S256())

	scriptCodeBytes, err := PublicKeyToP2WPKHScriptCode(publicKey.ToECDSA(), &chaincfg.MainNetParams)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bytes.Compare(expectedScriptCode, scriptCodeBytes) != 0 {
		t.Errorf(
			"unexpected script code\nexpected: %x\nactual:   %x",
			expectedScriptCode,
			scriptCodeBytes,
		)
	}
}

func TestConstructUnsignedTransaction(t *testing.T) {
	recipientAddresses := []string{
		"bcrt1q5sz7jly79m76a5e8py6kv402q07p725vm4s0zl",
		"bcrt1qlxt5a04pefwkl90mna2sn79nu7asq3excx60h0",
		"bcrt1qjhpgmmhaxfwj6t7zf3dvs2fhdhx02g8qn3xwsf",
	}

	previousOutputValue := int64(100000000)

	expectedTxHex := "01000000000101f19194baa0d12141a177f41ea218d93d10e2cf96921e009199215f65a9de990b000000000000000000039003fc0100000000160014a405e97c9e2efdaed32709356655ea03fc1f2a8c9003fc0100000000160014f9974ebea1ca5d6f95fb9f5509f8b3e7bb0047269003fc010000000016001495c28deefd325d2d2fc24c5ac829376dccf520e0024a00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002100000000000000000000000000000000000000000000000000000000000000000000000000"
	expectedTxBytes, err := hex.DecodeString(expectedTxHex)
	if err != nil {
		t.Fatal(err)
	}
	expectedTx := wire.NewMsgTx(0)
	expectedTx.Deserialize(bytes.NewReader(expectedTxBytes))

	actualTx, err := ConstructUnsignedTransaction(
		"0b99dea9655f219991001e9296cfe2103dd918a21ef477a14121d1a0ba9491f1",
		uint32(0),
		previousOutputValue,
		int64(700),
		recipientAddresses,
		&chaincfg.TestNet3Params,
	)
	if err != nil {
		t.Error(err)
	}

	assert.DeepEqual(t, actualTx, expectedTx)
}

func TestBuildSignedTransactionHexString(t *testing.T) {
	unsignedTxHex := "01000000000101f19194baa0d12141a177f41ea218d93d10e2cf96921e009199215f65a9de990b000000000000000000039003fc0100000000160014a405e97c9e2efdaed32709356655ea03fc1f2a8c9003fc0100000000160014f9974ebea1ca5d6f95fb9f5509f8b3e7bb0047269003fc010000000016001495c28deefd325d2d2fc24c5ac829376dccf520e0024a00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002100000000000000000000000000000000000000000000000000000000000000000000000000"
	expectedSignedTx := "01000000000101f19194baa0d12141a177f41ea218d93d10e2cf96921e009199215f65a9de990b000000000000000000039003fc0100000000160014a405e97c9e2efdaed32709356655ea03fc1f2a8c9003fc0100000000160014f9974ebea1ca5d6f95fb9f5509f8b3e7bb0047269003fc010000000016001495c28deefd325d2d2fc24c5ac829376dccf520e0020930060201030201070121020000000007de3ebb640d2b021590c09d5e739597d02d939224d227a17403607500000000"

	publicKey := &cecdsa.PublicKey{
		Curve: elliptic.P224(),
		X:     bigIntFromString(t, "828612351041249926199933036276541218289243364325366441967565889653"),
		Y:     bigIntFromString(t, "985040320797760939221216987624001720525496952574017416820319442840"),
	}

	signature := &ecdsa.Signature{
		R:          big.NewInt(int64(3)),
		S:          big.NewInt(int64(7)),
		RecoveryID: 1,
	}

	signedTxHex, err := BuildSignedTransactionHexString(
		decodeTransaction(t, unsignedTxHex),
		signature,
		publicKey,
	)
	if err != nil {
		t.Fatalf("failed to build signed transaction: %v", err)
	}

	if signedTxHex != expectedSignedTx {
		t.Errorf(
			"invalid signed transaction\n- actual\n+ expected\n%s",
			cmp.Diff(decodeTransaction(t, signedTxHex), decodeTransaction(t, expectedSignedTx)))
	}
}

func decodeTransaction(t *testing.T, txHex string) *wire.MsgTx {
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		t.Fatalf("failed to decode transaction [%s]: [%v]", txHex, err)
	}
	tx := wire.NewMsgTx(0)
	tx.BtcDecode(bytes.NewReader(txBytes), wire.ProtocolVersion, wire.WitnessEncoding)

	return tx
}

func bigIntFromString(t *testing.T, s string) *big.Int {
	bigInt, ok := new(big.Int).SetString(s, 0)
	if !ok {
		t.Errorf("Something went wrong creating a big int from %s", s)
	}
	return bigInt
}
