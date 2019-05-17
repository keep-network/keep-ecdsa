// Package event reflects structures of events emitted on an ethereum blockchain.
package event

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// MemberID is unique keep member identifier.
// TODO: Consider changing to different type e.g. address.
type MemberID = big.Int

// GroupRequested is an event emitted on new group creation request.
type GroupRequested struct { // TODO: Remove, it's replaced by ECDSAKeepRequested
	RequestID          *big.Int
	GroupID            *big.Int // currently single Signer ID
	GroupSize          uint32   // n
	DishonestThreshold uint32   // m
}

// ECDSAKeepRequested is an event emitted on a new keep creation request.
type ECDSAKeepRequested struct {
	KeepAddress        common.Address // keep contract address
	MemberIDs          []*MemberID    // keep members IDs
	DishonestThreshold *big.Int       // maximum number of dishonest members `m`
}
