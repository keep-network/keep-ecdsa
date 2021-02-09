// Package chain contains interface for interaction with a host chain along with
// structures reflecting events emitted on such a chain.
package chain // TODO: rename; this can be any host chain

import (
	cecdsa "crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
)

type KeepMemberID interface {
	fmt.Stringer

	// KeepMemberIDs should be convertable to their associated OperatorID. This
	// may be a 1-to-1 correspondence, i.e. a simple interface cast, or it may
	// be an internally-tracked association on the host chain.
	OperatorID() OperatorID
}

type OperatorID interface {
	fmt.Stringer

	// PublicKey returns the public key associated with this operator.
	PublicKey() cecdsa.PublicKey
	// NetworkID returns the network id associated with this operator.
	NetworkID() net.ID
}

// Should be chain-specific, unique across chains, and simple ([A-Za-z0-9-_.]).
// This can generally mean prefixing the common name of the chain followed by a
// dash followed by the chain-specific identifier (e.g., the contract address
// for Ethereum). For legacy reasons, Ethereum KeepIDs should always be the bare
// Keep contract address.
//
// Should not be expected to be comparable between types or amongst themselves.
type KeepID interface {
	fmt.Stringer
}

// package chain in keep-core -- or keep-common/chain?
// type BaseHandle interface {
// 	BlockCounter() (BlockCounter, error)
// 	StakeMonitor() (StakeMonitor, error)
// 	BalanceMonitor() (BalanceMonitor, error)
// }

// FIXME BalanceMonitor may be an Ethereum-specific concept, kick it
// FIXME from the chain interface?

// Handle represents a handle to an implementation of operator functions
// for a particular chain.
type Handle interface {
	// chain.BaseHandle from keep-core
	// StakeMonitor returns a stake monitor.
	StakeMonitor() (chain.StakeMonitor, error)
	// BalanceMonitor returns a balance monitor.
	BalanceMonitor() (chain.BalanceMonitor, error)
	// BlockCounter returns a block counter.
	BlockCounter() (chain.BlockCounter, error)
	// Signing returns a signer interface allowing to sign and verify messages
	// using the chain implementation-specific mechanism as well as to
	// convert between public key and address.
	Signing() chain.Signing
	// BlockTimestamp returns given block's timestamp.
	// In case the block is not yet mined, an error should be returned.
	BlockTimestamp(blockNumber *big.Int) (uint64, error)

	// BondedECDSAKeepManager returns a handle to the on-chain component used to
	// interact with bonded ECDSA keeps.
	BondedECDSAKeepManager() (BondedECDSAKeepManager, error)
}

// BondedECDSAKeepManager is an interface whose implementations allow for
// managing, monitoring, and creating bonded ECDSA keeps as well as the
// operator's eligibility for participation in such keeps.
type BondedECDSAKeepManager interface {
	// TBTCApplicationHandle returns a handle for interacting with the tBTC
	// application associated with this BondedECDSAKeepManager. Returns nil with
	// an error if no tBTC application exists for this manager.
	TBTCApplicationHandle() (BondedECDSAKeepApplicationHandle, error)

	// OnBondedECDSAKeepCreated installs a callback that is invoked when an
	// on-chain notification of a new bonded ECDSA keep creation is seen.
	OnBondedECDSAKeepCreated(
		handler func(event *BondedECDSAKeepCreatedEvent),
	) subscription.EventSubscription

	// GetKeepCount returns number of keeps.
	GetKeepCount() (*big.Int, error)

	// GetKeepAtIndex returns a handle to the keep at the given index.
	GetKeepAtIndex(keepIndex *big.Int) (BondedECDSAKeepHandle, error)
	// GetKeepAtIndex returns a handle to the keep with the given identifier.
	GetKeepWithID(keepID KeepID) (BondedECDSAKeepHandle, error)
}

type BondedECDSAKeepApplicationHandle interface {
	// RegisterAsMemberCandidate registers client as a candidate to be selected
	// to a keep.
	RegisterAsMemberCandidate() error

	// IsRegisteredForApplication checks if the operator is registered
	// as a signer candidate in the factory for the given application.
	IsRegisteredForApplication() (bool, error)

	// IsEligibleForApplication checks if the operator is eligible to register
	// as a signer candidate for the given application.
	IsEligibleForApplication() (bool, error)

	// IsStatusUpToDateForApplication checks if the operator's status
	// is up to date in the signers' pool of the given application.
	IsStatusUpToDateForApplication() (bool, error)

	// UpdateStatusForApplication updates the operator's status in the signers'
	// pool for the given application.
	UpdateStatusForApplication() error

	// IsOperatorAuthorized checks if the factory has the authorization to
	// operate on stake represented by the provided operator.
	IsOperatorAuthorized(operator OperatorID) (bool, error)
}

// BondedECDSAKeepHandle is an interface that provides ability to interact with
// a single BondedECDSAKeep ethereum contract.
type BondedECDSAKeepHandle interface {
	// ID should return the id of this keep in a host chain-agnostic format.
	ID() KeepID

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
	SubmitKeepPublicKey(publicKey [64]byte) error // TODO: Add promise *async.KeepPublicKeySubmissionPromise

	// SubmitSignature submits a signature to a keep contract deployed under a
	// given address.
	SubmitSignature(
		signature *ecdsa.Signature,
	) error // TODO: Add promise *async.SignatureSubmissionPromise

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
	// FIXME this should not be needed; instead, BelongsToKeep() or similar
	GetMembers() ([]KeepMemberID, error)

	// IsThisOperatorMember returns true if the current operator belongs to the
	// BondedECDSAKeep represented by this handle, false otherwise, or an error
	// if the process of determining this fails.
	//
	// FIXME IsOperatorMember
	IsThisOperatorMember() (bool, error)

	// OperatorIndex returns the index of the current operator in this keep's
	// set of members, or an error if the process of determining this fails. If
	// the operator is not a member this will return -1 (and no error) and
	// IsOperatorMember will return false.
	OperatorIndex() (int, error)

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
