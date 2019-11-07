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

	keepAddress := eth.KeepAddress([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	keepMembers := []common.Address{common.BytesToAddress([]byte{1, 2, 3})}

	expectedEvent := &eth.ECDSAKeepCreatedEvent{
		KeepAddress: keepAddress,
		Members:     keepMembers,
	}

	subscription, err := chain.OnECDSAKeepCreated(
		func(event *eth.ECDSAKeepCreatedEvent) {
			eventFired <- event
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer subscription.Unsubscribe()

	err = chain.CreateKeep(keepAddress, keepMembers)
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

	chain := initializeLocalChain()
	eventFired := make(chan *eth.SignatureRequestedEvent)

	keepAddress := eth.KeepAddress([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	keepMembers := []common.Address{}

	digest := [32]byte{1}

	err := chain.CreateKeep(keepAddress, keepMembers)
	if err != nil {
		t.Fatal(err)
	}

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

	err = chain.RequestSignature(keepAddress, digest)
	if err != nil {
		t.Fatal(err)
	}

	expectedEvent := &eth.SignatureRequestedEvent{
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
	chain := initializeLocalChain()

	keepAddress := common.HexToAddress("0x41048F9B90290A2e96D07f537F3A7E97620E9e47")
	keepMembers := []common.Address{}
	keepPublicKey := [64]byte{11, 12, 13, 14, 15, 16}

	expectedDuplicationError := fmt.Errorf(
		"public key already submitted for keep [%s]",
		keepAddress.String(),
	)

	err := chain.CreateKeep(keepAddress, keepMembers)
	if err != nil {
		t.Fatal(err)
	}

	err = chain.SubmitKeepPublicKey(
		keepAddress,
		keepPublicKey,
	)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(keepPublicKey, chain.keeps[keepAddress].publicKey) {
		t.Errorf(
			"unexpected result\nexpected: [%+v]\nactual:   [%+v]",
			keepPublicKey,
			chain.keeps[keepAddress].publicKey,
		)
	}

	err = chain.SubmitKeepPublicKey(
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

func initializeLocalChain() *LocalChain {
	return Connect().(*LocalChain)
}
