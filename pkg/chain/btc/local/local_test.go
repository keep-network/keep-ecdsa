package local

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/keep-network/keep-tecdsa/internal/testdata"
	"github.com/keep-network/keep-tecdsa/pkg/utils"
)

func TestPublishTransaction(t *testing.T) {
	chain := Connect()

	var tests = map[string]struct {
		rawTx          string
		expectedResult string
		expectedError  error
	}{
		"successful transaction publication": {
			rawTx:          testdata.ValidTx.SignedRaw,
			expectedResult: testdata.ValidTx.Hash,
		},
		"previous transaction not found for hash": {
			rawTx:          txPreviousTxNotRegistered(),
			expectedResult: "",
			expectedError:  fmt.Errorf("previous transaction not found for hash [0993fb7ebd5b259ba6ad426d92c9fa19354accd3780482d08659ef1544980105]"),
		},
		"transaction already registered": {
			rawTx:          testdata.InitialTx.SignedRaw,
			expectedResult: "",
			expectedError:  fmt.Errorf("transaction already registered on chain for hash [%s]", testdata.InitialTx.Hash),
		},
		"failed transaction publication": {
			rawTx:          "",
			expectedResult: "",
			expectedError:  fmt.Errorf("empty transaction provided"),
		},
		"transaction validation failed - unsigned transaction": {
			rawTx:          txUnsigned(),
			expectedResult: "",
			expectedError:  fmt.Errorf("failed to validate transaction: [should have exactly two items in witness, instead have 0]"),
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
	transaction, _ := hex.DecodeString(testdata.ValidTx.SignedRaw)

	msgTx, err := utils.DeserializeTransaction(transaction)
	if err != nil {
		t.Fatal(err)
	}

	if err := ValidateTransaction(
		testdata.ValidTx.PreviousOutScript,
		testdata.ValidTx.PreviousOutAmount,
		msgTx,
	); err != nil {
		t.Errorf("transaction validation failed: [%v]", err)
	}
}

func txPreviousTxNotRegistered() string {
	msgTx := wire.NewMsgTx(wire.TxVersion)

	// Fake hash of not registered transaction.
	hash, _ := chainhash.NewHashFromStr("0993fb7ebd5b259ba6ad426d92c9fa19354accd3780482d08659ef1544980105")

	outPoint := wire.NewOutPoint(hash, 0)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	msgTx.AddTxIn(txIn)

	serializedTx, _ := utils.SerializeTransaction(msgTx)

	return hex.EncodeToString(serializedTx)
}

func txUnsigned() string {
	msgTx := wire.NewMsgTx(wire.TxVersion)

	previousHash, _ := chainhash.NewHashFromStr(testdata.InitialTx.Hash)
	outPoint := wire.NewOutPoint(previousHash, 0)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	msgTx.AddTxIn(txIn)

	serializedTx, _ := utils.SerializeTransaction(msgTx)

	return hex.EncodeToString(serializedTx)
}
