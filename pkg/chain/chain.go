// Package chain contains interface for interaction with a blockchain that
// implements ECDSA keep functionality, along with structures reflecting events
// needed for that functionality.
package chain

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
)

// Handle represents a handle to an ethereum blockchain.
type Handle interface {
	// Address returns client's ethereum address.
	Address() common.Address // TODO: use implementation-agnostic type
	// StakeMonitor returns a stake monitor.
	StakeMonitor() (chain.StakeMonitor, error)
	// BlockCounter returns a block counter.
	BlockCounter() chain.BlockCounter
	// Signing returns a signer interface allowing to sign and verify messages
	// using the chain implementation-specific mechanism as well as to
	// convert between public key and address.
	Signing() chain.Signing
	// BlockTimestamp returns given block's timestamp.
	// In case the block is not yet mined, an error should be returned.
	BlockTimestamp(blockNumber *big.Int) (uint64, error)

	BondedECDSAKeepFactory
}

// BondedECDSAKeepFactory is an interface that provides ability to interact with
// BondedECDSAKeepFactory ethereum contracts.
type BondedECDSAKeepFactory interface {
	// RegisterAsMemberCandidate registers client as a candidate to be selected
	// to a keep.
	RegisterAsMemberCandidate(application common.Address) error

	// OnBondedECDSAKeepCreated installs a callback that is invoked when an
	// on-chain notification of a new bonded ECDSA keep creation is seen.
	OnBondedECDSAKeepCreated(
		handler func(event *BondedECDSAKeepCreatedEvent),
	) subscription.EventSubscription

	// IsRegisteredForApplication checks if the operator is registered
	// as a signer candidate in the factory for the given application.
	IsRegisteredForApplication(application common.Address) (bool, error)

	// IsEligibleForApplication checks if the operator is eligible to register
	// as a signer candidate for the given application.
	IsEligibleForApplication(application common.Address) (bool, error)

	// IsStatusUpToDateForApplication checks if the operator's status
	// is up to date in the signers' pool of the given application.
	IsStatusUpToDateForApplication(application common.Address) (bool, error)

	// UpdateStatusForApplication updates the operator's status in the signers'
	// pool for the given application.
	UpdateStatusForApplication(application common.Address) error

	// IsOperatorAuthorized checks if the factory has the authorization to
	// operate on stake represented by the provided operator.
	IsOperatorAuthorized(operator common.Address) (bool, error)

	// GetKeepCount returns number of keeps.
	GetKeepCount() (*big.Int, error)

	// GetKeepAtIndex returns a handle to the keep at the given index.
	GetKeepAtIndex(keepIndex *big.Int) (BondedECDSAKeepHandle, error)
	// GetKeepWithID returns a handle to the keep with the given ID.
	// FIXME currently this ID is still a common.Address 😬
	GetKeepWithID(keepID common.Address) (BondedECDSAKeepHandle, error)
}

// BondedECDSAKeepHandle is an interface that provides ability to interact with
// a single bonded ECDSA keep's on-chain component. A bonded ECDSA keep is a
// threshold signing group that has a corresponding bond amount securing its
// honest cooperation in the threshold signature application that the keep
// corresponds to.
type BondedECDSAKeepHandle interface {
	// GetKeepWithID returns the underlying keep's ID.
	// FIXME currently this ID is still a common.Address 😬
	ID() common.Address

	// OnSignatureRequested installs a callback that is invoked when an on-chain
	// notification of a new signing request for a given keep is seen.
	OnSignatureRequested(
		handler func(event *SignatureRequestedEvent),
	) (subscription.EventSubscription, error)

	// OnConflictingPublicKeySubmitted installs a callback that is invoked upon
	// notification of mismatched public keys that were submitted by keep members.
	OnConflictingPublicKeySubmitted(
		handler func(event *ConflictingPublicKeySubmittedEvent),
	) (subscription.EventSubscription, error)

	// OnPublicKeyPublished installs a callback that is invoked upon
	// notification of a published public key, which means that all members have
	// submitted the same key.
	OnPublicKeyPublished(
		handler func(event *PublicKeyPublishedEvent),
	) (subscription.EventSubscription, error)

	// SubmitKeepPublicKey submits a 64-byte serialized public key to a keep
	// contract deployed under a given address.
	SubmitKeepPublicKey(publicKey [64]byte) error

	// SubmitSignature submits a signature to a keep contract deployed under a
	// given address.
	SubmitSignature(signature *ecdsa.Signature) error

	// OnKeepClosed installs a callback that will be called on closing the
	// given keep.
	OnKeepClosed(
		handler func(event *KeepClosedEvent),
	) (subscription.EventSubscription, error)

	// OnKeepTerminated installs a callback that will be called on terminating
	// the given keep.
	OnKeepTerminated(
		handler func(event *KeepTerminatedEvent),
	) (subscription.EventSubscription, error)

	// IsAwaitingSignature checks if the keep is waiting for a signature to be
	// calculated for the given digest.
	IsAwaitingSignature(digest [32]byte) (bool, error)

	// IsActive checks if the keep with the given address is active and responds
	// to signing request. This function returns false only for closed keeps.
	IsActive() (bool, error)

	// LatestDigest returns the latest digest requested to be signed.
	LatestDigest() ([32]byte, error)

	// SignatureRequestedBlock returns block number from the moment when a
	// signature was requested for the given digest from a keep.
	// If a signature was not requested for the given digest, returns 0.
	SignatureRequestedBlock(digest [32]byte) (uint64, error)

	// GetPublicKey returns keep's public key. If there is no public key yet,
	// an empty slice is returned.
	GetPublicKey() ([]uint8, error)

	// GetMembers returns keep's members.
	GetMembers() ([]common.Address, error)

	// GetHonestThreshold returns keep's honest threshold.
	GetHonestThreshold() (uint64, error)

	// GetOpenedTimestamp returns timestamp when the keep was created.
	GetOpenedTimestamp() (time.Time, error)

	// PastSignatureSubmittedEvents returns all signature submitted events
	// for the given keep which occurred after the provided start block.
	// All implementations should returns those events sorted by the
	// block number in the ascending order.
	PastSignatureSubmittedEvents(
		startBlock uint64,
	) ([]*SignatureSubmittedEvent, error)
}
