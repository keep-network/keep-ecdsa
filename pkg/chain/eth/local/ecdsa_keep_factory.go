package local

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

func (c *LocalChain) CreateKeep(keepAddress common.Address) {
	c.handlerMutex.Lock()
	defer c.handlerMutex.Unlock()

	if _, ok := keeps[keepAddress]; ok {
		return
	}

	localKeep := &localKeep{
		signatureRequestedHandlers: make(map[int]func(event *eth.SignatureRequestedEvent)),
		publicKey:                  [64]byte{},
		signaturesMutex:            &sync.RWMutex{},
		signatures:                 make(map[[32]byte][]*ecdsa.Signature),
	}
	keeps[keepAddress] = localKeep

	keepCreatedEvent := &eth.ECDSAKeepCreatedEvent{
		KeepAddress: keepAddress,
	}

	for _, handler := range c.keepCreatedHandlers {
		go func(handler func(event *eth.ECDSAKeepCreatedEvent), keepCreatedEvent *eth.ECDSAKeepCreatedEvent) {
			handler(keepCreatedEvent)
		}(handler, keepCreatedEvent)
	}
}
