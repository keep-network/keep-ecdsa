// Package ethereum contains implementation of ethereum chain interface.
package ethereum

import (
	"fmt"

	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/gen/abi"
)

// OnECDSAKeepCreated is a callback that is invoked when an on-chain
// notification of a new ECDSA keep creation is seen.
func (ec *EthereumChain) OnECDSAKeepCreated(
	handle func(groupRequested *eth.ECDSAKeepCreatedEvent),
) (subscription.EventSubscription, error) {
	return ec.watchECDSAKeepCreated(
		func(
			chainEvent *abi.ECDSAKeepFactoryECDSAKeepCreated,
		) {
			handle(&eth.ECDSAKeepCreatedEvent{
				KeepAddress: chainEvent.KeepAddress,
			})
		},
		func(err error) error {
			return fmt.Errorf("keep requested callback failed: [%s]", err)
		},
	)
}
