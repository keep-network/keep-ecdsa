// Package ethereum contains implementation of ethereum chain interface.
package ethereum

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
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

// OnSignatureRequested is a callback that is invoked on-chain
// when a keep's signature is requested.
func (ec *EthereumChain) OnSignatureRequested(
	keepAddress common.Address,
	handle func(event *eth.SignatureRequestedEvent),
) (subscription.EventSubscription, error) {
	keep, err := abi.NewECDSAKeep(
		keepAddress,
		ec.client,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create contract ABI: [%v]", err)
	}

	return ec.watchSignatureRequested(
		keep,
		func(
			chainEvent *abi.ECDSAKeepSignatureRequested,
		) {
			handle(&eth.SignatureRequestedEvent{
				Digest: chainEvent.Digest,
			})
		},
		func(err error) error {
			return fmt.Errorf("keep requested callback failed: [%s]", err)
		},
	)
}
