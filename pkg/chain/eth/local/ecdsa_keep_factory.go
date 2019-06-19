package ethereum

import (
	"fmt"
	"sync"

	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/gen/abi"
)

func (c *localChain) createKeep(keepAddress common.Address) {
	c.handlerMutex.Lock()

	// keepAddress
	for _, handler := range c.keepCreatedHandlers {
		go func(handler func(event *eth.ECDSAKeepCreatedEvent) {
			handler(event)
		}(handler, event)
	}
	c.handlerMutex.Unlock()
}