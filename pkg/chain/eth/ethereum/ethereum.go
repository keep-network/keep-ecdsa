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

			// TODO(liamz): temporary for now
			ec.registerKeepContract(chainEvent.KeepAddress)
		},
		func(err error) error {
			return fmt.Errorf("keep requested callback failed: [%s]", err)
		},
	)
}

// liamz: ecdsaKeepFactoryContractAddress or address
// liamz: registerECDSAKeepContract or this?
func (ec *EthereumChain) registerKeepContract(address common.Address) error {
	ecdsaKeepContract, err := abi.NewECDSAKeep(
		address,
		ec.client,
	)
	if err != nil {
		return err
	}

	ec.keepContracts[address] = ecdsaKeepContract

	ec.watchECDSAKeepSignatureRequested(
		ecdsaKeepContract,
		func(
			chainEvent *abi.ECDSAKeepSignatureRequested,
		) {
			// handle(&eth.ECDSAKeepSignatureRequestedEvent{
			// 	Digest: chainEvent.Digest,
			// })
			fmt.Printf("signature requested: [%v]", chainEvent.Digest)
		},
		func(err error) error {
			return fmt.Errorf("keep requested callback failed: [%s]", err)
		},
	)

	return nil
}

func (ec *EthereumChain) OnECDSAKeepSignatureRequested(
	handle func(event *eth.ECDSAKeepSignatureRequestedEvent),
) (subscription.EventSubscription, error) {
	return nil, nil

	// return ec.watchECDSAKeepSignatureRequested(
	// 	func(
	// 		chainEvent *abi.ECDSAKeep,
	// 	) {
	// 		handle(&eth.ECDSAKeepSignatureRequestedEvent{
	// 			Digest: chainEvent.Digest,
	// 		})
	// 	},
	// 	func(err error) error {
	// 		return fmt.Errorf("keep requested callback failed: [%s]", err)
	// 	},
	// )
}
