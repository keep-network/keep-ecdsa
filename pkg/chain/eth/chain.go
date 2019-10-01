// Package eth contains interface for interaction with an ethereum blockchain
// along with structures reflecting events emitted on an ethereum blockchain.
package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

// KeepAddress is a keep contract address.
type KeepAddress = common.Address

// Handle represents a handle to an ethereum blockchain.
type Handle interface {
	// Address returns client's ethereum address.
	Address() common.Address

	ECDSAKeepFactory
	ECDSAKeep
}

// ECDSAKeepFactory is an interface that provides ability to interact with
// ECDSAKeepFactory ethereum contracts.
type ECDSAKeepFactory interface {
	// OnECDSAKeepCreated is a callback that is invoked when an on-chain
	// notification of a new ECDSA keep creation is seen.
	OnECDSAKeepCreated(
		handler func(event *ECDSAKeepCreatedEvent),
	) (subscription.EventSubscription, error)
}

// ECDSAKeep is an interface that provides ability to interact with ECDSAKeep
// ethereum contracts.
type ECDSAKeep interface {
	// OnSignatureRequested is a callback that is invoked when an on-chain
	// notification of a new signing request for a given keep is seen.
	OnSignatureRequested(
		keepAddress KeepAddress,
		handler func(event *SignatureRequestedEvent),
	) (subscription.EventSubscription, error)

	// SubmitKeepPublicKey submits a 64-byte serialized public key to a keep
	// contract deployed under a given address.
	SubmitKeepPublicKey(keepAddress KeepAddress, publicKey [64]byte) error // TODO: Add promise *async.KeepPublicKeySubmissionPromise

	// SubmitSignature submits a signature to a keep contract deployed under a
	// given address.
	SubmitSignature(
		keepAddress KeepAddress,
		digest [32]byte,
		signature *ecdsa.Signature,
	) error // TODO: Add promise *async.SignatureSubmissionPromise
}
