// Package eth contains interface for interaction with an ethereum blockchain
// along with structures reflecting events emitted on an ethereum blockchain.
package eth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-core/pkg/subscription"
)

// Interface is an interface that provides ability to interact with ethereum
// contracts.
type Interface interface {
	// OnECDSAKeepCreated is a callback that is invoked when an on-chain
	// notification of a new ECDSA keep creation is seen.
	OnECDSAKeepCreated(
		callback func(request *ECDSAKeepCreatedEvent),
	) (subscription.EventSubscription, error)

	// OnSignatureRequested is a callback that is invoked when an on-chain
	// notification of a new signing request is seen.
	OnSignatureRequested(
		keepAddress common.Address,
		callback func(request *SignatureRequestedEvent),
	) (subscription.EventSubscription, error)
}
