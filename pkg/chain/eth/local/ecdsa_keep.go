package local

import (
	"fmt"

	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
)

type localKeep struct {
	publicKey [64]byte

	signatureRequestedHandlers map[int]func(event *eth.SignatureRequestedEvent)
}

func (c *localChain) requestSignature(address eth.KeepAddress, digest []byte) error {
	c.handlerMutex.Lock()
	defer c.handlerMutex.Unlock()

	keep, ok := c.keeps[address]
	if !ok {
		return fmt.Errorf(
			"keep not found for address [%s]",
			address,
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
