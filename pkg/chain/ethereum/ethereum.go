// Package ethereum contains implementation of ethereum chain interface.
package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/eventlog"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-core/pkg/chain/ethereum"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/contract"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-ecdsa/pkg/utils/byteutils"
)

var logger = log.Logger("keep-chain-eth-ethereum")

// Address returns client's ethereum address.
func (ec *EthereumChain) Address() common.Address {
	return ec.accountKey.Address
}

// RegisterAsMemberCandidate registers client as a candidate to be selected
// to a keep.
func (ec *EthereumChain) RegisterAsMemberCandidate(application common.Address) error {
	gasEstimate, err := ec.bondedECDSAKeepFactoryContract.RegisterMemberCandidateGasEstimate(application)
	if err != nil {
		return fmt.Errorf("failed to estimate gas [%v]", err)
	}

	// If we have multiple sortition pool join transactions queued - and that
	// happens when multiple operators become eligible to join at the same time,
	// e.g. after lowering the minimum bond requirement, transactions mined at
	// the end may no longer have valid gas limits as they were estimated based
	// on a different state of the pool. We add 20% safety margin to the original
	// gas estimation to account for that.
	gasEstimateWithMargin := float64(gasEstimate) * float64(1.2)
	transaction, err := ec.bondedECDSAKeepFactoryContract.RegisterMemberCandidate(
		application,
		ethutil.TransactionOptions{
			GasLimit: uint64(gasEstimateWithMargin),
		},
	)
	if err != nil {
		return err
	}

	logger.Debugf("submitted RegisterMemberCandidate transaction with hash: [%x]", transaction.Hash())

	return nil
}

// OnBondedECDSAKeepCreated installs a callback that is invoked when an on-chain
// notification of a new ECDSA keep creation is seen.
func (ec *EthereumChain) OnBondedECDSAKeepCreated(
	handler func(event *eth.BondedECDSAKeepCreatedEvent),
) subscription.EventSubscription {
	subscription, err := ec.bondedECDSAKeepFactoryContract.WatchBondedECDSAKeepCreated(
		func(
			KeepAddress common.Address,
			Members []common.Address,
			Owner common.Address,
			Application common.Address,
			HonestThreshold *big.Int,
			blockNumber uint64,
		) {
			handler(&eth.BondedECDSAKeepCreatedEvent{
				KeepAddress:     KeepAddress,
				Members:         Members,
				HonestThreshold: HonestThreshold.Uint64(),
			})
		},
		func(err error) error {
			return fmt.Errorf("watch keep created failed: [%v]", err)
		},
		nil,
		nil,
		nil,
	)
	if err != nil {
		logger.Errorf("could not watch BondedECDSAKeepCreated event: [%v]", err)
	}

	return subscription
}

// OnKeepClosed installs a callback that is invoked on-chain when keep is closed.
func (ec *EthereumChain) OnKeepClosed(
	keepAddress common.Address,
	handler func(event *eth.KeepClosedEvent),
) (subscription.EventSubscription, error) {
	keepContract, err := ec.getKeepContract(keepAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract abi: [%v]", err)
	}
	return keepContract.WatchKeepClosed(
		func(blockNumber uint64) {
			handler(&eth.KeepClosedEvent{BlockNumber: blockNumber})
		},
		func(err error) error {
			return fmt.Errorf("keep closed callback failed: [%v]", err)
		},
	)
}

// OnKeepTerminated installs a callback that is invoked on-chain when keep
// is terminated.
func (ec *EthereumChain) OnKeepTerminated(
	keepAddress common.Address,
	handler func(event *eth.KeepTerminatedEvent),
) (subscription.EventSubscription, error) {
	keepContract, err := ec.getKeepContract(keepAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract abi: [%v]", err)
	}
	return keepContract.WatchKeepTerminated(
		func(blockNumber uint64) {
			handler(&eth.KeepTerminatedEvent{BlockNumber: blockNumber})
		},
		func(err error) error {
			return fmt.Errorf("keep terminated callback failed: [%v]", err)
		},
	)
}

// OnPublicKeyPublished installs a callback that is invoked when an on-chain
// event of a published public key was emitted.
func (ec *EthereumChain) OnPublicKeyPublished(
	keepAddress common.Address,
	handler func(event *eth.PublicKeyPublishedEvent),
) (subscription.EventSubscription, error) {
	keepContract, err := ec.getKeepContract(keepAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract abi: [%v]", err)
	}

	return keepContract.WatchPublicKeyPublished(
		func(
			PublicKey []byte,
			blockNumber uint64,
		) {
			handler(&eth.PublicKeyPublishedEvent{
				PublicKey: PublicKey,
			})
		},
		func(err error) error {
			return fmt.Errorf("keep created callback failed: [%v]", err)
		},
	)
}

// OnConflictingPublicKeySubmitted installs a callback that is invoked when an
// on-chain notification of a conflicting public key submission is seen.
func (ec *EthereumChain) OnConflictingPublicKeySubmitted(
	keepAddress common.Address,
	handler func(event *eth.ConflictingPublicKeySubmittedEvent),
) (subscription.EventSubscription, error) {
	keepContract, err := ec.getKeepContract(keepAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract abi: [%v]", err)
	}

	return keepContract.WatchConflictingPublicKeySubmitted(
		func(
			SubmittingMember common.Address,
			ConflictingPublicKey []byte,
			blockNumber uint64,
		) {
			handler(&eth.ConflictingPublicKeySubmittedEvent{
				SubmittingMember:     SubmittingMember,
				ConflictingPublicKey: ConflictingPublicKey,
			})
		},
		func(err error) error {
			return fmt.Errorf("keep created callback failed: [%v]", err)
		},
		nil,
	)
}

// OnSignatureRequested installs a callback that is invoked on-chain
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
				Digest:      Digest,
				BlockNumber: blockNumber,
			})
		},
		func(err error) error {
			return fmt.Errorf("keep signature requested callback failed: [%v]", err)
		},
		nil,
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

	submitPubKey := func() error {
		transaction, err := keepContract.SubmitPublicKey(
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
	if err := ec.withRetry(submitPubKey); err != nil {
		return err
	}

	return nil
}

func (ec *EthereumChain) withRetry(fn func() error) error {
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

func (ec *EthereumChain) getKeepContract(address common.Address) (*contract.BondedECDSAKeep, error) {
	bondedECDSAKeepContract, err := contract.NewBondedECDSAKeep(
		address,
		ec.accountKey,
		ec.client,
		ec.nonceManager,
		ec.miningWaiter,
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

// IsActive checks for current state of a keep on-chain.
func (ec *EthereumChain) IsActive(keepAddress common.Address) (bool, error) {
	keepContract, err := ec.getKeepContract(keepAddress)
	if err != nil {
		return false, err
	}

	return keepContract.IsActive()
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

// BlockCounter returns a block counter.
func (ec *EthereumChain) BlockCounter() chain.BlockCounter {
	return ec.blockCounter
}

// IsRegisteredForApplication checks if the operator is registered
// as a signer candidate in the factory for the given application.
func (ec *EthereumChain) IsRegisteredForApplication(application common.Address) (bool, error) {
	return ec.bondedECDSAKeepFactoryContract.IsOperatorRegistered(
		ec.Address(),
		application,
	)
}

// IsEligibleForApplication checks if the operator is eligible to register
// as a signer candidate for the given application.
func (ec *EthereumChain) IsEligibleForApplication(application common.Address) (bool, error) {
	return ec.bondedECDSAKeepFactoryContract.IsOperatorEligible(
		ec.Address(),
		application,
	)
}

// IsStatusUpToDateForApplication checks if the operator's status
// is up to date in the signers' pool of the given application.
func (ec *EthereumChain) IsStatusUpToDateForApplication(application common.Address) (bool, error) {
	return ec.bondedECDSAKeepFactoryContract.IsOperatorUpToDate(
		ec.Address(),
		application,
	)
}

// UpdateStatusForApplication updates the operator's status in the signers'
// pool for the given application.
func (ec *EthereumChain) UpdateStatusForApplication(application common.Address) error {
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

// IsOperatorAuthorized checks if the factory has the authorization to
// operate on stake represented by the provided operator.
func (ec *EthereumChain) IsOperatorAuthorized(operator common.Address) (bool, error) {
	return ec.bondedECDSAKeepFactoryContract.IsOperatorAuthorized(operator)
}

// GetKeepCount returns number of keeps.
func (ec *EthereumChain) GetKeepCount() (*big.Int, error) {
	return ec.bondedECDSAKeepFactoryContract.GetKeepCount()
}

// GetKeepAtIndex returns the address of the keep at the given index.
func (ec *EthereumChain) GetKeepAtIndex(
	keepIndex *big.Int,
) (common.Address, error) {
	return ec.bondedECDSAKeepFactoryContract.GetKeepAtIndex(keepIndex)
}

// LatestDigest returns the latest digest requested to be signed.
func (ec *EthereumChain) LatestDigest(keepAddress common.Address) ([32]byte, error) {
	keepContract, err := ec.getKeepContract(keepAddress)
	if err != nil {
		return [32]byte{}, err
	}

	return keepContract.Digest()
}

// SignatureRequestedBlock returns block number from the moment when a
// signature was requested for the given digest from a keep.
// If a signature was not requested for the given digest, returns 0.
func (ec *EthereumChain) SignatureRequestedBlock(
	keepAddress common.Address,
	digest [32]byte,
) (uint64, error) {
	keepContract, err := ec.getKeepContract(keepAddress)
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
func (ec *EthereumChain) GetPublicKey(keepAddress common.Address) ([]uint8, error) {
	keepContract, err := ec.getKeepContract(keepAddress)
	if err != nil {
		return []uint8{}, err
	}

	return keepContract.GetPublicKey()
}

// GetMembers returns keep's members.
func (ec *EthereumChain) GetMembers(
	keepAddress common.Address,
) ([]common.Address, error) {
	keepContract, err := ec.getKeepContract(keepAddress)
	if err != nil {
		return []common.Address{}, err
	}

	return keepContract.GetMembers()
}

// GetHonestThreshold returns keep's honest threshold.
func (ec *EthereumChain) GetHonestThreshold(
	keepAddress common.Address,
) (uint64, error) {
	keepContract, err := ec.getKeepContract(keepAddress)
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
func (ec *EthereumChain) GetOpenedTimestamp(keepAddress common.Address) (time.Time, error) {
	keepContract, err := ec.getKeepContract(keepAddress)
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
func (ec *EthereumChain) PastSignatureSubmittedEvents(
	keepAddress string,
	startBlock uint64,
) ([]*eth.SignatureSubmittedEvent, error) {
	if !common.IsHexAddress(keepAddress) {
		return nil, fmt.Errorf("invalid keep address: [%v]", keepAddress)
	}

	keepContractEventLog, err := eventlog.NewBondedECDSAKeepEventLog(
		common.HexToAddress(keepAddress),
		ec.client,
	)
	if err != nil {
		return nil, err
	}

	events, err := keepContractEventLog.PastSignatureSubmittedEvents(
		startBlock,
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
			BlockNumber: event.BlockNumber,
		})
	}

	// Make sure events are sorted by block number in ascending order.
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].BlockNumber < result[j].BlockNumber
	})

	return result, nil
}

// BlockTimestamp returns given block's timestamp.
func (ec *EthereumChain) BlockTimestamp(blockNumber *big.Int) (uint64, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelCtx()

	header, err := ec.client.HeaderByNumber(ctx, blockNumber)
	if err != nil {
		return 0, err
	}

	return header.Time, nil
}

//WeiBalanceOf returns the wei balance of the given address from the latest known block.
func (ec *EthereumChain) WeiBalanceOf(address common.Address) (*big.Int, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelCtx()

	return ec.client.BalanceAt(ctx, address, nil)
}

// BalanceMonitor returns a balance monitor.
func (ec *EthereumChain) BalanceMonitor() (chain.BalanceMonitor, error) {
	return ethereum.NewBalanceMonitor(ec.WeiBalanceOf), nil
}
