package local

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
)

func TestRequestSignatureNonexistentKeep(t *testing.T) {
	chain := initializeLocalChain()
	keepAddress := eth.KeepAddress([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	digest := []byte{1}

	expectedError := "keep not found for address [0x0000000000000000000000000000000000000001]"
	err := chain.requestSignature(keepAddress, digest)

	if err.Error() != expectedError {
		t.Fatalf(
			"unexpected error\nexpected: [%v]\nactual:   [%v]",
			expectedError,
			err.Error(),
		)
	}
}

func TestRequestSignatureNoHandler(t *testing.T) {
	chain := initializeLocalChain()
	keepAddress := eth.KeepAddress([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	digest := []byte{1}

	chain.createKeep(keepAddress)
	chain.requestSignature(keepAddress, digest)
}

func TestRequestSignature(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	chain := initializeLocalChain()
	keepAddress := eth.KeepAddress([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	digest := []byte{1}

	eventEmitted := make(chan *eth.SignatureRequestedEvent)

	handler := func(event *eth.SignatureRequestedEvent) {
		eventEmitted <- event
	}

	chain.createKeep(keepAddress)
	chain.keeps[keepAddress].signatureRequestedHandlers[0] = handler

	err := chain.requestSignature(keepAddress, digest)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	select {
	case event := <-eventEmitted:
		if !bytes.Equal(event.Digest, digest) {
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
