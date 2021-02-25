//+build celo

package celo

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/keep-network/keep-common/pkg/chain/celo"

	"github.com/keep-network/keep-common/pkg/chain/ethlike"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-common/pkg/chain/celo/celoutil"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-core/pkg/chain"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/celo/contract"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-ecdsa/pkg/utils/byteutils"
)

var logger = log.Logger("keep-chain-celo")

// Address returns client's Celo address.
func (c *Chain) Address() ExternalAddress {
	return toExternalAddress(c.accountKey.Address)
}

// Signing returns signing interface for creating and verifying signatures.
func (c *Chain) Signing() chain.Signing {
	return celoutil.NewSigner(c.accountKey.PrivateKey)
}

// BlockCounter returns a block counter.
func (c *Chain) BlockCounter() chain.BlockCounter {
	return c.blockCounter
}

// RegisterAsMemberCandidate registers client as a candidate to be selected
// to a keep.
func (c *Chain) RegisterAsMemberCandidate(
	application ExternalAddress,
) error {
	gasEstimate, err := c.bondedECDSAKeepFactoryContract.
		RegisterMemberCandidateGasEstimate(fromExternalAddress(application))
	if err != nil {
		return fmt.Errorf("failed to estimate gas [%v]", err)
	}

	// If we have multiple sortition pool join transactions queued - and that
	// happens when multiple operators become eligible to join at the same time,
	// e.g. after lowering the minimum bond requirement, transactions mined at
	// the end may no longer have valid gas limits as they were estimated based
	// on a different state of the pool. We add 20% safety margin to the
	// original gas estimation to account for that.
	gasEstimateWithMargin := float64(gasEstimate) * float64(1.2)
	transaction, err := c.bondedECDSAKeepFactoryContract.
		RegisterMemberCandidate(
			fromExternalAddress(application),
			celoutil.TransactionOptions{
				GasLimit: uint64(gasEstimateWithMargin),
			},
		)
	if err != nil {
		return err
	}

	logger.Debugf(
		"submitted RegisterMemberCandidate transaction with hash: [%x]",
		transaction.Hash(),
	)

	return nil
}

// OnBondedECDSAKeepCreated installs a callback that is invoked when an on-chain
// notification of a new ECDSA keep creation is seen.
func (c *Chain) OnBondedECDSAKeepCreated(
	handler func(event *eth.BondedECDSAKeepCreatedEvent),
) subscription.EventSubscription {
	onEvent := func(
		KeepAddress InternalAddress,
		Members []InternalAddress,
		Owner InternalAddress,
		Application InternalAddress,
		HonestThreshold *big.Int,
		blockNumber uint64,
	) {
		handler(&eth.BondedECDSAKeepCreatedEvent{
			KeepAddress:     toExternalAddress(KeepAddress),
			Members:         toExternalAddresses(Members),
			HonestThreshold: HonestThreshold.Uint64(),
			BlockNumber:     blockNumber,
		})
	}

	return c.bondedECDSAKeepFactoryContract.BondedECDSAKeepCreated(
		nil,
		nil,
		nil,
		nil,
	).OnEvent(onEvent)
}

// OnKeepClosed installs a callback that is invoked on-chain when keep is closed.
func (c *Chain) OnKeepClosed(
	keepAddress ExternalAddress,
	handler func(event *eth.KeepClosedEvent),
) (subscription.EventSubscription, error) {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract abi: [%v]", err)
	}

	onEvent := func(blockNumber uint64) {
		handler(&eth.KeepClosedEvent{BlockNumber: blockNumber})
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
	handler func(event *eth.KeepTerminatedEvent),
) (subscription.EventSubscription, error) {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract abi: [%v]", err)
	}

	onEvent := func(blockNumber uint64) {
		handler(&eth.KeepTerminatedEvent{BlockNumber: blockNumber})
	}
	return keepContract.KeepTerminated(&ethlike.SubscribeOpts{
		Tick:       4 * time.Hour,
		PastBlocks: 2000,
	}).OnEvent(onEvent), nil
}

// OnPublicKeyPublished installs a callback that is invoked when an on-chain
// event of a published public key was emitted.
func (c *Chain) OnPublicKeyPublished(
	keepAddress ExternalAddress,
	handler func(event *eth.PublicKeyPublishedEvent),
) (subscription.EventSubscription, error) {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract abi: [%v]", err)
	}

	onEvent := func(
		PublicKey []byte,
		blockNumber uint64,
	) {
		handler(&eth.PublicKeyPublishedEvent{
			PublicKey:   PublicKey,
			BlockNumber: blockNumber,
		})
	}
	return keepContract.PublicKeyPublished(nil).OnEvent(onEvent), nil
}

// OnConflictingPublicKeySubmitted installs a callback that is invoked when an
// on-chain notification of a conflicting public key submission is seen.
func (c *Chain) OnConflictingPublicKeySubmitted(
	keepAddress ExternalAddress,
	handler func(event *eth.ConflictingPublicKeySubmittedEvent),
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
		handler(&eth.ConflictingPublicKeySubmittedEvent{
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

// OnSignatureRequested installs a callback that is invoked on-chain
// when a keep's signature is requested.
func (c *Chain) OnSignatureRequested(
	keepAddress ExternalAddress,
	handler func(event *eth.SignatureRequestedEvent),
) (subscription.EventSubscription, error) {
	keepContract, err := c.getKeepContract(keepAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract abi: [%v]", err)
	}

	onEvent := func(
		Digest [32]uint8,
		blockNumber uint64,
	) {
		handler(&eth.SignatureRequestedEvent{
			Digest:      Digest,
			BlockNumber: blockNumber,
		})
	}
	return keepContract.SignatureRequested(
		nil,
		nil,
	).OnEvent(onEvent), nil
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
	if err := c.withRetry(submitPubKey); err != nil {
		return err
	}

	return nil
}

func (c *Chain) withRetry(fn func() error) error {
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

// HasMinimumStake returns true if the specified address is staked.  False will
// be returned if not staked.  If err != nil then it was not possible to determine
// if the address is staked or not.
func (c *Chain) HasMinimumStake(address ExternalAddress) (bool, error) {
	return c.bondedECDSAKeepFactoryContract.HasMinimumStake(
		fromExternalAddress(address),
	)
}

// BalanceOf returns the stake balance of the specified address.
func (c *Chain) BalanceOf(address ExternalAddress) (*big.Int, error) {
	return c.bondedECDSAKeepFactoryContract.BalanceOf(
		fromExternalAddress(address),
	)
}

// IsRegisteredForApplication checks if the operator is registered
// as a signer candidate in the factory for the given application.
func (c *Chain) IsRegisteredForApplication(
	application ExternalAddress,
) (bool, error) {
	return c.bondedECDSAKeepFactoryContract.IsOperatorRegistered(
		fromExternalAddress(c.Address()),
		fromExternalAddress(application),
	)
}

// IsEligibleForApplication checks if the operator is eligible to register
// as a signer candidate for the given application.
func (c *Chain) IsEligibleForApplication(
	application ExternalAddress,
) (bool, error) {
	return c.bondedECDSAKeepFactoryContract.IsOperatorEligible(
		fromExternalAddress(c.Address()),
		fromExternalAddress(application),
	)
}

// IsStatusUpToDateForApplication checks if the operator's status
// is up to date in the signers' pool of the given application.
func (c *Chain) IsStatusUpToDateForApplication(
	application ExternalAddress,
) (bool, error) {
	return c.bondedECDSAKeepFactoryContract.IsOperatorUpToDate(
		fromExternalAddress(c.Address()),
		fromExternalAddress(application),
	)
}

// UpdateStatusForApplication updates the operator's status in the signers'
// pool for the given application.
func (c *Chain) UpdateStatusForApplication(
	application ExternalAddress,
) error {
	transaction, err := c.bondedECDSAKeepFactoryContract.UpdateOperatorStatus(
		fromExternalAddress(c.Address()),
		fromExternalAddress(application),
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

// IsOperatorAuthorized checks if the factory has the authorization to
// operate on stake represented by the provided operator.
func (c *Chain) IsOperatorAuthorized(
	operator ExternalAddress,
) (bool, error) {
	return c.bondedECDSAKeepFactoryContract.IsOperatorAuthorized(
		fromExternalAddress(operator),
	)
}

// GetKeepCount returns number of keeps.
func (c *Chain) GetKeepCount() (*big.Int, error) {
	return c.bondedECDSAKeepFactoryContract.GetKeepCount()
}

// GetKeepAtIndex returns the address of the keep at the given index.
func (c *Chain) GetKeepAtIndex(
	keepIndex *big.Int,
) (ExternalAddress, error) {
	address, err := c.bondedECDSAKeepFactoryContract.GetKeepAtIndex(keepIndex)
	if err != nil {
		return ExternalAddress{}, nil
	}

	return toExternalAddress(address), err
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
) ([]*eth.SignatureSubmittedEvent, error) {
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

	result := make([]*eth.SignatureSubmittedEvent, 0)

	for _, event := range events {
		result = append(result, &eth.SignatureSubmittedEvent{
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

// BlockTimestamp returns given block's timestamp.
func (c *Chain) BlockTimestamp(blockNumber *big.Int) (uint64, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelCtx()

	header, err := c.client.HeaderByNumber(ctx, blockNumber)
	if err != nil {
		return 0, err
	}

	return header.Time, nil
}

// WeiBalanceOf returns the wei balance of the given address from the latest
// known block.
func (c *Chain) WeiBalanceOf(address ExternalAddress) (*celo.Wei, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelCtx()

	balance, err := c.client.BalanceAt(ctx, fromExternalAddress(address), nil)
	if err != nil {
		return nil, err
	}

	return celo.WrapWei(balance), err
}

// BalanceMonitor returns a balance monitor.
func (c *Chain) BalanceMonitor() (*celoutil.BalanceMonitor, error) {
	weiBalanceOf := func(address InternalAddress) (*celo.Wei, error) {
		return c.WeiBalanceOf(toExternalAddress(address))
	}

	return celoutil.NewBalanceMonitor(weiBalanceOf), nil
}
