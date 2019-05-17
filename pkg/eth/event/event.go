// Package event reflects structures of events emitted on an ethereum blockchain.
package event

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// MemberID is unique keep member identifier.
// TODO: Consider changing to different type e.g. address.
type MemberID = big.Int

// ECDSAKeepRequested is an event emitted on a new keep creation request.
type ECDSAKeepRequested struct {
	KeepAddress        common.Address // keep contract address
	MemberIDs          []*MemberID    // keep members IDs
	DishonestThreshold *big.Int       // maximum number of dishonest members `m`
}
