package local

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestOnECDSAKeepCreated(t *testing.T) {
	// TODO: Implement
	t.SkipNow()
}

func TestSubmitKeepPublicKey(t *testing.T) {
	keepAddress := "0x41048F9B90290A2e96D07f537F3A7E97620E9e47"
	keepPublicKey := [64]byte{11, 12, 13, 14, 15, 16}
	expectedDuplicationError := fmt.Errorf(
		"public key already submitted for keep [%s]",
		keepAddress,
	)

	chain := initializeLocalChain()

	err := chain.SubmitKeepPublicKey(
		common.HexToAddress(keepAddress),
		keepPublicKey,
	)
	if err != nil {
		t.Fatalf("unexpected error: [%s]", err)
	}

	if !reflect.DeepEqual(keepPublicKey, chain.keeps[keepAddress]) {
		t.Errorf(
			"unexpected result\nexpected: [%+v]\nactual:   [%+v]",
			keepPublicKey,
			chain.keeps[keepAddress],
		)
	}

	err = chain.SubmitKeepPublicKey(
		common.HexToAddress(keepAddress),
		keepPublicKey,
	)
	if !reflect.DeepEqual(expectedDuplicationError, err) {
		t.Errorf(
			"unexpected error\nexpected: [%+v]\nactual:   [%+v]",
			expectedDuplicationError,
			err,
		)
	}
}

func initializeLocalChain() *LocalChain {
	return Connect().(*LocalChain)
}
