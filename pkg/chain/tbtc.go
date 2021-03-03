package chain

import (
	"math/big"

	"github.com/keep-network/keep-common/pkg/subscription"
)

// TBTCHandle represents handle to the tBTC on-chain application. It extends the
// BondedECDSAKeepApplicationHandle interface with tBTC-specific functionality.
type TBTCHandle interface {
	BondedECDSAKeepApplicationHandle

	Deposit
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
	BlockNumber          uint64
}

// DepositState represents the deposit state.
type DepositState int

const (
	// Start is initial deposit state
	Start DepositState = iota
	// AwaitingSignerSetup represents that the system is awaiting the signers to set up
	AwaitingSignerSetup
	// AwaitingBtcFundingProof represents that the system is awaiting proof that BTC funding has occured
	AwaitingBtcFundingProof
	// FailedSetup represents that the setup has failed
	FailedSetup
	// Active represents that the BTC has been secured and the TDT has been issued
	Active
	// AwaitingWithdrawalSignature represents that the redemption process has started, and is waiting on a signature
	AwaitingWithdrawalSignature
	// AwaitingWithdrawalProof represents that the remption process is waiting on proof of redemption on the BTC blockchain
	AwaitingWithdrawalProof
	// Redeemed represents that the BTC has been dispersed and the TDT/TBTC has been destroyed
	Redeemed
	// CourtesyCall represents that the keep is in danger of being liquidated, and so should be redeemed immediately
	CourtesyCall
	// FraudLiquidationInProgress means that fraud was detected, and so the keep is being liquidated
	FraudLiquidationInProgress
	// LiquidationInProgress means that the system has seized the eth collateral and the signers are trying to recover the held BTC
	LiquidationInProgress
	// Liquidated means that the system seized the eth collateral and the signers recovered the held BTC
	Liquidated
)
