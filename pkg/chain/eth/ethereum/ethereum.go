// Package ethereum contains implementation of ethereum chain interface.
package ethereum

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/gen/abi"
	"github.com/keep-network/keep-tecdsa/pkg/sign"
)

// OnECDSAKeepCreated is a callback that is invoked when an on-chain
// notification of a new ECDSA keep creation is seen.
func (ec *EthereumChain) OnECDSAKeepCreated(
	handler func(event *eth.ECDSAKeepCreatedEvent),
) (subscription.EventSubscription, error) {
	return ec.watchECDSAKeepCreated(
		func(
			chainEvent *abi.ECDSAKeepFactoryECDSAKeepCreated,
		) {
			handler(&eth.ECDSAKeepCreatedEvent{
				KeepAddress: chainEvent.KeepAddress,
			})
		},
		func(err error) error {
			return fmt.Errorf("keep created callback failed: [%s]", err)
		},
	)
}

// OnSignatureRequested is a callback that is invoked on-chain
// when a keep's signature is requested.
func (ec *EthereumChain) OnSignatureRequested(
	keepAddress eth.KeepAddress,
	handler func(event *eth.SignatureRequestedEvent),
) (subscription.EventSubscription, error) {
	keepContract, err := ec.getKeepContract(keepAddress)
	if err != nil {
		return nil, fmt.Errorf("could not create contract ABI: [%v]", err)
	}

	return ec.watchSignatureRequested(
		keepContract,
		func(
			chainEvent *abi.ECDSAKeepSignatureRequested,
		) {
			handler(&eth.SignatureRequestedEvent{
				Digest: chainEvent.Digest,
			})
		},
		func(err error) error {
			return fmt.Errorf("keep signature requested callback failed: [%s]", err)
		},
	)
}

// SubmitKeepPublicKey submits a public key to a keep contract deployed under
// a given address.
func (ec *EthereumChain) SubmitKeepPublicKey(
	keepAddress eth.KeepAddress,
	publicKey [64]byte,
) error {
	keepContract, err := ec.getKeepContract(keepAddress)
	if err != nil {
		return err
	}

	transaction, err := keepContract.SetPublicKey(ec.transactorOptions, publicKey[:])
	if err != nil {
		return err
	}

	fmt.Printf("Transaction submitted with hash: [%x]\n", transaction.Hash())

	return nil
}

func (ec *EthereumChain) getKeepContract(address common.Address) (*abi.ECDSAKeep, error) {
	ecdsaKeepContract, err := abi.NewECDSAKeep(address, ec.client)
	if err != nil {
		return nil, err
	}

	return ecdsaKeepContract, nil
}

// SubmitSignature submits a signature to a keep contract deployed under a
// given address.
func (ec *EthereumChain) SubmitSignature(
	keepAddress eth.KeepAddress,
	digest [32]byte,
	signature *sign.Signature,
) error {
	keepContract, err := ec.getKeepContract(keepAddress)
	if err != nil {
		return err
	}

	transaction, err := keepContract.SubmitSignature(
		ec.transactorOptions,
		digest,
		signature.R.Bytes(),
		signature.S.Bytes(),
	)
	if err != nil {
		return err
	}

	fmt.Printf("Transaction submitted with hash: [%x]\n", transaction.Hash())

	return nil
}
