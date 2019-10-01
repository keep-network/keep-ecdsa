// Package ethereum contains implementation of ethereum chain interface.
package ethereum

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/gen/abi"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-tecdsa/pkg/utils/byteutils"
)

var logger = log.Logger("keep-chain-eth-ethereum")

// Address returns client's ethereum address.
func (ec *EthereumChain) Address() common.Address {
	return ec.transactorOptions.From
}

// RegisterCandidate registers client as a candidate to be selected to a keep.
func (ec *EthereumChain) RegisterCandidate() error {
	transaction, err := ec.ecdsaKeepFactoryContract.RegisterCandidate(ec.transactorOptions)
	if err != nil {
		return err
	}

	logger.Debugf("submitted RegisterCandidate transaction with hash: [%x]", transaction.Hash())

	return nil
}

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
				Members:     chainEvent.Members,
			})
		},
		func(err error) error {
			return fmt.Errorf("keep created callback failed: [%v]", err)
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
		return nil, fmt.Errorf("failed to create contract abi: [%v]", err)
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
			return fmt.Errorf("keep signature requested callback failed: [%v]", err)
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

	logger.Debugf("submitted SetPublicKey transaction with hash: [%x]", transaction.Hash())

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
	signature *ecdsa.Signature,
) error {
	keepContract, err := ec.getKeepContract(keepAddress)
	if err != nil {
		return err
	}

	signatureR, err := byteutils.BytesTo32Byte(signature.R.Bytes())
	if err != nil {
		return err
	}

	signatureS, err := byteutils.BytesTo32Byte(signature.S.Bytes())
	if err != nil {
		return err
	}

	transaction, err := keepContract.SubmitSignature(
		ec.transactorOptions,
		digest,
		signatureR,
		signatureS,
		uint8(signature.RecoveryID),
	)
	if err != nil {
		return err
	}

	logger.Debugf("submitted SubmitSignature transaction with hash: [%x]", transaction.Hash())

	return nil
}
