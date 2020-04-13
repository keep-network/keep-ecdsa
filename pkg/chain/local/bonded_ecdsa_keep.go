package local

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

type localKeep struct {
	publicKey [64]byte

	signatureRequestedHandlers map[int]func(event *chain.SignatureRequestedEvent)
}

func (c *localChain) requestSignature(keepAddress common.Address, digest [32]byte) error {
	c.handlerMutex.Lock()
	defer c.handlerMutex.Unlock()

	keep, ok := c.keeps[keepAddress]
	if !ok {
		return fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	signatureRequestedEvent := &chain.SignatureRequestedEvent{
		Digest: digest,
	}

	for _, handler := range keep.signatureRequestedHandlers {
		go func(handler func(event *chain.SignatureRequestedEvent), signatureRequestedEvent *chain.SignatureRequestedEvent) {
			handler(signatureRequestedEvent)
		}(handler, signatureRequestedEvent)
	}

	return nil
}
