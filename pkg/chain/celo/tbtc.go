//+build celo

package celo

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/keep-network/keep-common/pkg/chain/celo/celoutil"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/celo/contract"
	tbtcchain "github.com/keep-network/tbtc/pkg/chain/celo/gen/contract"
)

// tbtcApplication represents a tBTC application handle conforming to
// chain.TBTCHandle.
type tbtcApplication struct {
	chainHandle *celoChain

	bondedECDSAKeepFactoryContract *contract.BondedECDSAKeepFactory

	tbtcSystemAddress  common.Address
	tbtcSystemContract *tbtcchain.TBTCSystem
}

func (cc *celoChain) TBTCApplicationHandle() (chain.TBTCHandle, error) {
	var emptyAddress = common.Address{}
	if cc.tbtcSystemAddress == emptyAddress {
		return nil, fmt.Errorf("TBTCSystem address unset")
	}

	tbtcSystemContract, err := tbtcchain.NewTBTCSystem(
		cc.tbtcSystemAddress,
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

	return &tbtcApplication{
		chainHandle:                    cc,
		bondedECDSAKeepFactoryContract: cc.bondedECDSAKeepFactoryContract,
		tbtcSystemAddress:              cc.tbtcSystemAddress,
		tbtcSystemContract:             tbtcSystemContract,
	}, nil
}

func (ta *tbtcApplication) ID() chain.ID {
	return celoChainID(ta.tbtcSystemAddress)
}

func (ta *tbtcApplication) RegisterAsMemberCandidate() error {
	gasEstimate, err :=
		ta.bondedECDSAKeepFactoryContract.RegisterMemberCandidateGasEstimate(
			ta.tbtcSystemAddress,
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
	transaction, err := ta.bondedECDSAKeepFactoryContract.RegisterMemberCandidate(
		ta.tbtcSystemAddress,
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

// IsRegisteredForApplication checks if the operator is registered
// as a signer candidate in the factory for the given application.
func (ta *tbtcApplication) IsRegisteredForApplication() (bool, error) {
	return ta.bondedECDSAKeepFactoryContract.IsOperatorRegistered(
		ta.chainHandle.operatorAddress(),
		ta.tbtcSystemAddress,
	)
}

// IsEligibleForApplication checks if the operator is eligible to register
// as a signer candidate for the given application.
func (ta *tbtcApplication) IsEligibleForApplication() (bool, error) {
	return ta.bondedECDSAKeepFactoryContract.IsOperatorEligible(
		ta.chainHandle.operatorAddress(),
		ta.tbtcSystemAddress,
	)
}

// IsStatusUpToDateForApplication checks if the operator's status
// is up to date in the signers' pool of the given application.
func (ta *tbtcApplication) IsStatusUpToDateForApplication() (bool, error) {
	return ta.bondedECDSAKeepFactoryContract.IsOperatorUpToDate(
		ta.chainHandle.operatorAddress(),
		ta.tbtcSystemAddress,
	)
}

// UpdateStatusForApplication updates the operator's status in the signers'
// pool for the given application.
func (ta *tbtcApplication) UpdateStatusForApplication() error {
	transaction, err := ta.bondedECDSAKeepFactoryContract.UpdateOperatorStatus(
		ta.chainHandle.operatorAddress(),
		ta.tbtcSystemAddress,
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
func (ta *tbtcApplication) OnDepositCreated(
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

	return ta.tbtcSystemContract.Created(
		nil,
		nil,
		nil,
	).OnEvent(onEvent)
}

// OnDepositRegisteredPubkey installs a callback that is invoked when an
// on-chain notification of a deposit's pubkey registration is seen.
func (ta *tbtcApplication) OnDepositRegisteredPubkey(
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

	return ta.tbtcSystemContract.RegisteredPubkey(nil, nil).OnEvent(onEvent)
}

// OnDepositRedemptionRequested installs a callback that is invoked when an
// on-chain notification of a deposit redemption request is seen.
func (ta *tbtcApplication) OnDepositRedemptionRequested(
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

	return ta.tbtcSystemContract.RedemptionRequested(
		nil,
		nil,
		nil,
		nil,
	).OnEvent(onEvent)
}

// OnDepositGotRedemptionSignature installs a callback that is invoked when an
// on-chain notification of a deposit receiving a redemption signature is seen.
func (ta *tbtcApplication) OnDepositGotRedemptionSignature(
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

	return ta.tbtcSystemContract.GotRedemptionSignature(
		nil,
		nil,
		nil,
	).OnEvent(onEvent)
}

// OnDepositRedeemed installs a callback that is invoked when an
// on-chain notification of a deposit redemption is seen.
func (ta *tbtcApplication) OnDepositRedeemed(
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

	return ta.tbtcSystemContract.Redeemed(
		nil,
		nil,
		nil,
	).OnEvent(onEvent)
}

// PastDepositRedemptionRequestedEvents returns all redemption requested
// events for the given deposit which occurred after the provided start block.
// Returned events are sorted by the block number in the ascending order.
func (ta *tbtcApplication) PastDepositRedemptionRequestedEvents(
	startBlock uint64,
	depositAddress string,
) ([]*chain.DepositRedemptionRequestedEvent, error) {
	if !common.IsHexAddress(depositAddress) {
		return nil, fmt.Errorf("incorrect deposit contract address")
	}
	events, err := ta.tbtcSystemContract.PastRedemptionRequestedEvents(
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
		result = append(result, &chain.DepositRedemptionRequestedEvent{
			DepositAddress:       event.DepositContractAddress.Hex(),
			RequesterAddress:     event.Requester.Hex(),
			Digest:               event.Digest,
			UtxoValue:            event.UtxoValue,
			RedeemerOutputScript: event.RedeemerOutputScript,
			RequestedFee:         event.RequestedFee,
			Outpoint:             event.Outpoint,
			BlockNumber:          event.Raw.BlockNumber,
		})
	}

	// Make sure events are sorted by block number in ascending order.
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].BlockNumber < result[j].BlockNumber
	})

	return result, nil
}

func (ta *tbtcApplication) Keep(
	depositAddress string,
) (chain.BondedECDSAKeepHandle, error) {
	deposit, err := ta.getDepositContract(depositAddress)
	if err != nil {
		return nil, err
	}

	keepAddress, err := deposit.KeepAddress()
	if err != nil {
		return nil, err
	}

	return ta.chainHandle.GetKeepWithID(celoChainID(keepAddress))
}

// RetrieveSignerPubkey retrieves the signer public key for the
// provided deposit.
func (ta *tbtcApplication) RetrieveSignerPubkey(
	depositAddress string,
) error {
	deposit, err := ta.getDepositContract(depositAddress)
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
func (ta *tbtcApplication) ProvideRedemptionSignature(
	depositAddress string,
	v uint8,
	r [32]uint8,
	s [32]uint8,
) error {
	deposit, err := ta.getDepositContract(depositAddress)
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
func (ta *tbtcApplication) IncreaseRedemptionFee(
	depositAddress string,
	previousOutputValueBytes [8]uint8,
	newOutputValueBytes [8]uint8,
) error {
	deposit, err := ta.getDepositContract(depositAddress)
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
func (ta *tbtcApplication) ProvideRedemptionProof(
	depositAddress string,
	txVersion [4]uint8,
	txInputVector []uint8,
	txOutputVector []uint8,
	txLocktime [4]uint8,
	merkleProof []uint8,
	txIndexInBlock *big.Int,
	bitcoinHeaders []uint8,
) error {
	deposit, err := ta.getDepositContract(depositAddress)
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
func (ta *tbtcApplication) CurrentState(
	depositAddress string,
) (chain.DepositState, error) {
	deposit, err := ta.getDepositContract(depositAddress)
	if err != nil {
		return 0, err
	}

	state, err := deposit.CurrentState()
	if err != nil {
		return 0, err
	}

	return chain.DepositState(state.Uint64()), err
}

func (ta *tbtcApplication) getDepositContract(
	address string,
) (*tbtcchain.Deposit, error) {
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("incorrect deposit contract address")
	}

	depositContract, err := tbtcchain.NewDeposit(
		common.HexToAddress(address),
		ta.chainHandle.chainID,
		ta.chainHandle.accountKey,
		ta.chainHandle.client,
		ta.chainHandle.nonceManager,
		ta.chainHandle.miningWaiter,
		ta.chainHandle.blockCounter,
		ta.chainHandle.transactionMutex,
	)
	if err != nil {
		return nil, err
	}

	return depositContract, nil
}
