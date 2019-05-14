package event

import "math/big"

// GroupRequested is an event emitted on new group creation request.
type GroupRequested struct {
	RequestID          *big.Int
	GroupID            *big.Int // currently single Signer ID
	GroupSize          uint32   // n
	DishonestThreshold uint32   // m
}
