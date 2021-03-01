package local

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

type keepStatus int

const (
	active keepStatus = iota
	closed
	terminated
)

type localKeep struct {
	publicKey    [64]byte
	members      []common.Address
	status       keepStatus
	latestDigest [32]byte

	signatureRequestedHandlers map[int]func(event *chain.SignatureRequestedEvent)

	keepClosedHandlers     map[int]func(event *chain.KeepClosedEvent)
	keepTerminatedHandlers map[int]func(event *chain.KeepTerminatedEvent)

	signatureSubmittedEvents []*chain.SignatureSubmittedEvent
}

func (c *localChain) RequestSignature(keepAddress common.Address, digest [32]byte) error {
	c.localChainMutex.Lock()
	defer c.localChainMutex.Unlock()

	keep, ok := c.keeps[keepAddress]
	if !ok {
		return fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	// force the right workflow sequence
	if keep.publicKey == [64]byte{} {
		return fmt.Errorf(
			"public key for keep [%s] is not set",
			keepAddress.String(),
		)
	}

	keep.latestDigest = digest

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

func (c *localChain) closeKeep(keepAddress common.Address) error {
	c.localChainMutex.Lock()
	defer c.localChainMutex.Unlock()

	keep, ok := c.keeps[keepAddress]
	if !ok {
		return fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	if keep.status != active {
		return fmt.Errorf("only active keeps can be closed")
	}

	keep.status = closed

	keepClosedEvent := &chain.KeepClosedEvent{}

	for _, handler := range keep.keepClosedHandlers {
		go func(
			handler func(event *chain.KeepClosedEvent),
			keepClosedEvent *chain.KeepClosedEvent,
		) {
			handler(keepClosedEvent)
		}(handler, keepClosedEvent)
	}

	return nil
}

func (c *localChain) terminateKeep(keepAddress common.Address) error {
	c.localChainMutex.Lock()
	defer c.localChainMutex.Unlock()

	keep, ok := c.keeps[keepAddress]
	if !ok {
		return fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	if keep.status != active {
		return fmt.Errorf("only active keeps can be terminated")
	}

	keep.status = terminated

	keepTerminatedEvent := &chain.KeepTerminatedEvent{}

	for _, handler := range keep.keepTerminatedHandlers {
		go func(
			handler func(event *chain.KeepTerminatedEvent),
			keepTerminatedEvent *chain.KeepTerminatedEvent,
		) {
			handler(keepTerminatedEvent)
		}(handler, keepTerminatedEvent)
	}

	return nil
}
