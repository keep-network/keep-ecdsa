package tbtc

import (
	"github.com/keep-network/keep-common/pkg/subscription"
	chain "github.com/keep-network/keep-ecdsa/pkg/chain"
)

// Handle represents a chain handle extended with TBTC-specific capabilities.
type Handle interface {
	chain.Handle

	Deposit
	TBTCSystem
}

// Deposit is an interface that provides ability to interact
// with Deposit contracts.
type Deposit interface {
	// KeepAddress returns the underlying keep address for the
	// provided deposit.
	KeepAddress(depositAddress string) (string, error)

	// RetrieveSignerPubkey retrieves the signer public key for the
	// provided deposit.
	RetrieveSignerPubkey(depositAddress string) error
}

// TBTCSystem is an interface that provides ability to interact
// with TBTCSystem contract.
type TBTCSystem interface {
	// OnDepositCreated installs a callback that is invoked when an
	// on-chain notification of a new deposit creation is seen.
	OnDepositCreated(
		handler func(depositAddress string),
	) (subscription.EventSubscription, error)

	// OnDepositRegisteredPubkey installs a callback that is invoked when an
	// on-chain notification of a deposit's pubkey registration is seen.
	OnDepositRegisteredPubkey(
		handler func(depositAddress string),
	) (subscription.EventSubscription, error)
}
