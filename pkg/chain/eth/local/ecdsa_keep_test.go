package local

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
)

func TestRequestSignatureNonexistentKeep(t *testing.T) {
	chain := initializeLocalChain()
	keepAddress := common.Address([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	digest := [32]byte{1}
	expectedError := fmt.Errorf("failed to find keep with address: [0x0000000000000000000000000000000000000001]")

	err := chain.requestSignature(keepAddress, digest)

	if !reflect.DeepEqual(err, expectedError) {
		t.Fatalf(
			"unexpected error\nexpected: [%v]\nactual:   [%v]",
			expectedError,
			err.Error(),
		)
	}
}

func TestRequestSignatureNoHandler(t *testing.T) {
	chain := initializeLocalChain()
	keepAddress := common.Address([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	digest := [32]byte{1}

	err := chain.CreateKeep(keepAddress)
	if err != nil {
		t.Fatal(err)
	}

	err = chain.requestSignature(keepAddress, digest)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRequestSignature(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	chain := initializeLocalChain()
	keepAddress := common.Address([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	digest := [32]byte{1}
	eventEmitted := make(chan *eth.SignatureRequestedEvent)
	handler := func(event *eth.SignatureRequestedEvent) {
		eventEmitted <- event
	}

	err := chain.CreateKeep(keepAddress)
	if err != nil {
		t.Fatal(err)
	}
	chain.keeps[keepAddress].signatureRequestedHandlers[0] = handler

	err = chain.requestSignature(keepAddress, digest)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-eventEmitted:
		if !bytes.Equal(event.Digest[:], digest[:]) {
			t.Errorf(
				"unexpected digest from signature request\nexpected: %x\nactual:   %x\n",
				digest,
				event.Digest,
			)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}
