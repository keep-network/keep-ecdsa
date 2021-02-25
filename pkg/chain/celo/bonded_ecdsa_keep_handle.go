//+build celo

package celo

import (
	"fmt"
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

// OnSignatureRequested installs a callback that is invoked on-chain
// when a keep's signature is requested.
func (c *Chain) OnSignatureRequested(
	keepAddress ExternalAddress,
	handler func(event *chain.SignatureRequestedEvent),
) (subscription.EventSubscription, error) {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract abi: [%v]", err)
	}

	onEvent := func(
		Digest [32]uint8,
		blockNumber uint64,
	) {
		handler(&chain.SignatureRequestedEvent{
			Digest:      Digest,
			BlockNumber: blockNumber,
		})
	}
	return keepContract.SignatureRequested(
		nil,
		nil,
	).OnEvent(onEvent), nil
}

// OnConflictingPublicKeySubmitted installs a callback that is invoked when an
// on-chain notification of a conflicting public key submission is seen.
func (c *Chain) OnConflictingPublicKeySubmitted(
	keepAddress ExternalAddress,
	handler func(event *chain.ConflictingPublicKeySubmittedEvent),
) (subscription.EventSubscription, error) {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract abi: [%v]", err)
	}

	onEvent := func(
		SubmittingMember InternalAddress,
		ConflictingPublicKey []byte,
		blockNumber uint64,
	) {
		handler(&chain.ConflictingPublicKeySubmittedEvent{
			SubmittingMember:     toExternalAddress(SubmittingMember),
			ConflictingPublicKey: ConflictingPublicKey,
			BlockNumber:          blockNumber,
		})
	}
	return keepContract.ConflictingPublicKeySubmitted(
		nil,
		nil,
	).OnEvent(onEvent), nil
}

// OnPublicKeyPublished installs a callback that is invoked when an on-chain
// event of a published public key was emitted.
func (c *Chain) OnPublicKeyPublished(
	keepAddress ExternalAddress,
	handler func(event *chain.PublicKeyPublishedEvent),
) (subscription.EventSubscription, error) {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract abi: [%v]", err)
	}

	onEvent := func(
		PublicKey []byte,
		blockNumber uint64,
	) {
		handler(&chain.PublicKeyPublishedEvent{
			PublicKey:   PublicKey,
			BlockNumber: blockNumber,
		})
	}
	return keepContract.PublicKeyPublished(nil).OnEvent(onEvent), nil
}

// SubmitKeepPublicKey submits a public key to a keep contract deployed under
// a given address.
func (c *Chain) SubmitKeepPublicKey(
	keepAddress ExternalAddress,
	publicKey [64]byte,
) error {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return err
	}

	submitPubKey := func() error {
		transaction, err := keepContract.SubmitPublicKey(
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
	// a new cloned contract has not been registered by the Celo node. Common
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
func (c *Chain) SubmitSignature(
	keepAddress ExternalAddress,
	signature *ecdsa.Signature,
) error {
	keepContract, err := c.getKeepContract(keepAddress)
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

	logger.Debugf(
		"submitted SubmitSignature transaction with hash: [%x]",
		transaction.Hash(),
	)

	return nil
}

// OnKeepClosed installs a callback that is invoked on-chain when keep is closed.
func (c *Chain) OnKeepClosed(
	keepAddress ExternalAddress,
	handler func(event *chain.KeepClosedEvent),
) (subscription.EventSubscription, error) {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract abi: [%v]", err)
	}

	onEvent := func(blockNumber uint64) {
		handler(&chain.KeepClosedEvent{BlockNumber: blockNumber})
	}
	return keepContract.KeepClosed(&ethlike.SubscribeOpts{
		Tick:       4 * time.Hour,
		PastBlocks: 2000,
	}).OnEvent(onEvent), nil
}

// OnKeepTerminated installs a callback that is invoked on-chain when keep
// is terminated.
func (c *Chain) OnKeepTerminated(
	keepAddress ExternalAddress,
	handler func(event *chain.KeepTerminatedEvent),
) (subscription.EventSubscription, error) {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract abi: [%v]", err)
	}

	onEvent := func(blockNumber uint64) {
		handler(&chain.KeepTerminatedEvent{BlockNumber: blockNumber})
	}
	return keepContract.KeepTerminated(&ethlike.SubscribeOpts{
		Tick:       4 * time.Hour,
		PastBlocks: 2000,
	}).OnEvent(onEvent), nil
}

// IsAwaitingSignature checks if the keep is waiting for a signature to be
// calculated for the given digest.
func (c *Chain) IsAwaitingSignature(
	keepAddress ExternalAddress,
	digest [32]byte,
) (bool, error) {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return false, err
	}

	return keepContract.IsAwaitingSignature(digest)
}

// IsActive checks for current state of a keep on-chain.
func (c *Chain) IsActive(keepAddress ExternalAddress) (bool, error) {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return false, err
	}

	return keepContract.IsActive()
}

// LatestDigest returns the latest digest requested to be signed.
func (c *Chain) LatestDigest(
	keepAddress ExternalAddress,
) ([32]byte, error) {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return [32]byte{}, err
	}

	return keepContract.Digest()
}

// SignatureRequestedBlock returns block number from the moment when a
// signature was requested for the given digest from a keep.
// If a signature was not requested for the given digest, returns 0.
func (c *Chain) SignatureRequestedBlock(
	keepAddress ExternalAddress,
	digest [32]byte,
) (uint64, error) {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return 0, err
	}

	blockNumber, err := keepContract.Digests(digest)
	if err != nil {
		return 0, err
	}

	return blockNumber.Uint64(), nil
}

// GetPublicKey returns keep's public key. If there is no public key yet,
// an empty slice is returned.
func (c *Chain) GetPublicKey(keepAddress ExternalAddress) ([]uint8, error) {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return []uint8{}, err
	}

	return keepContract.GetPublicKey()
}

// GetMembers returns keep's members.
func (c *Chain) GetMembers(
	keepAddress ExternalAddress,
) ([]ExternalAddress, error) {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return []ExternalAddress{}, err
	}

	addresses, err := keepContract.GetMembers()
	if err != nil {
		return []ExternalAddress{}, err
	}

	return toExternalAddresses(addresses), err
}

// GetHonestThreshold returns keep's honest threshold.
func (c *Chain) GetHonestThreshold(
	keepAddress ExternalAddress,
) (uint64, error) {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return 0, err
	}

	threshold, err := keepContract.HonestThreshold()
	if err != nil {
		return 0, err
	}

	return threshold.Uint64(), nil
}

// GetOpenedTimestamp returns timestamp when the keep was created.
func (c *Chain) GetOpenedTimestamp(
	keepAddress ExternalAddress,
) (time.Time, error) {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return time.Unix(0, 0), err
	}

	timestamp, err := keepContract.GetOpenedTimestamp()
	if err != nil {
		return time.Unix(0, 0), err
	}

	keepOpenTime := time.Unix(timestamp.Int64(), 0)

	return keepOpenTime, nil
}

// PastSignatureSubmittedEvents returns all signature submitted events
// for the given keep which occurred after the provided start block.
// Returned events are sorted by the block number in the ascending order.
func (c *Chain) PastSignatureSubmittedEvents(
	keepAddress string,
	startBlock uint64,
) ([]*chain.SignatureSubmittedEvent, error) {
	if !common.IsHexAddress(keepAddress) {
		return nil, fmt.Errorf("invalid keep address: [%v]", keepAddress)
	}
	keepContract, err := c.getKeepContract(
		toExternalAddress(common.HexToAddress(keepAddress)),
	)
	if err != nil {
		return nil, err
	}

	events, err := keepContract.PastSignatureSubmittedEvents(
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

func (c *Chain) getKeepContract(
	address ExternalAddress,
) (*contract.BondedECDSAKeep, error) {
	bondedECDSAKeepContract, err := contract.NewBondedECDSAKeep(
		fromExternalAddress(address),
		c.accountKey,
		c.client,
		c.nonceManager,
		c.miningWaiter,
		c.blockCounter,
		c.transactionMutex,
	)
	if err != nil {
		return nil, err
	}

	return bondedECDSAKeepContract, nil
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
