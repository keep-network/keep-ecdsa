// Package eth contains interface for interaction with an ethereum blockchain
// along with structures reflecting events emitted on an ethereum blockchain.
package eth // TODO: rename; this can be any host chain

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
	// BalanceMonitor returns a balance monitor.
	BalanceMonitor() (chain.BalanceMonitor, error)
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
	BondedECDSAKeep
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

	// GetKeepAtIndex returns the address of the keep at the given index.
	GetKeepAtIndex(keepIndex *big.Int) (common.Address, error)
}

// BondedECDSAKeep is an interface that provides ability to interact with
// BondedECDSAKeep ethereum contracts.
type BondedECDSAKeep interface {
	// OnSignatureRequested installs a callback that is invoked when an on-chain
	// notification of a new signing request for a given keep is seen.
	OnSignatureRequested(
		keepAddress common.Address,
		handler func(event *SignatureRequestedEvent),
	) (subscription.EventSubscription, error)

	// OnConflictingPublicKeySubmitted installs a callback that is invoked upon
	// notification of mismatched public keys that were submitted by keep members.
	OnConflictingPublicKeySubmitted(
		keepAddress common.Address,
		handler func(event *ConflictingPublicKeySubmittedEvent),
	) (subscription.EventSubscription, error)

	// OnPublicKeyPublished installs a callback that is invoked upon
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

	// OnKeepClosed installs a callback that will be called on closing the
	// given keep.
	OnKeepClosed(
		keepAddress common.Address,
		handler func(event *KeepClosedEvent),
	) (subscription.EventSubscription, error)

	// OnKeepTerminated installs a callback that will be called on terminating
	// the given keep.
	OnKeepTerminated(
		keepAddress common.Address,
		handler func(event *KeepTerminatedEvent),
	) (subscription.EventSubscription, error)

	// IsAwaitingSignature checks if the keep is waiting for a signature to be
	// calculated for the given digest.
	IsAwaitingSignature(keepAddress common.Address, digest [32]byte) (bool, error)

	// IsActive checks if the keep with the given address is active and responds
	// to signing request. This function returns false only for closed keeps.
	IsActive(keepAddress common.Address) (bool, error)

	// LatestDigest returns the latest digest requested to be signed.
	LatestDigest(keepAddress common.Address) ([32]byte, error)

	// SignatureRequestedBlock returns block number from the moment when a
	// signature was requested for the given digest from a keep.
	// If a signature was not requested for the given digest, returns 0.
	SignatureRequestedBlock(keepAddress common.Address, digest [32]byte) (uint64, error)

	// GetPublicKey returns keep's public key. If there is no public key yet,
	// an empty slice is returned.
	GetPublicKey(keepAddress common.Address) ([]uint8, error)

	// GetMembers returns keep's members.
	GetMembers(keepAddress common.Address) ([]common.Address, error)

	// GetHonestThreshold returns keep's honest threshold.
	GetHonestThreshold(keepAddress common.Address) (uint64, error)

	// GetOpenedTimestamp returns timestamp when the keep was created.
	GetOpenedTimestamp(keepAddress common.Address) (time.Time, error)

	// PastSignatureSubmittedEvents returns all signature submitted events
	// for the given keep which occurred after the provided start block.
	// All implementations should returns those events sorted by the
	// block number in the ascending order.
	PastSignatureSubmittedEvents(
		keepAddress string,
		startBlock uint64,
	) ([]*SignatureSubmittedEvent, error)
}
