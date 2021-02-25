package local

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
)

func TestRequestSignatureNonexistentKeep(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	chain := initializeLocalChain(ctx)
	keepAddress := common.Address([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	digest := [32]byte{1}
	expectedError := fmt.Errorf("failed to find keep with address: [0x0000000000000000000000000000000000000001]")

	err := chain.RequestSignature(keepAddress, digest)

	if !reflect.DeepEqual(err, expectedError) {
		t.Fatalf(
			"unexpected error\nexpected: [%v]\nactual:   [%v]",
			expectedError,
			err.Error(),
		)
	}
}

func TestRequestSignatureNoHandler(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	chain := initializeLocalChain(ctx)
	keepAddress := common.Address([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	digest := [32]byte{1}

	err := chain.createKeep(keepAddress)
	if err != nil {
		t.Fatal(err)
	}

	var keepPubkey [64]byte
	rand.Read(keepPubkey[:])

	err = chain.SubmitKeepPublicKey(keepAddress, keepPubkey)
	if err != nil {
		t.Fatal(err)
	}

	err = chain.RequestSignature(keepAddress, digest)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRequestSignature(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelCtx()

	chain := initializeLocalChain(ctx)
	keepAddress := common.Address([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	digest := [32]byte{1}
	eventEmitted := make(chan *eth.SignatureRequestedEvent)
	handler := func(event *eth.SignatureRequestedEvent) {
		eventEmitted <- event
	}

	err := chain.createKeep(keepAddress)
	if err != nil {
		t.Fatal(err)
	}

	var keepPubkey [64]byte
	rand.Read(keepPubkey[:])

	err = chain.SubmitKeepPublicKey(keepAddress, keepPubkey)
	if err != nil {
		t.Fatal(err)
	}

	chain.keeps[keepAddress].signatureRequestedHandlers[0] = handler

	err = chain.RequestSignature(keepAddress, digest)
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
