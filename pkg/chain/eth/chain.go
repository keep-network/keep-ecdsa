// Package eth contains interface for interaction with an ethereum blockchain
// along with structures reflecting events emitted on an ethereum blockchain.
package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

// Handle represents a handle to an ethereum blockchain.
type Handle interface {
	// Address returns client's ethereum address.
	Address() common.Address

	BlockCounter

	ECDSAKeepFactory
	ECDSAKeep
}

// TODO: This duplicates BlockCounter from keep-core, we need to merge them.
type BlockCounter interface {
	// WatchBlocks returns a channel that will emit new block numbers as they
	// are mined. When the context provided as the parameter ends, new blocks
	// are no longer pushed to the channel and the channel is closed. If there
	// is no reader for the channel or reader is too slow, block updates can be
	// dropped.
	WatchBlocks(ctx context.Context) <-chan uint64
}

// ECDSAKeepFactory is an interface that provides ability to interact with
// ECDSAKeepFactory ethereum contracts.
type ECDSAKeepFactory interface { // TODO: Rename to BondedECDSAKeepFactory
	// IsRegistered checks if client is already registered as a member candidate
	// in the factory for the given application.
	IsRegistered(application common.Address) (bool, error)

	// EligibleStake returns client's current value of token stake balance for
	// the factory.
	EligibleStake() (*big.Int, error)

	// RegisterAsMemberCandidate registers client as a candidate to be selected
	// to a keep.
	RegisterAsMemberCandidate(application common.Address) error

	// OnECDSAKeepCreated is a callback that is invoked when an on-chain
	// notification of a new ECDSA keep creation is seen.
	OnECDSAKeepCreated(
		handler func(event *ECDSAKeepCreatedEvent),
	) (subscription.EventSubscription, error)
}

// ECDSAKeep is an interface that provides ability to interact with ECDSAKeep
// ethereum contracts.
type ECDSAKeep interface { // TODO: Rename to BondedECDSAKeep
	// OnSignatureRequested is a callback that is invoked when an on-chain
	// notification of a new signing request for a given keep is seen.
	OnSignatureRequested(
		keepAddress common.Address,
		handler func(event *SignatureRequestedEvent),
	) (subscription.EventSubscription, error)

	// SubmitKeepPublicKey submits a 64-byte serialized public key to a keep
	// contract deployed under a given address.
	SubmitKeepPublicKey(keepAddress common.Address, publicKey [64]byte) error // TODO: Add promise *async.KeepPublicKeySubmissionPromise

	// SubmitSignature submits a signature to a keep contract deployed under a
	// given address.
	SubmitSignature(
		keepAddress common.Address,
		signature *ecdsa.Signature,
	) error // TODO: Add promise *async.SignatureSubmissionPromise
}
