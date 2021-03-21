package event

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/local"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
)

const signStateConfirmTimeout = 10 * time.Second

var keepAddress1 = common.HexToAddress(keepID1String)
var digest = sha256.Sum256([]byte("Do or do not. There is no try."))

func TestDoGenerateKey(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, _, _ := newDeduplicator(ctx)

	canGenerate := deduplicator.NotifyKeyGenStarted(keepID1)
	if !canGenerate {
		t.Fatal("should be allowed to generate a key")
	}
}

func TestDoNotGenerateKeyIfCurrentlyGenerating(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, _, _ := newDeduplicator(ctx)

	deduplicator.NotifyKeyGenStarted(keepID1)

	canGenerate := deduplicator.NotifyKeyGenStarted(keepID1)
	if canGenerate {
		t.Fatal("should not be allowed to generate a key")
	}
}

func TestDoNotGenerateKeyIfAlreadyGenerated(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, registry, _ := newDeduplicator(ctx)

	deduplicator.NotifyKeyGenStarted(keepID1)
	registry.AddSigner(keepID1)
	deduplicator.NotifyKeyGenCompleted(keepID1)

	canGenerate := deduplicator.NotifyKeyGenStarted(keepID1)
	if canGenerate {
		t.Fatal("should not be allowed to generate a key")
	}
}

func TestDoSign(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, _, chain := newDeduplicator(ctx)

	keep := chain.OpenKeep(keepAddress1, []common.Address{})

	var keepPublicKey [64]byte
	rand.Read(keepPublicKey[:])

	err := keep.SubmitKeepPublicKey(keepPublicKey)
	if err != nil {
		t.Fatal(err)
	}

	err = chain.RequestSignature(keepAddress1, digest)
	if err != nil {
		t.Fatal(err)
	}

	canSign, err := deduplicator.NotifySigningStarted(
		signStateConfirmTimeout,
		keep,
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

	keep := chain.OpenKeep(keepAddress1, []common.Address{})

	var keepPublicKey [64]byte
	rand.Read(keepPublicKey[:])

	err := keep.SubmitKeepPublicKey(keepPublicKey)
	if err != nil {
		t.Fatal(err)
	}

	err = chain.RequestSignature(keepAddress1, digest)
	if err != nil {
		t.Fatal(err)
	}

	deduplicator.NotifySigningStarted(
		signStateConfirmTimeout,
		keep,
		digest,
	)

	canSign, err := deduplicator.NotifySigningStarted(
		signStateConfirmTimeout,
		keep,
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

	keep := chain.OpenKeep(keepAddress1, []common.Address{})

	var keepPublicKey [64]byte
	rand.Read(keepPublicKey[:])

	err := keep.SubmitKeepPublicKey(keepPublicKey)
	if err != nil {
		t.Fatal(err)
	}

	canSign, err := deduplicator.NotifySigningStarted(
		signStateConfirmTimeout,
		keep,
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

	keep := chain.OpenKeep(keepAddress1, []common.Address{})

	var keepPublicKey [64]byte
	rand.Read(keepPublicKey[:])

	err := keep.SubmitKeepPublicKey(keepPublicKey)
	if err != nil {
		t.Fatal(err)
	}

	//
	// request and provide a signature, notify it's been provided
	//
	err = chain.RequestSignature(keepAddress1, digest)
	if err != nil {
		t.Fatal(err)
	}

	canSign, err := deduplicator.NotifySigningStarted(
		signStateConfirmTimeout,
		keep,
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

	err = keep.SubmitSignature(signature)
	if err != nil {
		t.Fatal(err)
	}

	deduplicator.NotifySigningCompleted(keepID1, digest)

	//
	// request signature with the same digest one more time - should work
	//
	err = chain.RequestSignature(keepAddress1, digest)
	if err != nil {
		t.Fatal(err)
	}

	canSign, err = deduplicator.NotifySigningStarted(
		signStateConfirmTimeout,
		keep,
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
	registry.AddSigner(keepID1)

	canClose := deduplicator.NotifyClosingStarted(keepID1)
	if !canClose {
		t.Fatal("should be allowed to close a keep")
	}
}

func TestDoNotCloseIfCurrentlyClosing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, registry, _ := newDeduplicator(ctx)
	registry.AddSigner(keepID1)

	deduplicator.NotifyClosingStarted(keepID1)

	canClose := deduplicator.NotifyClosingStarted(keepID1)
	if canClose {
		t.Fatal("should not be allowed to close a keep")
	}
}

func TestDoNotCloseIfAlreadyClosed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, _, _ := newDeduplicator(ctx)
	// keep not in the registry

	canClose := deduplicator.NotifyClosingStarted(keepID1)
	if canClose {
		t.Fatal("should not be allowed to close a keep")
	}
}

func TestDoTerminate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, registry, _ := newDeduplicator(ctx)
	registry.AddSigner(keepID1)

	canTerminate := deduplicator.NotifyTerminatingStarted(keepID1)
	if !canTerminate {
		t.Fatal("should be allowed to terminate a keep")
	}
}

func TestDoNotTerminateIfCurrentlyTerminating(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, registry, _ := newDeduplicator(ctx)
	registry.AddSigner(keepID1)

	deduplicator.NotifyTerminatingStarted(keepID1)

	canTerminate := deduplicator.NotifyTerminatingStarted(keepID1)
	if canTerminate {
		t.Fatal("should not be allowed to terminate a keep")
	}
}

func TestDoNotTerminateIfAlreadyTerminated(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deduplicator, _, _ := newDeduplicator(ctx)
	// keep not in the registry

	canTerminate := deduplicator.NotifyTerminatingStarted(keepID1)
	if canTerminate {
		t.Fatal("should not be allowed to terminate a keep")
	}
}

func newDeduplicator(ctx context.Context) (
	*Deduplicator,
	*mockRegistry,
	local.Chain,

) {
	mockRegistry := &mockRegistry{
		keeps: make(map[chain.ID]bool),
	}

	chain := local.Connect(ctx)

	deduplicator := NewDeduplicator(
		mockRegistry,
		chain,
	)

	return deduplicator, mockRegistry, chain
}

type mockRegistry struct {
	keeps map[chain.ID]bool
}

func (mr *mockRegistry) AddSigner(keepID chain.ID) {
	mr.keeps[keepID] = true
}

func (mr *mockRegistry) HasSigner(keepID chain.ID) bool {
	return mr.keeps[keepID]
}
