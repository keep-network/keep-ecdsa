// Package ethereum contains implementation of ethereum chain interface.
package ethereum

import (
	"fmt"

	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-tecdsa/pkg/eth/chain/gen/abi"
	"github.com/keep-network/keep-tecdsa/pkg/eth/event"
)

// OnGroupRequested registers a callback that is invoked when an on-chain
// notification of a new group requested.
func (ec *EthereumChain) OnGroupRequested(
	handle func(groupRequested *event.GroupRequested),
) (subscription.EventSubscription, error) {
	return ec.watchGroupRequested(
		func(
			chainEvent *abi.KeepTECDSAGroupGroupRequested,
		) {
			handle(&event.GroupRequested{
				RequestID:          chainEvent.RequestID,
				GroupID:            chainEvent.GroupID,
				GroupSize:          chainEvent.GroupSize,
				DishonestThreshold: chainEvent.DishonestThreshold,
			})
		},
		func(err error) error {
			return fmt.Errorf("group requested callback failed: [%s]", err)
		},
	)
}
