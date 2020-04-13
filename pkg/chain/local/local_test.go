package local

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

func TestOnBondedECDSAKeepCreated(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	handle := initializeLocalChain()
	eventFired := make(chan *chain.BondedECDSAKeepCreatedEvent)
	keepAddress := common.Address([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	expectedEvent := &chain.BondedECDSAKeepCreatedEvent{
		KeepAddress: keepAddress,
	}

	subscription, err := handle.OnBondedECDSAKeepCreated(
		func(event *chain.BondedECDSAKeepCreatedEvent) {
			eventFired <- event
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer subscription.Unsubscribe()

	err = handle.createKeep(keepAddress)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-eventFired:
		if !reflect.DeepEqual(event, expectedEvent) {
			t.Fatalf(
				"unexpected keep creation event\nexpected: [%v]\nactual:   [%v]",
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

	handle := initializeLocalChain()
	eventFired := make(chan *chain.SignatureRequestedEvent)
	keepAddress := common.Address([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	digest := [32]byte{1}

	err := handle.createKeep(keepAddress)
	if err != nil {
		t.Fatal(err)
	}

	subscription, err := handle.OnSignatureRequested(
		keepAddress,
		func(event *chain.SignatureRequestedEvent) {
			eventFired <- event
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer subscription.Unsubscribe()

	err = handle.requestSignature(keepAddress, digest)
	if err != nil {
		t.Fatal(err)
	}

	expectedEvent := &chain.SignatureRequestedEvent{
		Digest: digest,
	}

	select {
	case event := <-eventFired:
		if !reflect.DeepEqual(event, expectedEvent) {
			t.Fatalf(
				"unexpected signature requested event\nexpected: [%v]\nactual:   [%v]",
				expectedEvent,
				event,
			)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}

func TestSubmitKeepPublicKey(t *testing.T) {
	handle := initializeLocalChain()
	keepAddress := common.HexToAddress("0x41048F9B90290A2e96D07f537F3A7E97620E9e47")
	keepPublicKey := [64]byte{11, 12, 13, 14, 15, 16}
	expectedDuplicationError := fmt.Errorf(
		"public key already submitted for keep [%s]",
		keepAddress.String(),
	)

	err := handle.createKeep(keepAddress)
	if err != nil {
		t.Fatal(err)
	}

	err = handle.SubmitKeepPublicKey(
		keepAddress,
		keepPublicKey,
	)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(keepPublicKey, handle.keeps[keepAddress].publicKey) {
		t.Errorf(
			"unexpected result\nexpected: [%+v]\nactual:   [%+v]",
			keepPublicKey,
			handle.keeps[keepAddress].publicKey,
		)
	}

	err = handle.SubmitKeepPublicKey(
		keepAddress,
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
