package local

import (
	"context"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

func TestOnECDSAKeepCreated(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	chain := initializeLocalChain()
	eventFired := make(chan *eth.ECDSAKeepCreatedEvent)
	keepAddress := common.Address([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	expectedEvent := &eth.ECDSAKeepCreatedEvent{
		KeepAddress: keepAddress,
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

	chain.CreateKeep(keepAddress)

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
	keepAddress := common.Address([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	digest := [32]byte{1}

	chain.CreateKeep(keepAddress)

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

	err = chain.requestSignature(keepAddress, digest)
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
	keepPublicKey := [64]byte{11, 12, 13, 14, 15, 16}

	chain.CreateKeep(keepAddress)

	err := chain.SubmitKeepPublicKey(
		keepAddress,
		keepPublicKey,
	)
	if err != nil {
		t.Fatal(err)
	}

	publicKey, err := chain.GetKeepPublicKey(keepAddress)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(keepPublicKey, publicKey) {
		t.Errorf(
			"unexpected result\nexpected: [%+v]\nactual:   [%+v]",
			keepPublicKey,
			keeps[keepAddress].publicKey,
		)
	}
}

func TestSubmitSignature(t *testing.T) {
	chain := initializeLocalChain()
	keepAddress := common.HexToAddress("0x41048F9B90290A2e96D07f537F3A7E97620E9e47")
	signature := &ecdsa.Signature{R: big.NewInt(8), S: big.NewInt(7)}

	chain.CreateKeep(keepAddress)

	err := chain.SubmitSignature(
		keepAddress,
		signature,
	)
	if err != nil {
		t.Fatal(err)
	}

	signatures, err := chain.GetSignatures(keepAddress, digest)
	if err != nil {
		t.Fatal(err)
	}

	if len(signatures) != 1 {
		t.Errorf(
			"invalid number of stored signatures\nexpected: %v\nactual:   %v",
			1,
			len(signatures),
		)
	}

	if !reflect.DeepEqual(signatures[0], signature) {
		t.Errorf(
			"invalid stored signature\nexpected: %v\nactual:   %v",
			signature,
			signatures[0],
		)
	}
}

func initializeLocalChain() *LocalChain {
	keeps = make(map[eth.KeepAddress]*localKeep)
	return Connect().(*LocalChain)
}
