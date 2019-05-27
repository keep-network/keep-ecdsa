// Package eth contains interface for interaction with an ethereum blockchain
// along with structures reflecting events emitted on an ethereum blockchain.
package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-core/pkg/subscription"
)

// Handle is an interface that provides ability to interact with the blockchain.
type Handle interface {
	// ECDSAKeepFactory returns an interface for ECDSA Keep Factory contract.
	ECDSAKeepFactory() (ECDSAKeepFactory, error)
	// ECDSAKeep returns an interface for ECDSA Keep contract of given address.
	ECDSAKeep(address *common.Address) (ECDSAKeep, error)
}

// ECDSAKeepFactory is an interface that provides ability to interact with
// ECDSA Keep Factory contract deployed on the blockchain.
type ECDSAKeepFactory interface {
	// OnECDSAKeepCreated is a callback that is invoked when an on-chain
	// notification of a new ECDSA keep creation is seen.
	OnECDSAKeepCreated(
		func(request *ECDSAKeepCreatedEvent),
	) (subscription.EventSubscription, error)
}

// ECDSAKeep is an interface that provides ability to interact with ECDSA Keep
// contract deployed on the blockchain.
type ECDSAKeep interface {
	SubmitKeepPublicKey(publicKey [64]byte) error // TODO: Add promise *async.KeepPublicKeySubmissionPromise
}
