// Package ethereum contains implementation of ethereum chain interface.
package ethereum

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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

// IsRegistered checks if client is already registered as a member candidate in
// the factory for the given application.
func (ec *EthereumChain) IsRegistered(application common.Address) (bool, error) {
	return ec.ecdsaKeepFactoryContract.ECDSAKeepFactoryCaller.IsOperatorRegistered(
		ec.callerOptions,
		ec.Address(),
		application,
	)
}

// EligibleStake returns client's current value of token stake balance for the
// factory.
func (ec *EthereumChain) EligibleStake() (*big.Int, error) {
	return ec.ecdsaKeepFactoryContract.ECDSAKeepFactoryCaller.EligibleStake(
		ec.callerOptions,
		ec.Address(),
	)
}

// RegisterAsMemberCandidate registers client as a candidate to be selected
// to a keep.
func (ec *EthereumChain) RegisterAsMemberCandidate(application common.Address) error {
	transaction, err := ec.ecdsaKeepFactoryContract.RegisterMemberCandidate(
		ec.transactorOptions,
		application,
	)
	if err != nil {
		return err
	}

	logger.Debugf("submitted RegisterMemberCandidate transaction with hash: [%x]", transaction.Hash())

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
	keepAddress common.Address,
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
	keepAddress common.Address,
	publicKey [64]byte,
) error {
	keepContract, err := ec.getKeepContract(keepAddress)
	if err != nil {
		return err
	}

	// TODO: this is absolutely enough for a group of 3 members but should we
	// support 50 or 100?
	transactorOptions := bind.TransactOpts(*ec.transactorOptions)
	transactorOptions.GasLimit = 200000

	transaction, err := keepContract.SubmitPublicKey(&transactorOptions, publicKey[:])
	if err != nil {
		return err
	}

	logger.Debugf("submitted SubmitPublicKey transaction with hash: [%x]", transaction.Hash())

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
	keepAddress common.Address,
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

// IsAwaitingSignature checks if the keep is waiting for a signature to be
// calculated for the given digest.
func (ec *EthereumChain) IsAwaitingSignature(keepAddress common.Address, digest [32]byte) (bool, error) {
	keepContract, err := ec.getKeepContract(keepAddress)
	if err != nil {
		return false, err
	}

	return keepContract.IsAwaitingSignature(
		ec.callerOptions,
		digest,
	)
}
