package client

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	chain "github.com/keep-network/keep-ecdsa/pkg/chain"
)

type eventDeduplicator struct {
	keepRegistry keepRegistry
	chain        chain.Handle

	requestedSigners    *requestedSignersTrack
	requestedSignatures *requestedSignaturesTrack
}

type keepRegistry interface {
	HasSigner(keepAddress common.Address) bool
}

func newEventDeduplicator(
	keepRegistry keepRegistry,
	chain chain.Handle,
) *eventDeduplicator {
	requestedSigners := &requestedSignersTrack{
		data:  make(map[string]bool),
		mutex: &sync.Mutex{},
	}
	requestedSignatures := &requestedSignaturesTrack{
		data:  make(map[string]map[string]bool),
		mutex: &sync.Mutex{},
	}

	return &eventDeduplicator{
		keepRegistry:        keepRegistry,
		chain:               chain,
		requestedSigners:    requestedSigners,
		requestedSignatures: requestedSignatures,
	}
}

func (ed *eventDeduplicator) generateKeyIfAllowed(keepAddress common.Address) bool {
	if ed.requestedSigners.has(keepAddress) {
		return false
	}

	if ed.keepRegistry.HasSigner(keepAddress) {
		return false
	}

	return ed.requestedSigners.add(keepAddress)
}

func (ed *eventDeduplicator) notifyKeyGenerated(keepAddress common.Address) {
	ed.requestedSigners.remove(keepAddress)
}

func (ed *eventDeduplicator) signIfAllowed(
	keepAddress common.Address,
	digest [32]byte,
) (bool, error) {
	if ed.requestedSignatures.has(keepAddress, digest) {
		return false, nil
	}

	isAwaiting, err := ed.chain.IsAwaitingSignature(keepAddress, digest)
	if err != nil {
		return false, fmt.Errorf("could not deduplicate events: [%v]", err)
	}

	if !isAwaiting {
		return false, nil
	}

	return ed.requestedSignatures.add(keepAddress, digest), nil
}

func (ed *eventDeduplicator) notifySigningCompleted(
	keepAddress common.Address,
	digest [32]byte,
) {
	ed.requestedSignatures.remove(keepAddress, digest)
}
