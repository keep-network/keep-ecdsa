package local

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
)

func TestCreateKeepDuplicate(t *testing.T) {
	chain := initializeLocalChain()

	keepAddress := eth.KeepAddress([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	keepMembers := []common.Address{}

	expectedError := fmt.Errorf("keep already exists for address [0x0000000000000000000000000000000000000001]")

	err := chain.CreateKeep(keepAddress, keepMembers)
	if err != nil {
		t.Fatal(err)
	}

	err = chain.CreateKeep(keepAddress, keepMembers)
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
	keepMembers := []common.Address{}

	expectedPublicKey := [64]byte{}

	err := chain.CreateKeep(keepAddress, keepMembers)
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
