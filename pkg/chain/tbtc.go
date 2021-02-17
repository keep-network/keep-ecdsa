package chain

import (
	"math/big"

	"github.com/keep-network/keep-common/pkg/subscription"
)

// TBTCHandle represents a chain handle extended with TBTC-specific capabilities.
type TBTCHandle interface {
	BondedECDSAKeepApplicationHandle

	Deposit // Instantiate this rather than include it? Encapsulate the address?
	TBTCSystem
}

// Deposit is an interface that provides ability to interact
// with Deposit contracts.
type Deposit interface {
	// Keep returns the underlying keep for the provided deposit.
	Keep(depositAddress string) (BondedECDSAKeepHandle, error)

	// RetrieveSignerPubkey retrieves the signer public key for the
	// provided deposit.
	RetrieveSignerPubkey(depositAddress string) error

	// ProvideRedemptionSignature provides the redemption signature for the
	// provided deposit.
	ProvideRedemptionSignature(
		depositAddress string,
		v uint8,
		r [32]uint8,
		s [32]uint8,
	) error

	// IncreaseRedemptionFee increases the redemption fee for the
	// provided deposit.
	IncreaseRedemptionFee(
		depositAddress string,
		previousOutputValueBytes [8]uint8,
		newOutputValueBytes [8]uint8,
	) error

	// ProvideRedemptionProof provides the redemption proof for the
	// provided deposit.
	ProvideRedemptionProof(
		depositAddress string,
		txVersion [4]uint8,
		txInputVector []uint8,
		txOutputVector []uint8,
		txLocktime [4]uint8,
		merkleProof []uint8,
		txIndexInBlock *big.Int,
		bitcoinHeaders []uint8,
	) error

	// CurrentState returns the current state for the provided deposit.
	CurrentState(depositAddress string) (DepositState, error)
}

// TBTCSystem is an interface that provides ability to interact
// with TBTCSystem contract.
type TBTCSystem interface {
	// OnDepositCreated installs a callback that is invoked when an
	// on-chain notification of a new deposit creation is seen.
	// Pass a Deposit instead of an address, here and below?
	OnDepositCreated(
		handler func(depositAddress string),
	) subscription.EventSubscription

	// OnDepositRegisteredPubkey installs a callback that is invoked when an
	// on-chain notification of a deposit's pubkey registration is seen.
	OnDepositRegisteredPubkey(
		handler func(depositAddress string),
	) subscription.EventSubscription

	// OnDepositRedemptionRequested installs a callback that is invoked when an
	// on-chain notification of a deposit redemption request is seen.
	OnDepositRedemptionRequested(
		handler func(depositAddress string),
	) subscription.EventSubscription

	// OnDepositGotRedemptionSignature installs a callback that is invoked
	// when an on-chain notification of a deposit receiving a redemption
	// signature is seen.
	OnDepositGotRedemptionSignature(
		handler func(depositAddress string),
	) subscription.EventSubscription

	// OnDepositRedeemed installs a callback that is invoked when an
	// on-chain notification of a deposit redemption is seen.
	OnDepositRedeemed(
		handler func(depositAddress string),
	) subscription.EventSubscription

	// PastDepositRedemptionRequestedEvents returns all redemption requested
	// events for the given deposit which occurred after the provided start block.
	// All implementations should returns those events sorted by the
	// block number in the ascending order.
	PastDepositRedemptionRequestedEvents(
		startBlock uint64,
		depositAddress string,
	) ([]*DepositRedemptionRequestedEvent, error)
}

// DepositRedemptionRequestedEvent is an event emitted when a deposit
// redemption has been requested or the redemption fee has been increased.
type DepositRedemptionRequestedEvent struct {
	DepositAddress       string
	RequesterAddress     string
	Digest               [32]byte
	UtxoValue            *big.Int
	RedeemerOutputScript []byte
	RequestedFee         *big.Int
	Outpoint             []byte
	Timestamp            uint64
	BlockNumber          uint64
}

// DepositState represents the deposit state.
type DepositState int

const (
	Start DepositState = iota
	AwaitingSignerSetup
	AwaitingBtcFundingProof
	FailedSetup
	Active
	AwaitingWithdrawalSignature
	AwaitingWithdrawalProof
	Redeemed
	CourtesyCall
	FraudLiquidationInProgress
	LiquidationInProgress
	Liquidated
)
