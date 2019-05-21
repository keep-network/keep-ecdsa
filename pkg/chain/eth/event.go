package eth

import (
	"github.com/ethereum/go-ethereum/common"
)

// ECDSAKeepCreatedEvent is an event emitted on a new keep creation.
type ECDSAKeepCreatedEvent struct {
	KeepAddress common.Address // keep contract address
}
