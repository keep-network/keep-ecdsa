package chain

import (
	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-tecdsa/pkg/eth/event"
)

// Interface is an interface that provides ability to interact with ethereum
// contracts.
type Interface interface {
	OnGroupRequested(
		func(request *event.GroupRequested),
	) (subscription.EventSubscription, error)
}
