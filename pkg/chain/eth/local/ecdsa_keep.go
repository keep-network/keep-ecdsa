package local

import (
	"github.com/ethereum/go-ethereum/common"
	// "github.com/keep-network/keep-tecdsa/pkg/chain/eth"
)

func (c *localChain) requestSignature(keepAddress common.Address, digest []byte) {
	c.handlerMutex.Lock()
	defer c.handlerMutex.Unlock()
}
