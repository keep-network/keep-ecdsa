package chain

import (
	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-tecdsa/pkg/eth/event"
)

// Interface is an interface that provides ability to interact with ethereum
// contracts.
type Interface interface {
	OnGroupRequested( // TODO: Remove it, it's replaced by OnECDSAKeepRequested
		func(request *event.GroupRequested),
	) (subscription.EventSubscription, error)

	// OnECDSAKeepRequested is a callback that is invoked when an on-chain
	// notification of a new ECDSA keep request is seen.
	OnECDSAKeepRequested(
		func(request *event.ECDSAKeepRequested),
	) (subscription.EventSubscription, error)
}
