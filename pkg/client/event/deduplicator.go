package event

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	chain "github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/utils"
)

// Deduplicator decides whether the given event should be handled by the
// client or not.
//
// Event subscription may emit the same event two or more times. The same event
// can be emitted right after it's been emitted for the first time. The same
// event can also be emitted a long time after it's been emitted for the first
// time. It is deduplicator's responsibility to decide whether the given
// event is a duplicate and should be ignored or if it is not a duplicate and
// should be handled.
//
// Four events are supported:
// - key generation request for a new keep,
// - signature request for a keep,
// - keep close request,
// - keep terminate request.
type Deduplicator struct {
	keepRegistry keepRegistry
	chain        chain.Handle

	keyGenKeeps         *uniqueEventTrack
	requestedSignatures *requestedSignaturesTrack
	closingKeeps        *uniqueEventTrack
	terminatingKeeps    *uniqueEventTrack
}

type keepRegistry interface {
	// HasSigner returns true if keep with the given address already exists
	// in the registry. In the context of event deduplicator, it means that
	// the key for the given keep has already been generated.
	HasSigner(keepAddress common.Address) bool
}

func NewDeduplicator(
	keepRegistry keepRegistry,
	chain chain.Handle,
) *Deduplicator {
	keyGenKeeps := &uniqueEventTrack{
		data: make(map[string]bool),
	}
	requestedSignatures := &requestedSignaturesTrack{
		data: make(map[string]map[string]bool),
	}
	closingKeeps := &uniqueEventTrack{
		data: make(map[string]bool),
	}
	terminatingKeeps := &uniqueEventTrack{
		data: make(map[string]bool),
	}

	return &Deduplicator{
		keepRegistry:        keepRegistry,
		chain:               chain,
		keyGenKeeps:         keyGenKeeps,
		requestedSignatures: requestedSignatures,
		closingKeeps:        closingKeeps,
		terminatingKeeps:    terminatingKeeps,
	}
}

// NotifyKeyGenStarted notifies the client wants to start key generation for
// a keep upon receiving an event. It returns boolean indicating whether the
// client should proceed with the execution or ignore the event as a duplicate.
//
// In case the client proceeds with the key generation, it should call
// NotifyKeyGenCompleted once the protocol completes, no matter if it failed or
// succeeded.
func (d *Deduplicator) NotifyKeyGenStarted(keepAddress common.Address) bool {
	if d.keyGenKeeps.has(keepAddress) {
		return false
	}

	// If the event is not currently being handled but keep with the given
	// address already exists in the registry, the event should be rejected
	// as a duplicate. It is an old event that has already been handled.
	if d.keepRegistry.HasSigner(keepAddress) {
		return false
	}

	return d.keyGenKeeps.add(keepAddress)
}

// NotifyKeyGenCompleted should be called once client completed key generation
// protocol, no matter if it succeeded or not.
func (d *Deduplicator) NotifyKeyGenCompleted(keepAddress common.Address) {
	d.keyGenKeeps.remove(keepAddress)
}

// NotifySigningStarted notifies the client wants to start signature generation
// for the given keep and digest upon receiving an event. It returns boolean
// indicating whether the client should proceed with the execution or ignore the
// event as a duplicate.
//
// In case the client proceeds with signing, it should call
// NotifySigningCompleted once the protocol completes, no matter if it failed or
// succeeded.
func (d *Deduplicator) NotifySigningStarted(
	keepAddress common.Address,
	digest [32]byte,
) (bool, error) {
	if d.requestedSignatures.has(keepAddress, digest) {
		return false, nil
	}

	// If the event is not currently being handled, we need to confirm on-chain
	// if signing is pending. The event could be an old one that has already
	// been handled.
	// Repeat the check in case of a small chain reorg or if chain nodes
	// are out of sync.
	isAwaitingSignature, err := utils.ConfirmWithTimeout(
		10*time.Second,
		10*time.Second,
		30*time.Second,
		func(ctx context.Context) (bool, error) {
			return d.chain.IsAwaitingSignature(keepAddress, digest)
		},
	)
	if err != nil {
		return false, fmt.Errorf(
			"could not check if keep is awaiting for a signature "+
				"when deduplicating events: [%v]",
			err,
		)
	}

	if !isAwaitingSignature {
		return false, nil
	}

	return d.requestedSignatures.add(keepAddress, digest), nil
}

// NotifySigningCompleted should be called once client completed signature
// generation for the given keep and digest, no matter if the protocol succeeded
// or not.
func (d *Deduplicator) NotifySigningCompleted(
	keepAddress common.Address,
	digest [32]byte,
) {
	d.requestedSignatures.remove(keepAddress, digest)
}

// NotifyClosingStarted notifies the client wants to close a keep upon receiving
// an event. It returns boolean indicating whether the client should proceed
// with the execution or ignore the event as a duplicate.
//
// In case the client proceeds with closing the keep, it should call
// NotifyClosingCompleted once the protocol completes, no matter if it failed or
// succeeded.
func (d *Deduplicator) NotifyClosingStarted(keepAddress common.Address) bool {
	if d.closingKeeps.has(keepAddress) {
		return false
	}

	// If the event is not currently being handled but keep with the given
	// address does no longer exist in the registry, the event should be
	// rejected as a duplicate. It is an old event that has already been
	// handled.
	if !d.keepRegistry.HasSigner(keepAddress) {
		return false
	}

	return d.closingKeeps.add(keepAddress)
}

// NotifyClosingCompleted should be called once client completed closing
// the keep, no matter if the execution succeeded or failed.
func (d *Deduplicator) NotifyClosingCompleted(keepAddress common.Address) {
	d.closingKeeps.remove(keepAddress)
}

// NotifyTerminatingStarted notifies the client wants to terminate a keep upon
// receiving an event. It returns boolean indicating whether the client should
// proceed with the execution or ignore the event as a duplicate.
//
// In case the client proceeds with terminating the keep, it should call
// NotifyTerminatingCompleted once the protocol completes, no matter if it
// failed or succeeded.
func (d *Deduplicator) NotifyTerminatingStarted(keepAddress common.Address) bool {
	if d.terminatingKeeps.has(keepAddress) {
		return false
	}

	// If the event is not currently being handled but keep with the given
	// address does no longer exist in the registry, the event should be
	// rejected as a duplicate. It is an old event that has already been
	// handled.
	if !d.keepRegistry.HasSigner(keepAddress) {
		return false
	}

	return d.terminatingKeeps.add(keepAddress)
}

// NotifyTerminatingCompleted should be called once client completed terminating
// the keep, no matter if the execution succeeded or failed.
func (d *Deduplicator) NotifyTerminatingCompleted(keepAddress common.Address) {
	d.terminatingKeeps.remove(keepAddress)
}
