package client

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	chain "github.com/keep-network/keep-ecdsa/pkg/chain"
)

// eventDeduplicator decides whether the given event should be handled by the
// client or not.
//
// Event subscription may emit the same event two or more times. The same event
// can be emitted right after it's been emitted for the first time. The same
// event can also be emitted a long time after it's been emitted for the first
// time. It is eventDeduplicator's responsibility to decide whether the given
// event is a duplicate and should be ignored or if it is not a duplicate and
// should be handled.
//
// Four events are supported:
// - key generation request for a new keep,
// - signature request for a keep,
// - keep close request,
// - keep terminate request.
type eventDeduplicator struct {
	keepRegistry keepRegistry
	chain        chain.Handle

	keyGenKeeps         *keyGenKeepTrack
	requestedSignatures *requestedSignaturesTrack
	closingKeeps        *closeKeepTrack
	terminatingKeeps    *terminateKeepTrack
}

type keepRegistry interface {
	HasSigner(keepAddress common.Address) bool
}

func newEventDeduplicator(
	keepRegistry keepRegistry,
	chain chain.Handle,
) *eventDeduplicator {
	keyGenKeeps := &keyGenKeepTrack{
		&keepEventTrack{
			data:  make(map[string]bool),
			mutex: &sync.Mutex{},
		},
	}
	requestedSignatures := &requestedSignaturesTrack{
		data:  make(map[string]map[string]bool),
		mutex: &sync.Mutex{},
	}
	closingKeeps := &closeKeepTrack{
		&keepEventTrack{
			data:  make(map[string]bool),
			mutex: &sync.Mutex{},
		},
	}
	terminatingKeeps := &terminateKeepTrack{
		&keepEventTrack{
			data:  make(map[string]bool),
			mutex: &sync.Mutex{},
		},
	}

	return &eventDeduplicator{
		keepRegistry:        keepRegistry,
		chain:               chain,
		keyGenKeeps:         keyGenKeeps,
		requestedSignatures: requestedSignatures,
		closingKeeps:        closingKeeps,
		terminatingKeeps:    terminatingKeeps,
	}
}

// notifyKeyGenStarted notifies the client wants to start key generation for
// a keep upon receiving an event. It returns boolean indicating whether the
// client should proceed with the execution or ignore the event as a duplicate.
//
// In case the client proceeds with the key generation, it should call
// notifyKeyGenCompleted once the protocol completes, no matter if it failed or
// succeeded.
func (ed *eventDeduplicator) notifyKeyGenStarted(keepAddress common.Address) bool {
	if ed.keyGenKeeps.has(keepAddress) {
		return false
	}

	if ed.keepRegistry.HasSigner(keepAddress) {
		return false
	}

	return ed.keyGenKeeps.add(keepAddress)
}

// notifyKeyGenCompleted should be called once client completed key generation
// protocol, no matter if it succeeded or not.
func (ed *eventDeduplicator) notifyKeyGenCompleted(keepAddress common.Address) {
	ed.keyGenKeeps.remove(keepAddress)
}

// notifySigningStarted notifies the client wants to start signature generation
// for the given keep and digest upon receiving an event. It returns boolean
// indicating whether the client should proceed with the execution or ignore the
// event as a duplicate.
//
// In case the client proceeds with signing, it should call
// notifySigningCompleted once the protocol completes, no matter if it failed or
// succeeded.
func (ed *eventDeduplicator) notifySigningStarted(
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

// notifySigningCompleted should be called once client completed signature
// generation for the given keep and digest, no matter if the protocol succeeded
// or not.
func (ed *eventDeduplicator) notifySigningCompleted(
	keepAddress common.Address,
	digest [32]byte,
) {
	ed.requestedSignatures.remove(keepAddress, digest)
}

// notifyClosingStarted notifies the client wants to close a keep upon receiving
// an event. It returns boolean indicating whether the client should proceed
// with the execution or ignore the event as a duplicate.
//
// In case the client proceeds with closing the keep, it should call
// notifyClosingCompleted once the protocol completes, no matter if it failed or
// succeeded.
func (ed *eventDeduplicator) notifyClosingStarted(keepAddress common.Address) bool {
	if ed.closingKeeps.has(keepAddress) {
		return false
	}

	if !ed.keepRegistry.HasSigner(keepAddress) {
		return false
	}

	return ed.closingKeeps.add(keepAddress)
}

// notifyClosingCompleted should be called once client completed closing
// the keep, no matter if the execution succeeded or failed.
func (ed *eventDeduplicator) notifyClosingCompleted(keepAddress common.Address) {
	ed.closingKeeps.remove(keepAddress)
}

// notifyTerminatingStarted notifies the client wants to terminate a keep upon
// receiving an event. It returns boolean indicating whether the client should
// proceed with the execution or ignore the event as a duplicate.
//
// In case the client proceeds with terminating the keep, it should call
// notifyTerminatingCompleted once the protocol completes, no matter if it
// failed or succeeded.
func (ed *eventDeduplicator) notifyTerminatingStarted(keepAddress common.Address) bool {
	if ed.terminatingKeeps.has(keepAddress) {
		return false
	}

	if !ed.keepRegistry.HasSigner(keepAddress) {
		return false
	}

	return ed.terminatingKeeps.add(keepAddress)
}

// notifyTerminatingCompleted should be called once client completed terminating
// the keep, no matter if the execution succeeded or failed.
func (ed *eventDeduplicator) notifyTerminatingCompleted(keepAddress common.Address) {
	ed.terminatingKeeps.remove(keepAddress)
}
