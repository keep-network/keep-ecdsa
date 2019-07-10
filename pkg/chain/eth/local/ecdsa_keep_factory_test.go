package local

import (
	"bytes"
	"testing"

	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
)

func TestCreateKeep(t *testing.T) {
	chain := initializeLocalChain()

	keepAddress := eth.KeepAddress([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	expectedPublicKey := [64]byte{}

	err := chain.createKeep(keepAddress)
	if err != nil {
		t.Fatal(err)
	}

	keep, ok := chain.keeps[keepAddress]
	if !ok {
		t.Fatal("keep not found after creation")
	}

	if !bytes.Equal(keep.publicKey[:], expectedPublicKey[:]) {
		t.Errorf(
			"unexpected publicKey value for keep\nexpected: %x\nactual:   %x\n",
			expectedPublicKey,
			keep.publicKey,
		)
	}
}
