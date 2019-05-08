package local

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/keep-network/keep-tecdsa/pkg/utils"
)

const (
	validTxRaw  = "0100000000010199dffb02e8f3e3f8053eecf6110a95f77ba658690dfc3a447b7e52cf34ca135e0000000000ffffffff02581b000000000000160014d849b1e1cede2ac7d7188cf8700e97d6975c91c4e8030000000000001976a914d849b1e1cede2ac7d7188cf8700e97d6975c91c488ac02483045022100ecadce07f5c9d84b4fa1b2728806135acd81ad9398c9673eeb4e161d42364b92022076849daa2108ed2a135d16eb9e103c5819db014ea2bad5c92f4aeecf47bf9ac80121028896955d043b5a43957b21901f2cce9f0bfb484531b03ad6cd3153e45e73ee2e00000000"
	validTxHash = "ec367c260ead9e3c91583175f35382e22b66df6d59fd0aac175bb36519b664f7"
)

func TestPublishTransaction(t *testing.T) {
	chain := Connect()

	var tests = map[string]struct {
		rawTx          string
		expectedResult string
		expectedError  error
	}{
		"successful transaction publication": {
			rawTx:          validTxRaw,
			expectedResult: validTxHash,
		},
		"previous transaction not found for hash": {
			rawTx:          txInvalidMissingPreviousHash(),
			expectedResult: "",
			expectedError:  fmt.Errorf("previous transaction not found for hash [b7243a2bb8721011c226d15d9a3495032f218bf93e2de23acf2d752b255dc8c5]"),
		},
		"transaction already registered": {
			rawTx:          initialTransaction,
			expectedResult: "",
			expectedError:  fmt.Errorf("transaction already registered on chain for hash [5e13ca34cf527e7b443afc0d6958a67bf7950a11f6ec3e05f8e3f3e802fbdf99]"),
		},
		"failed transaction publication": {
			rawTx:          "",
			expectedResult: "",
			expectedError:  fmt.Errorf("empty transaction provided"),
		},
		"transaction validation failed - unsigned transaction": {
			rawTx:          txUnsigned(),
			expectedResult: "",
			expectedError:  fmt.Errorf("transaction validation failed [should have exactly two items in witness, instead have 0]"),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			result, err := chain.PublishTransaction(test.rawTx)

			if !reflect.DeepEqual(test.expectedError, err) {
				t.Errorf(
					"unexpected error\nexpected: %v\nactual:   %v\n",
					test.expectedError,
					err,
				)
			}

			if test.expectedResult != result {
				t.Errorf(
					"unexpected result\nexpected: %v\nactual:   %v\n",
					test.expectedResult,
					result,
				)
			}
		})
	}
}

func TestValidateTransaction(t *testing.T) {
	previousOutputScript := "0014d849b1e1cede2ac7d7188cf8700e97d6975c91c4"
	previousOutputAmout := int64(9000)
	transaction, _ := hex.DecodeString(validTxRaw)

	msgTx, err := utils.DeserializeTransaction(transaction)
	if err != nil {
		t.Fatal(err)
	}

	if err := ValidateTransaction(
		previousOutputScript,
		previousOutputAmout,
		msgTx,
	); err != nil {
		t.Errorf("transaction validation failed [%s]", err)
	}
}

func txInvalidMissingPreviousHash() string {
	msgTx := wire.NewMsgTx(wire.TxVersion)

	hash := chainhash.DoubleHashH([]byte("invalid"))
	outPoint := wire.NewOutPoint(&hash, 0)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	msgTx.AddTxIn(txIn)

	serializedTx, _ := utils.SerializeTransaction(msgTx)

	return hex.EncodeToString(serializedTx)
}

func txUnsigned() string {
	msgTx := wire.NewMsgTx(wire.TxVersion)

	previousHash, _ := chainhash.NewHashFromStr("5e13ca34cf527e7b443afc0d6958a67bf7950a11f6ec3e05f8e3f3e802fbdf99")
	outPoint := wire.NewOutPoint(previousHash, 0)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	msgTx.AddTxIn(txIn)

	serializedTx, _ := utils.SerializeTransaction(msgTx)

	return hex.EncodeToString(serializedTx)
}
