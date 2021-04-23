package local

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
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
	chain  *localChain
	keepID common.Address

	publicKey    [64]byte
	members      []common.Address
	status       keepStatus
	latestDigest [32]byte

	signatureRequestedHandlers map[int]func(event *chain.SignatureRequestedEvent)

	keepClosedHandlers     map[int]func(event *chain.KeepClosedEvent)
	keepTerminatedHandlers map[int]func(event *chain.KeepTerminatedEvent)

	signatureSubmittedEvents []*chain.SignatureSubmittedEvent
}

func (lc *localChain) GetKeepWithID(
	keepID chain.ID,
) (chain.BondedECDSAKeepHandle, error) {
	keepAddress, err := fromChainID(keepID)
	if err != nil {
		return nil, err
	}

	return lc.keeps[keepAddress], nil
}

func (lc *localChain) GetKeepAtIndex(
	keepIndex *big.Int,
) (chain.BondedECDSAKeepHandle, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	index := int(keepIndex.Uint64())

	if index > len(lc.keepAddresses) {
		return nil, fmt.Errorf("out of bounds")
	}

	return lc.GetKeepWithID(localChainID(lc.keepAddresses[index]))
}

func (lk *localKeep) ID() chain.ID {
	return localChainID(lk.keepID)
}

// OnSignatureRequested is a callback that is invoked on-chain
// when a keep's signature is requested.
func (lk *localKeep) OnSignatureRequested(
	handler func(event *chain.SignatureRequestedEvent),
) (subscription.EventSubscription, error) {
	lk.chain.localChainMutex.Lock()
	defer lk.chain.localChainMutex.Unlock()

	handlerID := generateHandlerID()

	lk.signatureRequestedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lk.chain.localChainMutex.Lock()
		defer lk.chain.localChainMutex.Unlock()

		delete(lk.signatureRequestedHandlers, handlerID)
	}), nil
}

func (lk *localKeep) OnConflictingPublicKeySubmitted(
	handler func(event *chain.ConflictingPublicKeySubmittedEvent),
) (subscription.EventSubscription, error) {
	panic("implement")
}

func (lk *localKeep) OnPublicKeyPublished(
	handler func(event *chain.PublicKeyPublishedEvent),
) (subscription.EventSubscription, error) {
	panic("implement")
}

// SubmitKeepPublicKey checks if public key has been already submitted for given
// keep address, if not it stores the key in a map.
func (lk *localKeep) SubmitKeepPublicKey(publicKey [64]byte) error {
	lk.chain.localChainMutex.Lock()
	defer lk.chain.localChainMutex.Unlock()

	if lk.publicKey != [64]byte{} {
		return fmt.Errorf(
			"public key already submitted for keep [%s]",
			lk.ID().String(),
		)
	}

	lk.publicKey = publicKey

	return nil
}

// SubmitSignature submits a signature to a keep contract deployed under a
// given address.
func (lk *localKeep) SubmitSignature(signature *ecdsa.Signature) error {
	lk.chain.localChainMutex.Lock()
	defer lk.chain.localChainMutex.Unlock()

	// force the right workflow sequence
	if lk.latestDigest == [32]byte{} {
		return fmt.Errorf(
			"keep [%s] is not awaiting for a signature",
			lk.ID().String(),
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
		&chain.SignatureSubmittedEvent{
			Digest:     lk.latestDigest,
			R:          rBytes,
			S:          sBytes,
			RecoveryID: uint8(signature.RecoveryID),
		},
	)

	return nil
}

func (lk *localKeep) OnKeepClosed(
	handler func(event *chain.KeepClosedEvent),
) (subscription.EventSubscription, error) {
	lk.chain.localChainMutex.Lock()
	defer lk.chain.localChainMutex.Unlock()

	handlerID := generateHandlerID()

	lk.keepClosedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lk.chain.localChainMutex.Lock()
		defer lk.chain.localChainMutex.Unlock()

		delete(lk.keepClosedHandlers, handlerID)
	}), nil
}

func (lk *localKeep) OnKeepTerminated(
	handler func(event *chain.KeepTerminatedEvent),
) (subscription.EventSubscription, error) {
	lk.chain.localChainMutex.Lock()
	defer lk.chain.localChainMutex.Unlock()

	handlerID := generateHandlerID()

	lk.keepTerminatedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lk.chain.localChainMutex.Lock()
		defer lk.chain.localChainMutex.Unlock()

		delete(lk.keepTerminatedHandlers, handlerID)
	}), nil
}

// IsAwaitingSignature checks if the keep is waiting for a signature to be
// calculated for the given digest.
func (lk *localKeep) IsAwaitingSignature(digest [32]byte) (bool, error) {
	lk.chain.localChainMutex.Lock()
	defer lk.chain.localChainMutex.Unlock()

	return lk.latestDigest != [32]byte{}, nil
}

// IsActive checks for current state of a keep on-chain.
func (lk *localKeep) IsActive() (bool, error) {
	lk.chain.localChainMutex.Lock()
	defer lk.chain.localChainMutex.Unlock()

	return lk.status == active, nil
}

func (lk *localKeep) LatestDigest() ([32]byte, error) {
	panic("implement")
}

func (lk *localKeep) SignatureRequestedBlock(digest [32]byte) (uint64, error) {
	panic("implement")
}

func (lk *localKeep) GetPublicKey() ([]uint8, error) {
	lk.chain.localChainMutex.Lock()
	defer lk.chain.localChainMutex.Unlock()

	return lk.publicKey[:], nil
}

func (lk *localKeep) IsThisOperatorMember() (bool, error) {
	operatorIndex, err := lk.OperatorIndex()
	if err != nil {
		return false, err
	}

	return operatorIndex != -1, nil
}

// Unsafe version of OperatorIndex that does no mutex locking. This is for use
// with callers who have already locked the mutex.
func (lk *localKeep) unsafeOperatorIndex() int {
	operatorAddress := lk.chain.OperatorAddress()

	for i, memberID := range lk.members {
		if operatorAddress.String() == memberID.String() {
			return i
		}
	}

	return -1
}

func (lk *localKeep) OperatorIndex() (int, error) {
	lk.chain.localChainMutex.Lock()
	defer lk.chain.localChainMutex.Unlock()

	operatorAddress := lk.chain.OperatorAddress()

	for i, memberID := range lk.members {
		if operatorAddress.String() == memberID.String() {
			return i, nil
		}
	}

	return -1, nil
}

func (lk *localKeep) GetMembers() ([]chain.ID, error) {
	lk.chain.localChainMutex.Lock()
	defer lk.chain.localChainMutex.Unlock()

	return toIDSlice(lk.members), nil
}

func (lk *localKeep) GetHonestThreshold() (uint64, error) {
	panic("implement")
}

func (lk *localKeep) GetOpenedTimestamp() (time.Time, error) {
	panic("implement")
}

func (lk *localKeep) PastSignatureSubmittedEvents(
	startBlock uint64,
) ([]*chain.SignatureSubmittedEvent, error) {
	lk.chain.localChainMutex.Lock()
	defer lk.chain.localChainMutex.Unlock()

	return lk.signatureSubmittedEvents, nil
}

func (lc *localChain) RequestSignature(keepAddress common.Address, digest [32]byte) error {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	keep, ok := lc.keeps[keepAddress]
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
