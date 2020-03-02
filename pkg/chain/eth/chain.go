// Package eth contains interface for interaction with an ethereum blockchain
// along with structures reflecting events emitted on an ethereum blockchain.
package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

// Handle represents a handle to an ethereum blockchain.
type Handle interface {
	// Address returns client's ethereum address.
	Address() common.Address
	// StakeMonitor returns a stake monitor.
	StakeMonitor() (chain.StakeMonitor, error)

	BondedECDSAKeepFactory
	BondedECDSAKeep
}

// BondedECDSAKeepFactory is an interface that provides ability to interact with
// BondedECDSAKeepFactory ethereum contracts.
type BondedECDSAKeepFactory interface {
	// RegisterAsMemberCandidate registers client as a candidate to be selected
	// to a keep.
	RegisterAsMemberCandidate(application common.Address) error

	// OnBondedECDSAKeepCreated is a callback that is invoked when an on-chain
	// notification of a new bonded ECDSA keep creation is seen.
	OnBondedECDSAKeepCreated(
		handler func(event *BondedECDSAKeepCreatedEvent),
	) (subscription.EventSubscription, error)
}

// BondedECDSAKeep is an interface that provides ability to interact with BondedECDSAKeep
// ethereum contracts.
type BondedECDSAKeep interface {
	// OnSignatureRequested is a callback that is invoked when an on-chain
	// notification of a new signing request for a given keep is seen.
	OnSignatureRequested(
		keepAddress common.Address,
		handler func(event *SignatureRequestedEvent),
	) (subscription.EventSubscription, error)

	// OnConflictingPublicKeySubmitted is a callback that is invoked upon
	// notification of mismatched public keys that were submitted by keep members.
	OnConflictingPublicKeySubmitted(
		keepAddress common.Address,
		handler func(event *ConflictingPublicKeySubmittedEvent),
	) (subscription.EventSubscription, error)

	// OnPublicKeyPublished is a callback that is invoked upon
	// notification of a published public key, which means that all members have
	// submitted the same key.
	OnPublicKeyPublished(
		keepAddress common.Address,
		handler func(event *PublicKeyPublishedEvent),
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

	OnKeepClosed(
		keepAddress common.Address,
		handler func(event *KeepClosedEvent),
	) (subscription.EventSubscription, error)

	// IsAwaitingSignature checks if the keep is waiting for a signature to be
	// calculated for the given digest.
	IsAwaitingSignature(keepAddress common.Address, digest [32]byte) (bool, error)
}
