package ethereum

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/tbtc/pkg/chain/ethereum/gen/contract"
)

type tbtcApplicationHandle struct {
	bekm               *bondedEcdsaKeepManager
	tbtcSystemAddress  common.Address
	tbtcSystemContract *contract.TBTCSystem
}

func (bekm *bondedEcdsaKeepManager) TBTCApplicationHandle() (chain.TBTCHandle, error) {
	return &tbtcApplicationHandle{
		bekm:              bekm,
		tbtcSystemAddress: common.Address{},
	}, nil
}

// ID returns the Keep Application ID for this application.
func (tah *tbtcApplicationHandle) ID() chain.KeepApplicationID {
	return combinedChainID(tah.tbtcSystemAddress)
}

// RegisterAsMemberCandidate registers client as a candidate to be selected
// to a keep.
func (tah *tbtcApplicationHandle) RegisterAsMemberCandidate() error {
	gasEstimate, err := tah.bekm.handle.bondedECDSAKeepFactoryContract.RegisterMemberCandidateGasEstimate(
		tah.tbtcSystemAddress,
	)
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
	transaction, err := tah.bekm.handle.bondedECDSAKeepFactoryContract.RegisterMemberCandidate(
		tah.tbtcSystemAddress,
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

// IsRegisteredForApplication checks if the operator is registered
// as a signer candidate in the factory for the given application.
func (tah *tbtcApplicationHandle) IsRegisteredForApplication() (bool, error) {
	return tah.bekm.handle.bondedECDSAKeepFactoryContract.IsOperatorRegistered(
		tah.bekm.operatorAddress,
		tah.tbtcSystemAddress,
	)
}

// IsEligibleForApplication checks if the operator is eligible to register
// as a signer candidate for the given application.
func (tah *tbtcApplicationHandle) IsEligibleForApplication() (bool, error) {
	return tah.bekm.handle.bondedECDSAKeepFactoryContract.IsOperatorEligible(
		tah.bekm.operatorAddress,
		tah.tbtcSystemAddress,
	)
}

// IsStatusUpToDateForApplication checks if the operator's status
// is up to date in the signers' pool of the given application.
func (tah *tbtcApplicationHandle) IsStatusUpToDateForApplication() (bool, error) {
	return tah.bekm.handle.bondedECDSAKeepFactoryContract.IsOperatorUpToDate(
		tah.bekm.operatorAddress,
		tah.tbtcSystemAddress,
	)
}

// UpdateStatusForApplication updates the operator's status in the signers'
// pool for the given application.
func (tah *tbtcApplicationHandle) UpdateStatusForApplication() error {
	transaction, err := tah.bekm.handle.bondedECDSAKeepFactoryContract.UpdateOperatorStatus(
		tah.bekm.operatorAddress,
		tah.tbtcSystemAddress,
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

// OnDepositCreated installs a callback that is invoked when an
// on-chain notification of a new deposit creation is seen.
func (tah *tbtcApplicationHandle) OnDepositCreated(
	handler func(depositAddress string),
) subscription.EventSubscription {
	onEvent := func(
		DepositContractAddress common.Address,
		KeepAddress common.Address,
		Timestamp *big.Int,
		blockNumber uint64,
	) {
		handler(DepositContractAddress.Hex())
	}

	return tah.tbtcSystemContract.Created(
		nil,
		nil,
		nil,
	).OnEvent(onEvent)
}

// OnDepositRegisteredPubkey installs a callback that is invoked when an
// on-chain notification of a deposit's pubkey registration is seen.
func (tah *tbtcApplicationHandle) OnDepositRegisteredPubkey(
	handler func(depositAddress string),
) subscription.EventSubscription {
	onEvent := func(
		DepositContractAddress common.Address,
		SigningGroupPubkeyX [32]uint8,
		SigningGroupPubkeyY [32]uint8,
		Timestamp *big.Int,
		blockNumber uint64,
	) {
		handler(DepositContractAddress.Hex())
	}

	return tah.tbtcSystemContract.RegisteredPubkey(nil, nil).OnEvent(onEvent)
}

// OnDepositRedemptionRequested installs a callback that is invoked when an
// on-chain notification of a deposit redemption request is seen.
func (tah *tbtcApplicationHandle) OnDepositRedemptionRequested(
	handler func(depositAddress string),
) subscription.EventSubscription {
	onEvent := func(
		DepositContractAddress common.Address,
		Requester common.Address,
		Digest [32]uint8,
		UtxoValue *big.Int,
		RedeemerOutputScript []uint8,
		RequestedFee *big.Int,
		Outpoint []uint8,
		blockNumber uint64,
	) {
		handler(DepositContractAddress.Hex())
	}

	return tah.tbtcSystemContract.RedemptionRequested(
		nil,
		nil,
		nil,
		nil,
	).OnEvent(onEvent)
}

// OnDepositGotRedemptionSignature installs a callback that is invoked when an
// on-chain notification of a deposit receiving a redemption signature is seen.
func (tah *tbtcApplicationHandle) OnDepositGotRedemptionSignature(
	handler func(depositAddress string),
) subscription.EventSubscription {
	onEvent := func(
		DepositContractAddress common.Address,
		Digest [32]uint8,
		R [32]uint8,
		S [32]uint8,
		Timestamp *big.Int,
		blockNumber uint64,
	) {
		handler(DepositContractAddress.Hex())
	}

	return tah.tbtcSystemContract.GotRedemptionSignature(
		nil,
		nil,
		nil,
	).OnEvent(onEvent)
}

// OnDepositRedeemed installs a callback that is invoked when an
// on-chain notification of a deposit redemption is seen.
func (tah *tbtcApplicationHandle) OnDepositRedeemed(
	handler func(depositAddress string),
) subscription.EventSubscription {
	onEvent := func(
		DepositContractAddress common.Address,
		Txid [32]uint8,
		Timestamp *big.Int,
		blockNumber uint64,
	) {
		handler(DepositContractAddress.Hex())
	}

	return tah.tbtcSystemContract.Redeemed(
		nil,
		nil,
		nil,
	).OnEvent(onEvent)
}

// PastDepositRedemptionRequestedEvents returns all redemption requested
// events for the given deposit which occurred after the provided start block.
// Returned events are sorted by the block number in the ascending order.
func (tah *tbtcApplicationHandle) PastDepositRedemptionRequestedEvents(
	startBlock uint64,
	depositAddress string,
) ([]*chain.DepositRedemptionRequestedEvent, error) {
	if !common.IsHexAddress(depositAddress) {
		return nil, fmt.Errorf("incorrect deposit contract address")
	}
	events, err := tah.tbtcSystemContract.PastRedemptionRequestedEvents(
		startBlock,
		nil,
		[]common.Address{
			common.HexToAddress(depositAddress),
		},
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	result := make([]*chain.DepositRedemptionRequestedEvent, 0)

	for _, event := range events {
		timestamp, err := tah.bekm.handle.BlockTimestamp(new(big.Int).SetUint64(event.Raw.BlockNumber))
		if err != nil {
			return nil, fmt.Errorf(
				"failed to resolve timestamp for RedemptionRequested event "+
					"for deposit [%s] in block [%v]: [%v]",
				event.DepositContractAddress,
				event.Raw.BlockNumber,
				err,
			)
		}

		result = append(result, &chain.DepositRedemptionRequestedEvent{
			DepositAddress:       event.DepositContractAddress.Hex(),
			RequesterAddress:     event.Requester.Hex(),
			Digest:               event.Digest,
			UtxoValue:            event.UtxoValue,
			RedeemerOutputScript: event.RedeemerOutputScript,
			RequestedFee:         event.RequestedFee,
			Outpoint:             event.Outpoint,
			BlockNumber:          event.Raw.BlockNumber,
			Timestamp:            timestamp,
		})
	}

	// Make sure events are sorted by timestamp in ascending order.
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Timestamp < result[j].Timestamp
	})

	return result, nil
}

// KeepAddress returns the underlying keep address for the
// provided deposit.
func (tah *tbtcApplicationHandle) Keep(
	depositAddress string,
) (chain.BondedECDSAKeepHandle, error) {
	deposit, err := tah.getDepositContract(depositAddress)
	if err != nil {
		return nil, err
	}

	keepAddress, err := deposit.KeepAddress()
	if err != nil {
		return nil, err
	}

	return tah.bekm.GetKeepWithID(keepAddress)
}

// RetrieveSignerPubkey retrieves the signer public key for the
// provided deposit.
func (tah *tbtcApplicationHandle) RetrieveSignerPubkey(
	depositAddress string,
) error {
	deposit, err := tah.getDepositContract(depositAddress)
	if err != nil {
		return err
	}

	transaction, err := deposit.RetrieveSignerPubkey()
	if err != nil {
		return err
	}

	logger.Debugf(
		"submitted RetrieveSignerPubkey transaction with hash: [%x]",
		transaction.Hash(),
	)

	return nil
}

// ProvideRedemptionSignature provides the redemption signature for the
// provided deposit.
func (tah *tbtcApplicationHandle) ProvideRedemptionSignature(
	depositAddress string,
	v uint8,
	r [32]uint8,
	s [32]uint8,
) error {
	deposit, err := tah.getDepositContract(depositAddress)
	if err != nil {
		return err
	}

	transaction, err := deposit.ProvideRedemptionSignature(v, r, s)
	if err != nil {
		return err
	}

	logger.Debugf(
		"submitted ProvideRedemptionSignature transaction with hash: [%x]",
		transaction.Hash(),
	)

	return nil
}

// IncreaseRedemptionFee increases the redemption fee for the provided deposit.
func (tah *tbtcApplicationHandle) IncreaseRedemptionFee(
	depositAddress string,
	previousOutputValueBytes [8]uint8,
	newOutputValueBytes [8]uint8,
) error {
	deposit, err := tah.getDepositContract(depositAddress)
	if err != nil {
		return err
	}

	transaction, err := deposit.IncreaseRedemptionFee(
		previousOutputValueBytes,
		newOutputValueBytes,
	)
	if err != nil {
		return err
	}

	logger.Debugf(
		"submitted IncreaseRedemptionFee transaction with hash: [%x]",
		transaction.Hash(),
	)

	return nil
}

// ProvideRedemptionProof provides the redemption proof for the provided deposit.
func (tah *tbtcApplicationHandle) ProvideRedemptionProof(
	depositAddress string,
	txVersion [4]uint8,
	txInputVector []uint8,
	txOutputVector []uint8,
	txLocktime [4]uint8,
	merkleProof []uint8,
	txIndexInBlock *big.Int,
	bitcoinHeaders []uint8,
) error {
	deposit, err := tah.getDepositContract(depositAddress)
	if err != nil {
		return err
	}

	transaction, err := deposit.ProvideRedemptionProof(
		txVersion,
		txInputVector,
		txOutputVector,
		txLocktime,
		merkleProof,
		txIndexInBlock,
		bitcoinHeaders,
	)
	if err != nil {
		return err
	}

	logger.Debugf(
		"submitted ProvideRedemptionProof transaction with hash: [%x]",
		transaction.Hash(),
	)

	return nil
}

// CurrentState returns the current state for the provided deposit.
func (tah *tbtcApplicationHandle) CurrentState(
	depositAddress string,
) (chain.DepositState, error) {
	deposit, err := tah.getDepositContract(depositAddress)
	if err != nil {
		return 0, err
	}

	state, err := deposit.CurrentState()
	if err != nil {
		return 0, err
	}

	return chain.DepositState(state.Uint64()), err
}

func (tah *tbtcApplicationHandle) getDepositContract(
	address string,
) (*contract.Deposit, error) {
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("incorrect deposit contract address")
	}

	depositContract, err := contract.NewDeposit(
		common.HexToAddress(address),
		tah.bekm.handle.accountKey,
		tah.bekm.handle.client,
		tah.bekm.handle.nonceManager,
		tah.bekm.handle.miningWaiter,
		tah.bekm.handle.blockCounter,
		tah.bekm.handle.transactionMutex,
	)
	if err != nil {
		return nil, err
	}

	return depositContract, nil
}
