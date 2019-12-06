package tss

import (
	"math/big"
)

// MemberID is an unique identifier of a member across the network.
type MemberID string

func (id MemberID) bigInt() *big.Int {
	return new(big.Int).SetBytes([]byte(id))
}

// BaseMember holds base member's information.
type BaseMember struct {
	id           MemberID
	groupMembers []MemberID
}
