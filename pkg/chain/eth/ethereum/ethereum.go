// Package ethereum contains implementation of ethereum chain interface.
package ethereum

import (
	cecdsa "crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
	publicKeyHighBytes, publicKeyLowBytes := splitPublicKey(ec.publicKey)

	transaction, err := ec.ecdsaKeepFactoryContract.RegisterMemberCandidate(
		ec.transactorOptions,
		application,
		publicKeyHighBytes,
		publicKeyLowBytes,
	)
	if err != nil {
		return err
	}

	logger.Debugf("submitted RegisterMemberCandidate transaction with hash: [%x]", transaction.Hash())

	return nil
}

func splitPublicKey(publicKey *cecdsa.PublicKey) ([32]byte, [32]byte) {
	publicKeyBytes := crypto.FromECDSAPub(publicKey)[1:] // remove the first 0x04 byte

	// split the key to two 32 bytes buckets
	var publicKeyHighBytes, publicKeyLowBytes [32]byte
	copy(publicKeyHighBytes[:], publicKeyBytes[:32])
	copy(publicKeyLowBytes[:], publicKeyBytes[32:])

	return publicKeyHighBytes, publicKeyLowBytes
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
				KeepAddress:       chainEvent.KeepAddress,
				Members:           chainEvent.Members,
				MembersPublicKeys: getMembersPublicKeys(chainEvent),
			})
		},
		func(err error) error {
			return fmt.Errorf("keep created callback failed: [%v]", err)
		},
	)
}

func getMembersPublicKeys(chainEvent *abi.ECDSAKeepFactoryECDSAKeepCreated) []cecdsa.PublicKey {
	membersPublicKeys := make([]cecdsa.PublicKey, len(chainEvent.Members))

	for i := range membersPublicKeys {
		publicKeyBytes := make([]byte, 0)
		publicKeyBytes = append(publicKeyBytes, uint8(4)) // add the first 0x04 byte
		publicKeyBytes = append(publicKeyBytes, chainEvent.MembersPublicKeysHighBytes[i][:]...)
		publicKeyBytes = append(publicKeyBytes, chainEvent.MembersPublicKeysLowBytes[i][:]...)

		publicKey, _ := crypto.UnmarshalPubkey(publicKeyBytes)
		membersPublicKeys[i] = *publicKey
	}

	return membersPublicKeys
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

// HasMinimumStake returns true if the specified address is staked.  False will
// be returned if not staked.  If err != nil then it was not possible to determine
// if the address is staked or not.
func (ec *EthereumChain) HasMinimumStake(address common.Address) (bool, error) {
	return ec.keepRandomBeaconOperatorContract.HasMinimumStake(&bind.CallOpts{}, address)
}
