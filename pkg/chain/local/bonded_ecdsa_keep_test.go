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
	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

func TestKeepLookupNonexistentKeep(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	localChain := initializeLocalChain(ctx)
	keepAddress := common.Address([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	expectedError := fmt.Errorf("failed to find keep with address: [0x0000000000000000000000000000000000000001]")

	manager, err := localChain.BondedECDSAKeepManager()
	if err != nil {
		t.Fatalf(
			"unexpected error\nexpected: [%v]\nactual:   [%v]",
			expectedError,
			err.Error(),
		)
	}

	_, err = manager.GetKeepWithID(keepAddress)
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

	localChain := initializeLocalChain(ctx)
	keepAddress := common.Address([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	digest := [32]byte{1}

	keep, err := localChain.createKeep(keepAddress)
	if err != nil {
		t.Fatal(err)
	}

	var keepPubkey [64]byte
	rand.Read(keepPubkey[:])

	err = keep.SubmitKeepPublicKey(keepPubkey)
	if err != nil {
		t.Fatal(err)
	}

	err = keep.RequestSignature(digest)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRequestSignature(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelCtx()

	localChain := initializeLocalChain(ctx)
	keepAddress := common.Address([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	digest := [32]byte{1}
	eventEmitted := make(chan *chain.SignatureRequestedEvent)
	handler := func(event *chain.SignatureRequestedEvent) {
		eventEmitted <- event
	}

	keep, err := localChain.createKeep(keepAddress)
	if err != nil {
		t.Fatal(err)
	}

	var keepPubkey [64]byte
	rand.Read(keepPubkey[:])

	err = keep.SubmitKeepPublicKey(keepPubkey)
	if err != nil {
		t.Fatal(err)
	}

	keep.signatureRequestedHandlers[0] = handler

	err = keep.RequestSignature(digest)
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
