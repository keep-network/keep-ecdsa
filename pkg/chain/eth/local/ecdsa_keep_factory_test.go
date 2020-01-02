package local

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestCreateKeep(t *testing.T) {
	chain := initializeLocalChain()
	keepAddress := common.Address([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	expectedPublicKey := [64]byte{}

	chain.CreateKeep(keepAddress)

	keep, ok := keeps[keepAddress]
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
