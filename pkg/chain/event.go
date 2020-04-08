package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// BondedECDSAKeepCreatedEvent is an event emitted on a new keep creation.
type BondedECDSAKeepCreatedEvent struct {
	KeepAddress common.Address   // keep contract address
	Members     []common.Address // keep members addresses
}

// ConflictingPublicKeySubmittedEvent is an event emitted each time when one of
// the members of a keep has submitted a key that does not match the keys submitted
// so far by other members.
type ConflictingPublicKeySubmittedEvent struct {
	SubmittingMember     common.Address
	ConflictingPublicKey []byte
}

// PublicKeyPublishedEvent is an event emitted once all the members have submitted
// the same public key and it was accepted by keep as its public key.
type PublicKeyPublishedEvent struct {
	PublicKey []byte
}

// SignatureRequestedEvent is an event emitted when a user requests
// a digest to be signed.
type SignatureRequestedEvent struct {
	Digest      [32]byte
	BlockNumber uint64
}

// KeepClosedEvent is an event emitted when a keep has been closed.
type KeepClosedEvent struct {
	BlockNumber uint64
}

// KeepTerminatedEvent is an event emitted when a keep has been terminated.
type KeepTerminatedEvent struct {
	BlockNumber uint64
}

// UnbondedValueWithdrawnEvent is an event emitted when unbonded value has been
// withdrawn for an operator.
type UnbondedValueWithdrawnEvent struct {
	Operator common.Address
	Amount   *big.Int
}

// BondCreatedEvent is an event emitted a bond has been created.
type BondCreatedEvent struct {
	Operator    common.Address
	Holder      common.Address
	SignerPool  common.Address
	ReferenceID *big.Int
	Amount      *big.Int
}

// TokensSlashedEvent is an event emitted a tokens has been slashed for an operator.
type TokensSlashedEvent struct {
	Operator common.Address
	Amount   *big.Int
}

// TokensSeizedEvent is an event emitted a tokens has been seized for an operator.
type TokensSeizedEvent struct {
	Operator common.Address
	Amount   *big.Int
}

// IsMember checks if list of members contains the given address.
func (e *BondedECDSAKeepCreatedEvent) IsMember(address common.Address) bool {
	for _, member := range e.Members {
		if member == address {
			return true
		}
	}
	return false
}
