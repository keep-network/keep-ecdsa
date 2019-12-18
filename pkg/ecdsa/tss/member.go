package tss

import (
	"math/big"
)

// MemberID is an unique identifier of a member across the network.
type MemberID string

// string converts MemberID to string.
func (id MemberID) string() string {
	return string(id)
}

// bigInt converts MemberID to big.Int.
func (id MemberID) bigInt() *big.Int {
	return new(big.Int).SetBytes(id.bytes())
}

// bytes converts MemberID to bytes slice.
func (id MemberID) bytes() []byte {
	return []byte(id)
}

// memberIDFromBytes converts bytes slice to MemberID.
func memberIDFromBytes(bytes []byte) MemberID {
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
