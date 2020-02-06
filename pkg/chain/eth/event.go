package eth

import (
	cecdsa "crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
)

// ECDSAKeepCreatedEvent is an event emitted on a new keep creation.
type ECDSAKeepCreatedEvent struct {
	KeepAddress       common.Address   // keep contract address
	Members           []common.Address // keep members addresses
	MembersPublicKeys []cecdsa.PublicKey
}

// IsMember checks if list of members contains the given address.
func (e *ECDSAKeepCreatedEvent) IsMember(address common.Address) bool {
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
