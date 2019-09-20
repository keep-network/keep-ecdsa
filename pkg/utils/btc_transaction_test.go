package utils

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestSerializeTransactionRoundTrip(t *testing.T) {
	expectedTransaction, _ := hex.DecodeString("0100000000010199dffb02e8f3e3f8053eecf6110a95f77ba658690dfc3a447b7e52cf34ca135e0000000000ffffffff02581b000000000000160014d849b1e1cede2ac7d7188cf8700e97d6975c91c4e8030000000000001976a914d849b1e1cede2ac7d7188cf8700e97d6975c91c488ac02483045022100ecadce07f5c9d84b4fa1b2728806135acd81ad9398c9673eeb4e161d42364b92022076849daa2108ed2a135d16eb9e103c5819db014ea2bad5c92f4aeecf47bf9ac80121028896955d043b5a43957b21901f2cce9f0bfb484531b03ad6cd3153e45e73ee2e00000000")

	msgTx, err := DeserializeTransaction(expectedTransaction)
	if err != nil {
		t.Fatalf("transaction deserialization failed: [%v]", err)
	}

	result, err := SerializeTransaction(msgTx)
	if err != nil {
		t.Fatalf("transaction serialization failed: [%v]", err)
	}

	if !bytes.Equal(result, expectedTransaction) {
		t.Errorf("unexpected transaction\nexpected: %x\nactual:   %x\n",
			expectedTransaction,
			result,
		)
	}
}
