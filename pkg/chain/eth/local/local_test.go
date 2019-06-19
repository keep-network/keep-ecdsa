package local

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
)

func TestOnECDSAKeepCreated(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	chain := initializeLocalChain()

	eventFired := make(chan *eth.ECDSAKeepCreatedEvent)

	subscription, err := chain.OnECDSAKeepCreated(
		func(event *eth.ECDSAKeepCreatedEvent) {
			eventFired <- event
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer subscription.Unsubscribe()

	keepAddress := common.Address([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	chain.createKeep(keepAddress)

	expectedEvent := &eth.ECDSAKeepCreatedEvent{
		KeepAddress: keepAddress,
	}

	select {
	case event := <-eventFired:
		if !reflect.DeepEqual(event, expectedEvent) {
			t.Fatalf(
				"Unexpected keep creation event\nExpected: [%v]\nActual:   [%v]",
				expectedEvent,
				event,
			)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}

func TestOnSignatureRequested(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	chain := initializeLocalChain()
	eventFired := make(chan *eth.SignatureRequestedEvent)

	keepAddress := common.Address([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	digest := []byte{0,}
	chain.createKeep(keepAddress)

	subscription, err := chain.OnSignatureRequested(
		keepAddress,
		func(event *eth.SignatureRequestedEvent) {
			eventFired <- event
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer subscription.Unsubscribe()

	chain.requestSignature(keepAddress, digest)

	expectedEvent := &eth.SignatureRequestedEvent{
		Digest: digest,
	}

	select {
	case event := <-eventFired:
		if !reflect.DeepEqual(event, expectedEvent) {
			t.Fatalf(
				"Unexpected signature requested event\nExpected: [%v]\nActual:   [%v]",
				expectedEvent,
				event,
			)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
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

func initializeLocalChain() *localChain {
	return Connect().(*localChain)
}
