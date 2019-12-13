package tss

import (
	"math/big"
)

// MemberID is an unique identifier of a member across the network.
type MemberID string

// BigInt converts MemberID to string.
func (id MemberID) String() string {
	return string(id)
}

// BigInt converts MemberID to big.Int.
func (id MemberID) BigInt() *big.Int {
	return new(big.Int).SetBytes(id.Bytes())
}

// Bytes converts MemberID to bytes slice.
func (id MemberID) Bytes() []byte {
	return []byte(id)
}

// MemberIDFromBytes converts bytes slice to MemberID.
func MemberIDFromBytes(bytes []byte) MemberID {
	return MemberID(string(bytes))
}

// groupInfo holds information about the group selected for protocol execution.
type groupInfo struct {
	groupID        string // globally unique group identifier
	memberID       MemberID
	groupMemberIDs []MemberID
	// Dishonest threshold `t` defines a maximum number of signers controlled by the
	// adversary such that the adversary still cannot produce a signature. Any subset
	// of `t + 1` players can jointly sign, but any smaller subset cannot.
	dishonestThreshold int
}
