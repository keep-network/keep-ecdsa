// Package ethereum contains implementation of ethereum chain interface.
package ethereum

import (
	"fmt"
	"time"

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

// RegisterAsMemberCandidate registers client as a candidate to be selected
// to a keep.
func (ec *EthereumChain) RegisterAsMemberCandidate(application common.Address) error {
	transaction, err := ec.bondedECDSAKeepFactoryContract.RegisterMemberCandidate(
		ec.transactorOptions,
		application,
	)
	if err != nil {
		return err
	}

	logger.Debugf("submitted RegisterMemberCandidate transaction with hash: [%x]", transaction.Hash())

	return nil
}

// OnBondedECDSAKeepCreated is a callback that is invoked when an on-chain
// notification of a new ECDSA keep creation is seen.
func (ec *EthereumChain) OnBondedECDSAKeepCreated(
	handler func(event *eth.BondedECDSAKeepCreatedEvent),
) (subscription.EventSubscription, error) {
	return ec.watchECDSAKeepCreated(
		func(
			chainEvent *abi.BondedECDSAKeepFactoryBondedECDSAKeepCreated,
		) {
			handler(&eth.BondedECDSAKeepCreatedEvent{
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
			chainEvent *abi.BondedECDSAKeepSignatureRequested,
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
	transactorOptions := bind.TransactOpts(*ec.transactorOptions)
	transactorOptions.GasLimit = 3000000 // enough for a group size of 16

	numberOfRetries := 4
	delay := 250 * time.Millisecond
	
	// There might be a scenario, when a public key submission fails because of
	// a new cloned contract has not been registered by the ethereum node. Common
	// case is when Ethereum nodes are behind a load balancer and not fully synced
	// with each other. To mitigate this issue, a client will retry submitting
	// a public key up to 4 times with a 250ms interval.
	for i := 1; ; i++ {
		transaction, err := keepContract.SubmitPublicKey(&transactorOptions, publicKey[:])
	
		if err != nil {
			logger.Errorf("Error occurred [%v]; on [%v] retry", err, i)
			if i == numberOfRetries {
				return err
			}
			time.Sleep(delay)
		} else {
			logger.Debugf("submitted SubmitPublicKey transaction with hash: [%x]", transaction.Hash())
			return nil
		}
	}
}

func (ec *EthereumChain) getKeepContract(address common.Address) (*abi.BondedECDSAKeep, error) {
	bondedECDSAKeepContract, err := abi.NewBondedECDSAKeep(address, ec.client)
	if err != nil {
		return nil, err
	}

	return bondedECDSAKeepContract, nil
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
