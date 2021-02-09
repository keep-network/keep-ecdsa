package local

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
	chain "github.com/keep-network/keep-ecdsa/pkg/chain"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-ecdsa/pkg/utils/byteutils"
)

type keepStatus int

const (
	active keepStatus = iota
	closed
	terminated
)

type localKeep struct {
	handle  *localChain
	address common.Address

	publicKey    [64]byte
	members      []common.Address
	status       keepStatus
	latestDigest [32]byte

	signatureRequestedHandlers map[int]func(event *eth.SignatureRequestedEvent)

	keepClosedHandlers     map[int]func(event *eth.KeepClosedEvent)
	keepTerminatedHandlers map[int]func(event *eth.KeepTerminatedEvent)

	signatureSubmittedEvents []*eth.SignatureSubmittedEvent
}

func (lk *localKeep) ID() chain.KeepID {
	return lk.address
}

func (lk *localKeep) RequestSignature(digest [32]byte) error {
	lk.handle.localChainMutex.Lock()
	defer lk.handle.localChainMutex.Unlock()

	// force the right workflow sequence
	if lk.publicKey == [64]byte{} {
		return fmt.Errorf(
			"public key for keep [%s] is not set",
			lk.address.String(),
		)
	}

	lk.latestDigest = digest

	signatureRequestedEvent := &eth.SignatureRequestedEvent{
		Digest: digest,
	}

	for _, handler := range lk.signatureRequestedHandlers {
		go func(handler func(event *eth.SignatureRequestedEvent), signatureRequestedEvent *eth.SignatureRequestedEvent) {
			handler(signatureRequestedEvent)
		}(handler, signatureRequestedEvent)
	}

	return nil
}

func (lk *localKeep) OnKeepClosed(
	handler func(event *eth.KeepClosedEvent),
) (subscription.EventSubscription, error) {
	lk.handle.localChainMutex.Lock()
	defer lk.handle.localChainMutex.Unlock()

	handlerID := generateHandlerID()
	lk.keepClosedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lk.handle.localChainMutex.Lock()
		defer lk.handle.localChainMutex.Unlock()

		delete(lk.keepClosedHandlers, handlerID)
	}), nil
}

// OnSignatureRequested is a callback that is invoked on-chain
// when a keep's signature is requested.
func (lk *localKeep) OnSignatureRequested(
	handler func(event *eth.SignatureRequestedEvent),
) (subscription.EventSubscription, error) {
	lk.handle.localChainMutex.Lock()
	defer lk.handle.localChainMutex.Unlock()

	handlerID := generateHandlerID()

	lk.signatureRequestedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lk.handle.localChainMutex.Lock()
		defer lk.handle.localChainMutex.Unlock()

		delete(lk.signatureRequestedHandlers, handlerID)
	}), nil
}

// SubmitKeepPublicKey checks if public key has been already submitted for given
// keep address, if not it stores the key in a map.
func (lk *localKeep) SubmitKeepPublicKey(
	publicKey [64]byte,
) error {
	lk.handle.localChainMutex.Lock()
	defer lk.handle.localChainMutex.Unlock()

	if lk.publicKey != [64]byte{} {
		return fmt.Errorf(
			"public key already submitted for keep [%s]",
			lk.address.String(),
		)
	}

	lk.publicKey = publicKey

	return nil
}

// SubmitSignature submits a signature to a keep contract deployed under a
// given address.
func (lk *localKeep) SubmitSignature(
	signature *ecdsa.Signature,
) error {
	lk.handle.localChainMutex.Lock()
	defer lk.handle.localChainMutex.Unlock()

	// force the right workflow sequence
	if lk.latestDigest == [32]byte{} {
		return fmt.Errorf(
			"keep [%s] is not awaiting for a signature",
			lk.address.String(),
		)
	}

	rBytes, err := byteutils.BytesTo32Byte(signature.R.Bytes())
	if err != nil {
		return err
	}

	sBytes, err := byteutils.BytesTo32Byte(signature.S.Bytes())
	if err != nil {
		return err
	}

	lk.signatureSubmittedEvents = append(
		lk.signatureSubmittedEvents,
		&eth.SignatureSubmittedEvent{
			Digest:     lk.latestDigest,
			R:          rBytes,
			S:          sBytes,
			RecoveryID: uint8(signature.RecoveryID),
		},
	)

	return nil
}

// IsAwaitingSignature checks if the keep is waiting for a signature to be
// calculated for the given digest.
func (lk *localKeep) IsAwaitingSignature(digest [32]byte) (bool, error) {
	lk.handle.localChainMutex.Lock()
	defer lk.handle.localChainMutex.Unlock()

	return lk.latestDigest != [32]byte{}, nil
}

func (lk *localKeep) GetPublicKey() ([]uint8, error) {
	lk.handle.localChainMutex.Lock()
	defer lk.handle.localChainMutex.Unlock()

	return lk.publicKey[:], nil
}

// IsActive checks for current state of a keep on-chain.
func (lk *localKeep) IsActive() (bool, error) {
	lk.handle.localChainMutex.Lock()
	defer lk.handle.localChainMutex.Unlock()

	return lk.status == active, nil
}

func (lc *localChain) closeKeep(keepAddress common.Address) error {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	keep, ok := lc.keeps[keepAddress]
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

	keepClosedEvent := &eth.KeepClosedEvent{}

	for _, handler := range keep.keepClosedHandlers {
		go func(
			handler func(event *eth.KeepClosedEvent),
			keepClosedEvent *eth.KeepClosedEvent,
		) {
			handler(keepClosedEvent)
		}(handler, keepClosedEvent)
	}

	return nil
}

func (lk *localKeep) OnKeepTerminated(
	handler func(event *eth.KeepTerminatedEvent),
) (subscription.EventSubscription, error) {
	lk.handle.localChainMutex.Lock()
	defer lk.handle.localChainMutex.Unlock()

	handlerID := generateHandlerID()

	lk.keepTerminatedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lk.handle.localChainMutex.Lock()
		defer lk.handle.localChainMutex.Unlock()

		delete(lk.keepTerminatedHandlers, handlerID)
	}), nil
}

func (lk *localKeep) OnConflictingPublicKeySubmitted(
	handler func(event *eth.ConflictingPublicKeySubmittedEvent),
) (subscription.EventSubscription, error) {
	panic("implement")
}

func (lk *localKeep) OnPublicKeyPublished(
	handler func(event *eth.PublicKeyPublishedEvent),
) (subscription.EventSubscription, error) {
	panic("implement")
}

func (lk *localKeep) LatestDigest() ([32]byte, error) {
	panic("implement")
}

func (lk *localKeep) SignatureRequestedBlock(
	digest [32]byte,
) (uint64, error) {
	panic("implement")
}

func (lk *localKeep) GetMembers() ([]common.Address, error) {
	lk.handle.localChainMutex.Lock()
	defer lk.handle.localChainMutex.Unlock()

	return lk.members, nil
}

func (lk *localKeep) GetHonestThreshold() (uint64, error) {
	panic("implement")
}

func (lk *localKeep) GetOpenedTimestamp() (time.Time, error) {
	panic("implement")
}

func (lk *localKeep) PastSignatureSubmittedEvents(
	startBlock uint64,
) ([]*eth.SignatureSubmittedEvent, error) {
	lk.handle.localChainMutex.Lock()
	defer lk.handle.localChainMutex.Unlock()

	return lk.signatureSubmittedEvents, nil
}

func (lc *localChain) terminateKeep(keepAddress common.Address) error {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	keep, ok := lc.keeps[keepAddress]
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

	keepTerminatedEvent := &eth.KeepTerminatedEvent{}

	for _, handler := range keep.keepTerminatedHandlers {
		go func(
			handler func(event *eth.KeepTerminatedEvent),
			keepTerminatedEvent *eth.KeepTerminatedEvent,
		) {
			handler(keepTerminatedEvent)
		}(handler, keepTerminatedEvent)
	}

	return nil
}
