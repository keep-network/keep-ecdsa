package eth

import (
	"github.com/ethereum/go-ethereum/common"
)

// BondedECDSAKeepCreatedEvent is an event emitted on a new keep creation.
type BondedECDSAKeepCreatedEvent struct {
	KeepAddress common.Address   // keep contract address
	Members     []common.Address // keep members addresses
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

// SignatureRequestedEvent is an event emitted when a user requests
// a digest to be signed.
type SignatureRequestedEvent struct {
	Digest [32]byte
}

// ConflictingPublicKeySubmittedEvent is an event emitted each time when one of
// the members of a keep has submitted a key that does not match existing one.
type ConflictingPublicKeySubmittedEvent struct {
	SubmittingMember common.Address
	ConflictingPublicKey []byte
}

// PublicKeyPublishedEvent is an event emitted once all the members have submitted
// the same public key and it was accepted by keep as its public key.
type PublicKeyPublishedEvent struct {
	PublicKey []byte
}