//+build celo

package celo

import (
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/celo-org/celo-blockchain/common"

	"github.com/keep-network/keep-common/pkg/chain/celo/celoutil"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/celo/contract"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-ecdsa/pkg/utils/byteutils"
)

type bondedEcdsaKeepHandle struct {
	keepID     chain.ID
	operatorID chain.ID
	contract   *contract.BondedECDSAKeep
}

func (cc *celoChain) GetKeepWithID(
	keepID chain.ID,
) (chain.BondedECDSAKeepHandle, error) {
	keepAddress, err := fromChainID(keepID)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to interpret keep ID [%v]: [%v]",
			keepID,
			err,
		)
	}

	bondedECDSAKeepContract, err := contract.NewBondedECDSAKeep(
		keepAddress,
		cc.chainID,
		cc.accountKey,
		cc.client,
		cc.nonceManager,
		cc.miningWaiter,
		cc.blockCounter,
		cc.transactionMutex,
	)
	if err != nil {
		return nil, err
	}

	return &bondedEcdsaKeepHandle{
		keepID:     keepID,
		operatorID: cc.OperatorID(),
		contract:   bondedECDSAKeepContract,
	}, nil
}

func (cc *celoChain) GetKeepAtIndex(
	keepIndex *big.Int,
) (chain.BondedECDSAKeepHandle, error) {
	keepAddress, err := cc.bondedECDSAKeepFactoryContract.GetKeepAtIndex(
		keepIndex,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to look up keep address for index [%v]: [%v]",
			keepIndex,
			err,
		)
	}

	return cc.GetKeepWithID(celoChainID(keepAddress))
}

func (bekh *bondedEcdsaKeepHandle) ID() chain.ID {
	return bekh.keepID
}

// OnSignatureRequested installs a callback that is invoked on-chain
// when a keep's signature is requested.
func (bekh *bondedEcdsaKeepHandle) OnSignatureRequested(
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
	return bekh.contract.SignatureRequested(
		nil,
		nil,
	).OnEvent(onEvent), nil
}

// OnConflictingPublicKeySubmitted installs a callback that is invoked when an
// on-chain notification of a conflicting public key submission is seen.
func (bekh *bondedEcdsaKeepHandle) OnConflictingPublicKeySubmitted(
	handler func(event *chain.ConflictingPublicKeySubmittedEvent),
) (subscription.EventSubscription, error) {
	onEvent := func(
		SubmittingMember common.Address,
		ConflictingPublicKey []byte,
		blockNumber uint64,
	) {
		handler(&chain.ConflictingPublicKeySubmittedEvent{
			SubmittingMember:     celoChainID(SubmittingMember),
			ConflictingPublicKey: ConflictingPublicKey,
			BlockNumber:          blockNumber,
		})
	}
	return bekh.contract.ConflictingPublicKeySubmitted(
		nil,
		nil,
	).OnEvent(onEvent), nil
}

// OnPublicKeyPublished installs a callback that is invoked when an on-chain
// event of a published public key was emitted.
func (bekh *bondedEcdsaKeepHandle) OnPublicKeyPublished(
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
	return bekh.contract.PublicKeyPublished(nil).OnEvent(onEvent), nil
}

// SubmitKeepPublicKey submits a public key to a keep contract deployed under
// a given address.
func (bekh *bondedEcdsaKeepHandle) SubmitKeepPublicKey(
	publicKey [64]byte,
) error {
	submitPubKey := func() error {
		transaction, err := bekh.contract.SubmitPublicKey(
			publicKey[:],
			celoutil.TransactionOptions{
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
	// case is when Celo nodes are behind a load balancer and not fully synced
	// with each other. To mitigate this issue, a client will retry submitting
	// a public key up to 10 times with a 250ms interval.
	if err := withRetry(submitPubKey); err != nil {
		return err
	}

	return nil
}

// SubmitSignature submits a signature to a keep contract deployed under a
// given address.
func (bekh *bondedEcdsaKeepHandle) SubmitSignature(
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

	transaction, err := bekh.contract.SubmitSignature(
		signatureR,
		signatureS,
		uint8(signature.RecoveryID),
	)
	if err != nil {
		return err
	}

	logger.Debugf(
		"submitted SubmitSignature transaction with hash: [%x]",
		transaction.Hash(),
	)

	return nil
}

// OnKeepClosed installs a callback that is invoked on-chain when keep is closed.
func (bekh *bondedEcdsaKeepHandle) OnKeepClosed(
	handler func(event *chain.KeepClosedEvent),
) (subscription.EventSubscription, error) {
	onEvent := func(blockNumber uint64) {
		handler(&chain.KeepClosedEvent{BlockNumber: blockNumber})
	}
	return bekh.contract.KeepClosed(&ethlike.SubscribeOpts{
		Tick:       4 * time.Hour,
		PastBlocks: 2000,
	}).OnEvent(onEvent), nil
}

// OnKeepTerminated installs a callback that is invoked on-chain when keep
// is terminated.
func (bekh *bondedEcdsaKeepHandle) OnKeepTerminated(
	handler func(event *chain.KeepTerminatedEvent),
) (subscription.EventSubscription, error) {
	onEvent := func(blockNumber uint64) {
		handler(&chain.KeepTerminatedEvent{BlockNumber: blockNumber})
	}
	return bekh.contract.KeepTerminated(&ethlike.SubscribeOpts{
		Tick:       4 * time.Hour,
		PastBlocks: 2000,
	}).OnEvent(onEvent), nil
}

// IsAwaitingSignature checks if the keep is waiting for a signature to be
// calculated for the given digest.
func (bekh *bondedEcdsaKeepHandle) IsAwaitingSignature(digest [32]byte) (bool, error) {
	return bekh.contract.IsAwaitingSignature(digest)
}

// IsActive checks for current state of a keep on-chain.
func (bekh *bondedEcdsaKeepHandle) IsActive() (bool, error) {
	return bekh.contract.IsActive()
}

// LatestDigest returns the latest digest requested to be signed.
func (bekh *bondedEcdsaKeepHandle) LatestDigest() ([32]byte, error) {
	return bekh.contract.Digest()
}

// SignatureRequestedBlock returns block number from the moment when a
// signature was requested for the given digest from a keep.
// If a signature was not requested for the given digest, returns 0.
func (bekh *bondedEcdsaKeepHandle) SignatureRequestedBlock(
	digest [32]byte,
) (uint64, error) {
	blockNumber, err := bekh.contract.Digests(digest)
	if err != nil {
		return 0, err
	}

	return blockNumber.Uint64(), nil
}

// GetPublicKey returns keep's public key. If there is no public key yet,
// an empty slice is returned.
func (bekh *bondedEcdsaKeepHandle) GetPublicKey() ([]uint8, error) {
	return bekh.contract.GetPublicKey()
}

// GetMembers returns keep's members.
func (bekh *bondedEcdsaKeepHandle) GetMembers() ([]chain.ID, error) {
	addresses, err := bekh.contract.GetMembers()
	if err != nil {
		return nil, err
	}

	return toIDSlice(addresses), err
}

func (bekh *bondedEcdsaKeepHandle) IsThisOperatorMember() (bool, error) {
	operatorIndex, err := bekh.OperatorIndex()
	if err != nil {
		return false, err
	}

	return operatorIndex != -1, nil
}

func (bekh *bondedEcdsaKeepHandle) OperatorIndex() (int, error) {
	memberIDs, err := bekh.GetMembers()
	if err != nil {
		return -1, err
	}

	operatorMemberID := bekh.operatorID

	for i, memberID := range memberIDs {
		if operatorMemberID.String() == memberID.String() {
			return i, nil
		}
	}

	return -1, nil
}

// GetHonestThreshold returns keep's honest threshold.
func (bekh *bondedEcdsaKeepHandle) GetHonestThreshold() (uint64, error) {
	threshold, err := bekh.contract.HonestThreshold()
	if err != nil {
		return 0, err
	}

	return threshold.Uint64(), nil
}

// GetOpenedTimestamp returns timestamp when the keep was created.
func (bekh *bondedEcdsaKeepHandle) GetOpenedTimestamp() (time.Time, error) {
	timestamp, err := bekh.contract.GetOpenedTimestamp()
	if err != nil {
		return time.Unix(0, 0), err
	}

	keepOpenTime := time.Unix(timestamp.Int64(), 0)

	return keepOpenTime, nil
}

// PastSignatureSubmittedEvents returns all signature submitted events
// for the given keep which occurred after the provided start block.
// Returned events are sorted by the block number in the ascending order.
func (bekh *bondedEcdsaKeepHandle) PastSignatureSubmittedEvents(
	startBlock uint64,
) ([]*chain.SignatureSubmittedEvent, error) {
	events, err := bekh.contract.PastSignatureSubmittedEvents(
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

// TODO Move to keep-common and parametrize by number of retries and delay?
func withRetry(fn func() error) error {
	const numberOfRetries = 10
	const delay = 12 * time.Second

	for i := 1; ; i++ {
		err := fn()
		if err != nil {
			logger.Errorf("Error occurred [%v]; on [%v] retry", err, i)
			if i == numberOfRetries {
				return err
			}
			time.Sleep(delay)
		} else {
			return nil
		}
	}
}
