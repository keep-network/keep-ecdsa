package tss

import (
	"bytes"
	cecdsa "crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
)

// MemberID is an unique identifier of a member across the network.
type MemberID []byte

// MemberIDFromHex converts hexadecimal string to MemberID.
func MemberIDFromHex(id string) (MemberID, error) {
	// Skip `0x` or `0X` prefix.
	if len(id) >= 2 && (id[:2] == "0x" || id[:2] == "0X") {
		id = id[2:]
	}

	if len(id) == 0 {
		return nil, fmt.Errorf("empty string")
	}

	memberID, err := hex.DecodeString(id)
	if err != nil {
		return nil, fmt.Errorf("failed to decode string: [%v]", err)
	}

	return memberID, nil
}

// String converts MemberID to string.
func (id MemberID) String() string {
	return hex.EncodeToString(id)
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
	return MemberID(bytes)
}

// Equal checks if member IDs are equal.
func (id MemberID) Equal(memberID MemberID) bool {
	return bytes.Equal(id, memberID)
}

// groupInfo holds information about the group selected for protocol execution.
type groupInfo struct {
	groupID               string // globally unique group identifier
	memberID              MemberID
	groupMemberIDs        []MemberID
	groupMemberPublicKeys map[string]cecdsa.PublicKey
	// Dishonest threshold `t` defines a maximum number of signers controlled by the
	// adversary such that the adversary still cannot produce a signature. Any subset
	// of `t + 1` players can jointly sign, but any smaller subset cannot.
	dishonestThreshold int
}
