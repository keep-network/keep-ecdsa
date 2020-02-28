// Package ethereum contains implementation of ethereum chain interface.
package ethereum

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/gen/contract"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-tecdsa/pkg/utils/byteutils"
)

var logger = log.Logger("keep-chain-eth-ethereum")

// Address returns client's ethereum address.
func (ec *EthereumChain) Address() common.Address {
	return ec.accountKey.Address
}

// RegisterAsMemberCandidate registers client as a candidate to be selected
// to a keep.
func (ec *EthereumChain) RegisterAsMemberCandidate(application common.Address) error {
	transaction, err := ec.bondedECDSAKeepFactoryContract.RegisterMemberCandidate(
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
	return ec.bondedECDSAKeepFactoryContract.WatchBondedECDSAKeepCreated(
		func(
			KeepAddress common.Address,
			Members []common.Address,
			Owner common.Address,
			Application common.Address,
			blockNumber uint64,
		) {
			handler(&eth.BondedECDSAKeepCreatedEvent{
				KeepAddress: KeepAddress,
				Members:     Members,
			})
		},
		func(err error) error {
			return fmt.Errorf("watch keep created failed: [%v]", err)
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

	return keepContract.WatchSignatureRequested(
		func(
			Digest [32]uint8,
			blockNumber uint64,
		) {
			handler(&eth.SignatureRequestedEvent{
				Digest: Digest,
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

	transaction, err := keepContract.SubmitPublicKey(
		publicKey[:],
		ethutil.TransactionOptions{
			GasLimit: 3000000, // enough for a group size of 16
		},
	)
	if err != nil {
		return err
	}

	logger.Debugf("submitted SubmitPublicKey transaction with hash: [%x]", transaction.Hash())

	return nil
}

func (ec *EthereumChain) getKeepContract(address common.Address) (*contract.BondedECDSAKeep, error) {
	bondedECDSAKeepContract, err := contract.NewBondedECDSAKeep(
		address,
		ec.accountKey,
		ec.client,
		ec.transactionMutex,
	)
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

	return keepContract.IsAwaitingSignature(digest)
}

// HasMinimumStake returns true if the specified address is staked.  False will
// be returned if not staked.  If err != nil then it was not possible to determine
// if the address is staked or not.
func (ec *EthereumChain) HasMinimumStake(address common.Address) (bool, error) {
	return ec.bondedECDSAKeepFactoryContract.HasMinimumStake(address)
}

// BalanceOf returns the stake balance of the specified address.
func (ec *EthereumChain) BalanceOf(address common.Address) (*big.Int, error) {
	return ec.bondedECDSAKeepFactoryContract.BalanceOf(address)
}

func (ec *EthereumChain) BlockCounter() chain.BlockCounter {
	return ec.blockCounter
}

func (ec *EthereumChain) IsRegistered(application common.Address) (bool, error) {
	return ec.bondedECDSAKeepFactoryContract.IsOperatorRegistered(
		ec.Address(),
		application,
	)
}

func (ec *EthereumChain) IsEligible(application common.Address) (bool, error) {
	return ec.bondedECDSAKeepFactoryContract.IsOperatorEligible(
		ec.Address(),
		application,
	)
}

func (ec *EthereumChain) IsUpToDate(application common.Address) (bool, error) {
	return ec.bondedECDSAKeepFactoryContract.IsOperatorUpToDate(
		ec.Address(),
		application,
	)
}

func (ec *EthereumChain) UpdateStatus(application common.Address) error {
	transaction, err := ec.bondedECDSAKeepFactoryContract.UpdateOperatorStatus(
		ec.Address(),
		application,
	)
	if err != nil {
		return err
	}

	logger.Debugf(
		"submitted UpdateOperatorStatus transaction with hash: [%x]",
		transaction.Hash(),
	)

	return nil
}
