package local

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

func (c *localChain) createKeep(keepAddress common.Address) error {
	c.handlerMutex.Lock()
	defer c.handlerMutex.Unlock()

	if _, ok := c.keeps[keepAddress]; ok {
		return fmt.Errorf(
			"keep already exists for address [%s]",
			keepAddress.String(),
		)
	}

	localKeep := &localKeep{
		signatureRequestedHandlers: make(map[int]func(event *chain.SignatureRequestedEvent)),
		publicKey:                  [64]byte{},
	}
	c.keeps[keepAddress] = localKeep

	keepCreatedEvent := &chain.BondedECDSAKeepCreatedEvent{
		KeepAddress: keepAddress,
	}

	for _, handler := range c.keepCreatedHandlers {
		go func(handler func(event *chain.BondedECDSAKeepCreatedEvent), keepCreatedEvent *chain.BondedECDSAKeepCreatedEvent) {
			handler(keepCreatedEvent)
		}(handler, keepCreatedEvent)
	}

	return nil
}
