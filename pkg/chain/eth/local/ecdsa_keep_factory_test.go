package local

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
)

func TestCreateKeepDuplicate(t *testing.T) {
	chain := initializeLocalChain()

	keepAddress := eth.KeepAddress([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	expectedError := fmt.Errorf("keep already exists for address [0x0000000000000000000000000000000000000001]")

	err := chain.createKeep(keepAddress)
	if err != nil {
		t.Fatal(err)
	}

	err = chain.createKeep(keepAddress)
	if !reflect.DeepEqual(err, expectedError) {
		t.Fatalf(
			"unexpected error\nexpected: [%v]\nactual:   [%v]",
			expectedError,
			err.Error(),
		)
	}
}

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
