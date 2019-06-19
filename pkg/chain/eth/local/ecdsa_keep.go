package local

import (
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"

	"fmt"
)

func (c *localChain) requestSignature(address eth.KeepAddress, digest []byte) error {
	c.handlerMutex.Lock()
	defer c.handlerMutex.Unlock()

	keep, ok := c.keeps[address]
	if !ok {
		return fmt.Errorf(
			"keep not found for address [%s]",
			address.String(),
		)
	}

	signatureRequestedEvent := &eth.SignatureRequestedEvent{
		Digest: digest,
	}

	for _, handler := range keep.signatureRequestedHandlers {
		go func(handler func(event *eth.SignatureRequestedEvent), signatureRequestedEvent *eth.SignatureRequestedEvent) {
			handler(signatureRequestedEvent)
		}(handler, signatureRequestedEvent)
	}

	return nil
}
