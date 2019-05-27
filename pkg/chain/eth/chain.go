// Package eth contains interface for interaction with an ethereum blockchain
// along with structures reflecting events emitted on an ethereum blockchain.
package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-core/pkg/subscription"
)

// KeepAddress is a keep contract address.
type KeepAddress *common.Address

// KeepPublicKey is a public key of a signer related to the keep.
// It is concatenation of X and Y coordinates each of 32-byte length. Value
// shorter than 32-byte should be preceded with zeros.
type KeepPublicKey [64]byte

// Interface is an interface that provides ability to interact with ethereum
// contracts.
type Interface interface {
	// OnECDSAKeepCreated is a callback that is invoked when an on-chain
	// notification of a new ECDSA keep creation is seen.
	OnECDSAKeepCreated(
		func(request *ECDSAKeepCreatedEvent),
	) (subscription.EventSubscription, error)

	// SubmitKeepPublicKey submits a public key to a keep contract deployed under
	// a given address.
	SubmitKeepPublicKey(address KeepAddress, publicKey KeepPublicKey) error // TODO: Add promise *async.KeepPublicKeySubmissionPromise
}
