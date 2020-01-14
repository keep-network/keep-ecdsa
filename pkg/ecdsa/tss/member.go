package tss

import (
	"math/big"
	"strconv"

	"github.com/keep-network/keep-tecdsa/pkg/net"
)

// MemberID is an unique identifier of a member across the signing group.
// TSS protocol requires that value of the ID is greater than 0.
type MemberID uint32

// String converts MemberID to string.
func (id MemberID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

// bigInt converts MemberID to big.Int.
func (id MemberID) bigInt() *big.Int {
	return new(big.Int).SetBytes(id.bytes())
}

// bytes converts MemberID to bytes slice.
func (id MemberID) bytes() []byte {
	return new(big.Int).SetUint64(uint64(id)).Bytes()
}

// memberIDFromBytes converts bytes slice to MemberID.
func memberIDFromBytes(bytes []byte) MemberID {
	bigInt := new(big.Int).SetBytes(bytes)

	return MemberID(bigInt.Int64())
}

// Equal checks if member IDs are equal.
func (id MemberID) Equal(memberID MemberID) bool {
	return id == memberID
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
	// References from unique MemberID used in protocol to an operator's network
	// layer transport ID. The mapping is used to route unicast messages to an
	// operator's channel as one operator can serve multiple members.
	membersNetworkIDs map[MemberID]net.TransportIdentifier
}
