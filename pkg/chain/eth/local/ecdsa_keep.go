package local

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

type localKeep struct {
	publicKey [64]byte

	signatureRequestedHandlers map[int]func(event *eth.SignatureRequestedEvent)

	signaturesMutex *sync.RWMutex
	signatures      map[[32]byte][]*ecdsa.Signature
}

func (c *LocalChain) requestSignature(keepAddress common.Address, digest [32]byte) error {
	c.handlerMutex.Lock()
	defer c.handlerMutex.Unlock()

	keep, ok := keeps[keepAddress]
	if !ok {
		return fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
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

func (lc *LocalChain) GetKeepPublicKey(keepAddress eth.KeepAddress) ([64]byte, error) {
	keepsMutex.RLock()
	defer keepsMutex.RUnlock()

	keep, ok := keeps[keepAddress]
	if !ok {
		return [64]byte{}, fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	return keep.publicKey, nil
}
