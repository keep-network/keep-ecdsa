package client

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-ecdsa/pkg/chain/local"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
)

const signStateConfirmTimeout = 10 * time.Second

var keepAddress = common.HexToAddress("0x4e09cadc7037afa36603138d1c0b76fe2aa5039c")
var digest = sha256.Sum256([]byte("Do or do not. There is no try."))

func TestDoGenerateKey(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, _, _ := newDeduplicator(ctx)

	canGenerate := deduplicator.notifyKeyGenStarted(keepAddress)
	if !canGenerate {
		t.Fatal("should be allowed to generate a key")
	}
}

func TestDoNotGenerateKeyIfCurrentlyGenerating(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, _, _ := newDeduplicator(ctx)

	deduplicator.notifyKeyGenStarted(keepAddress)

	canGenerate := deduplicator.notifyKeyGenStarted(keepAddress)
	if canGenerate {
		t.Fatal("should not be allowed to generate a key")
	}
}

func TestDoNotGenerateKeyIfAlreadyGenerated(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, registry, _ := newDeduplicator(ctx)

	deduplicator.notifyKeyGenStarted(keepAddress)
	registry.AddSigner(keepAddress)
	deduplicator.notifyKeyGenCompleted(keepAddress)

	canGenerate := deduplicator.notifyKeyGenStarted(keepAddress)
	if canGenerate {
		t.Fatal("should not be allowed to generate a key")
	}
}

func TestDoSign(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, _, chain := newDeduplicator(ctx)

	chain.OpenKeep(keepAddress, []common.Address{})

	var keepPublicKey [64]byte
	rand.Read(keepPublicKey[:])

	err := chain.SubmitKeepPublicKey(keepAddress, keepPublicKey)
	if err != nil {
		t.Fatal(err)
	}

	err = chain.RequestSignature(keepAddress, digest)
	if err != nil {
		t.Fatal(err)
	}

	canSign, err := deduplicator.notifySigningStarted(
		signStateConfirmTimeout,
		keepAddress,
		digest,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !canSign {
		t.Errorf("should be allowed to sign")
	}
}

func TestDoNotSignIfCurrentlySigning(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, _, chain := newDeduplicator(ctx)

	chain.OpenKeep(keepAddress, []common.Address{})

	var keepPublicKey [64]byte
	rand.Read(keepPublicKey[:])

	err := chain.SubmitKeepPublicKey(keepAddress, keepPublicKey)
	if err != nil {
		t.Fatal(err)
	}

	err = chain.RequestSignature(keepAddress, digest)
	if err != nil {
		t.Fatal(err)
	}

	deduplicator.notifySigningStarted(
		signStateConfirmTimeout,
		keepAddress,
		digest,
	)

	canSign, err := deduplicator.notifySigningStarted(
		signStateConfirmTimeout,
		keepAddress,
		digest,
	)
	if err != nil {
		t.Fatal(err)
	}
	if canSign {
		t.Errorf("should not be allowed to sign")
	}
}

func TestDoNotSignIfNotAwaitingASignature(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, _, chain := newDeduplicator(ctx)

	chain.OpenKeep(keepAddress, []common.Address{})

	var keepPublicKey [64]byte
	rand.Read(keepPublicKey[:])

	err := chain.SubmitKeepPublicKey(keepAddress, keepPublicKey)
	if err != nil {
		t.Fatal(err)
	}

	canSign, err := deduplicator.notifySigningStarted(
		signStateConfirmTimeout,
		keepAddress,
		digest,
	)
	if err != nil {
		t.Fatal(err)
	}
	if canSign {
		t.Errorf("should not be allowed to sign")
	}
}

func TestDoSignOneMoreTime(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, _, chain := newDeduplicator(ctx)

	chain.OpenKeep(keepAddress, []common.Address{})

	var keepPublicKey [64]byte
	rand.Read(keepPublicKey[:])

	err := chain.SubmitKeepPublicKey(keepAddress, keepPublicKey)
	if err != nil {
		t.Fatal(err)
	}

	//
	// request and provide a signature, notify it's been provided
	//
	err = chain.RequestSignature(keepAddress, digest)
	if err != nil {
		t.Fatal(err)
	}

	canSign, err := deduplicator.notifySigningStarted(
		signStateConfirmTimeout,
		keepAddress,
		digest,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !canSign {
		t.Errorf("should be allowed to sign")
	}

	signature := &ecdsa.Signature{
		R:          big.NewInt(10),
		S:          big.NewInt(11),
		RecoveryID: 1,
	}

	err = chain.SubmitSignature(keepAddress, signature)
	if err != nil {
		t.Fatal(err)
	}

	deduplicator.notifySigningCompleted(keepAddress, digest)

	//
	// request signature with the same digest one more time - should work
	//
	err = chain.RequestSignature(keepAddress, digest)
	if err != nil {
		t.Fatal(err)
	}

	canSign, err = deduplicator.notifySigningStarted(
		signStateConfirmTimeout,
		keepAddress,
		digest,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !canSign {
		t.Errorf("should be allowed to sign")
	}
}

func TestDoClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, registry, _ := newDeduplicator(ctx)
	registry.AddSigner(keepAddress)

	canClose := deduplicator.notifyClosingStarted(keepAddress)
	if !canClose {
		t.Fatal("should be allowed to close a keep")
	}
}

func TestDoNotCloseIfCurrentlyClosing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, registry, _ := newDeduplicator(ctx)
	registry.AddSigner(keepAddress)

	deduplicator.notifyClosingStarted(keepAddress)

	canClose := deduplicator.notifyClosingStarted(keepAddress)
	if canClose {
		t.Fatal("should not be allowed to close a keep")
	}
}

func TestDoNotCloseIfAlreadyClosed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, _, _ := newDeduplicator(ctx)
	// keep not in the registry

	canClose := deduplicator.notifyClosingStarted(keepAddress)
	if canClose {
		t.Fatal("should not be allowed to close a keep")
	}
}

func TestDoTerminate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, registry, _ := newDeduplicator(ctx)
	registry.AddSigner(keepAddress)

	canTerminate := deduplicator.notifyTerminatingStarted(keepAddress)
	if !canTerminate {
		t.Fatal("should be allowed to terminate a keep")
	}
}

func TestDoNotTerminateIfCurrentlyTerminating(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, registry, _ := newDeduplicator(ctx)
	registry.AddSigner(keepAddress)

	deduplicator.notifyTerminatingStarted(keepAddress)

	canTerminate := deduplicator.notifyTerminatingStarted(keepAddress)
	if canTerminate {
		t.Fatal("should not be allowed to terminate a keep")
	}
}

func TestDoNotTerminateIfAlreadyTerminated(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, _, _ := newDeduplicator(ctx)
	// keep not in the registry

	canTerminate := deduplicator.notifyTerminatingStarted(keepAddress)
	if canTerminate {
		t.Fatal("should not be allowed to terminate a keep")
	}
}

func newDeduplicator(ctx context.Context) (
	*eventDeduplicator,
	*mockRegistry,
	local.Chain,

) {
	mockRegistry := &mockRegistry{
		keeps: make(map[common.Address]bool),
	}

	chain := local.Connect(ctx)

	deduplicator := newEventDeduplicator(
		mockRegistry,
		chain,
	)

	return deduplicator, mockRegistry, chain
}

type mockRegistry struct {
	keeps map[common.Address]bool
}

func (mr *mockRegistry) AddSigner(keepAddress common.Address) {
	mr.keeps[keepAddress] = true
}

func (mr *mockRegistry) HasSigner(keepAddress common.Address) bool {
	return mr.keeps[keepAddress]
}
