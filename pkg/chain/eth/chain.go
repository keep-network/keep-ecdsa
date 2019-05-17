// Package eth contains interface for interaction with an ethereum blockchain
// along with structures reflecting events emitted on an ethereum blockchain.
package eth

import (
	"github.com/keep-network/keep-core/pkg/subscription"
)

// Interface is an interface that provides ability to interact with ethereum
// contracts.
type Interface interface {
	// OnECDSAKeepRequested is a callback that is invoked when an on-chain
	// notification of a new ECDSA keep request is seen.
	OnECDSAKeepRequested(
		func(request *ECDSAKeepRequestedEvent),
	) (subscription.EventSubscription, error)
}
