package ethereum

import (
	"fmt"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/contract"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-ecdsa/pkg/utils/byteutils"
)

type bondedEcdsaKeep struct {
	handle   *ethereumChain
	contract *contract.BondedECDSAKeep

	contractAddress common.Address // FIXME Expose address on contract.BondedECDSAKeep?
}

// GetKeepWithID returns a handle to the BondedECDSAKeep with the provided id.
func (bekm *bondedEcdsaKeepManager) GetKeepWithID(
	keepID chain.KeepID,
) (chain.BondedECDSAKeepHandle, error) {
	// Inside the Ethereum chain, keep ids are always addresses.
	keepAddressString := keepID.String()
	if !common.IsHexAddress(keepAddressString) {
		return nil, fmt.Errorf("incorrect keep address [%s]", keepAddressString)
	}

	bondedECDSAKeepContract, err := contract.NewBondedECDSAKeep(
		common.HexToAddress(keepAddressString),
		bekm.handle.accountKey,
		bekm.handle.client,
		bekm.handle.nonceManager,
		bekm.handle.miningWaiter,
		bekm.handle.blockCounter,
		bekm.handle.transactionMutex,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract abi: [%v]", err)
	}

	return &bondedEcdsaKeep{
		handle:          bekm.handle,
		contract:        bondedECDSAKeepContract,
		contractAddress: common.HexToAddress(keepAddressString),
	}, nil
}

// ID returns this keep's Ethereum address as a chain-agnostic KeepID.
func (bek *bondedEcdsaKeep) ID() chain.KeepID {
	return bek.contractAddress
}

// OnSignatureRequested installs a callback that is invoked on-chain
// when a keep's signature is requested.
func (bek *bondedEcdsaKeep) OnSignatureRequested(
	handler func(event *chain.SignatureRequestedEvent),
) (subscription.EventSubscription, error) {
	onEvent := func(
		Digest [32]uint8,
		blockNumber uint64,
	) {
		handler(&chain.SignatureRequestedEvent{
			Digest:      Digest,
			BlockNumber: blockNumber,
		})
	}
	return bek.contract.SignatureRequested(
		nil,
		nil,
	).OnEvent(onEvent), nil
}

// SubmitKeepPublicKey submits a public key to a keep contract deployed under
// a given address.
func (bek *bondedEcdsaKeep) SubmitKeepPublicKey(
	publicKey [64]byte,
) error {
	submitPubKey := func() error {
		transaction, err := bek.contract.SubmitPublicKey(
			publicKey[:],
			ethutil.TransactionOptions{
				GasLimit: 350000, // enough for a group size of 16
			},
		)
		if err != nil {
			return err
		}

		logger.Debugf("submitted SubmitPublicKey transaction with hash: [%x]", transaction.Hash())
		return nil
	}

	// There might be a scenario, when a public key submission fails because of
	// a new cloned contract has not been registered by the ethereum node. Common
	// case is when Ethereum nodes are behind a load balancer and not fully synced
	// with each other. To mitigate this issue, a client will retry submitting
	// a public key up to 10 times with a 250ms interval.
	if err := bek.handle.withRetry(submitPubKey); err != nil {
		return err
	}

	return nil
}

// OnConflictingPublicKeySubmitted installs a callback that is invoked when an
// on-chain notification of a conflicting public key submission is seen.
func (bek *bondedEcdsaKeep) OnConflictingPublicKeySubmitted(
	handler func(event *chain.ConflictingPublicKeySubmittedEvent),
) (subscription.EventSubscription, error) {
	onEvent := func(
		SubmittingMember common.Address,
		ConflictingPublicKey []byte,
		blockNumber uint64,
	) {
		handler(&chain.ConflictingPublicKeySubmittedEvent{
			SubmittingMember:     SubmittingMember,
			ConflictingPublicKey: ConflictingPublicKey,
			BlockNumber:          blockNumber,
		})
	}
	return bek.contract.ConflictingPublicKeySubmitted(
		nil,
		nil,
	).OnEvent(onEvent), nil
}

// SubmitSignature submits a signature to a keep contract deployed under a
// given address.
func (bek *bondedEcdsaKeep) SubmitSignature(
	signature *ecdsa.Signature,
) error {
	signatureR, err := byteutils.BytesTo32Byte(signature.R.Bytes())
	if err != nil {
		return err
	}

	signatureS, err := byteutils.BytesTo32Byte(signature.S.Bytes())
	if err != nil {
		return err
	}

	transaction, err := bek.contract.SubmitSignature(
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
func (bek *bondedEcdsaKeep) IsAwaitingSignature(digest [32]byte) (bool, error) {
	return bek.contract.IsAwaitingSignature(digest)
}

// IsActive checks for current state of a keep on-chain.
func (bek *bondedEcdsaKeep) IsActive() (bool, error) {
	return bek.contract.IsActive()
}

// LatestDigest returns the latest digest requested to be signed.
func (bek *bondedEcdsaKeep) LatestDigest() ([32]byte, error) {
	return bek.contract.Digest()
}

// SignatureRequestedBlock returns block number from the moment when a
// signature was requested for the given digest from a keep.
// If a signature was not requested for the given digest, returns 0.
func (bek *bondedEcdsaKeep) SignatureRequestedBlock(
	digest [32]byte,
) (uint64, error) {
	blockNumber, err := bek.contract.Digests(digest)
	if err != nil {
		return 0, err
	}

	return blockNumber.Uint64(), nil
}

// GetPublicKey returns keep's public key. If there is no public key yet,
// an empty slice is returned.
func (bek *bondedEcdsaKeep) GetPublicKey() ([]uint8, error) {
	return bek.contract.GetPublicKey()
}

// GetMembers returns keep's members.
func (bek *bondedEcdsaKeep) GetMembers() ([]common.Address, error) {
	return bek.contract.GetMembers()
}

// GetHonestThreshold returns keep's honest threshold.
func (bek *bondedEcdsaKeep) GetHonestThreshold() (uint64, error) {
	threshold, err := bek.contract.HonestThreshold()
	if err != nil {
		return 0, err
	}

	return threshold.Uint64(), nil
}

// GetOpenedTimestamp returns timestamp when the keep was created.
func (bek *bondedEcdsaKeep) GetOpenedTimestamp() (time.Time, error) {
	timestamp, err := bek.contract.GetOpenedTimestamp()
	if err != nil {
		return time.Unix(0, 0), err
	}

	keepOpenTime := time.Unix(timestamp.Int64(), 0)

	return keepOpenTime, nil
}

// PastSignatureSubmittedEvents returns all signature submitted events
// for the given keep which occurred after the provided start block.
// Returned events are sorted by the block number in the ascending order.
func (bek *bondedEcdsaKeep) PastSignatureSubmittedEvents(
	startBlock uint64,
) ([]*chain.SignatureSubmittedEvent, error) {
	events, err := bek.contract.PastSignatureSubmittedEvents(
		startBlock,
		nil, // latest block
		nil,
	)
	if err != nil {
		return nil, err
	}

	result := make([]*chain.SignatureSubmittedEvent, 0)

	for _, event := range events {
		result = append(result, &chain.SignatureSubmittedEvent{
			Digest:      event.Digest,
			R:           event.R,
			S:           event.S,
			RecoveryID:  event.RecoveryID,
			BlockNumber: event.Raw.BlockNumber,
		})
	}

	// Make sure events are sorted by block number in ascending order.
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].BlockNumber < result[j].BlockNumber
	})

	return result, nil
}

// OnKeepClosed installs a callback that is invoked on-chain when keep is closed.
func (bek *bondedEcdsaKeep) OnKeepClosed(
	handler func(event *chain.KeepClosedEvent),
) (subscription.EventSubscription, error) {
	onEvent := func(blockNumber uint64) {
		handler(&chain.KeepClosedEvent{BlockNumber: blockNumber})
	}
	return bek.contract.KeepClosed(&ethutil.SubscribeOpts{
		Tick:       4 * time.Hour,
		PastBlocks: 2000,
	}).OnEvent(onEvent), nil
}

// OnKeepTerminated installs a callback that is invoked on-chain when keep
// is terminated.
func (bek *bondedEcdsaKeep) OnKeepTerminated(
	handler func(event *chain.KeepTerminatedEvent),
) (subscription.EventSubscription, error) {
	onEvent := func(blockNumber uint64) {
		handler(&chain.KeepTerminatedEvent{BlockNumber: blockNumber})
	}
	return bek.contract.KeepTerminated(&ethutil.SubscribeOpts{
		Tick:       4 * time.Hour,
		PastBlocks: 2000,
	}).OnEvent(onEvent), nil
}

// OnPublicKeyPublished installs a callback that is invoked when an on-chain
// event of a published public key was emitted.
func (bek *bondedEcdsaKeep) OnPublicKeyPublished(
	handler func(event *chain.PublicKeyPublishedEvent),
) (subscription.EventSubscription, error) {
	onEvent := func(
		PublicKey []byte,
		blockNumber uint64,
	) {
		handler(&chain.PublicKeyPublishedEvent{
			PublicKey:   PublicKey,
			BlockNumber: blockNumber,
		})
	}
	return bek.contract.PublicKeyPublished(nil).OnEvent(onEvent), nil
}
