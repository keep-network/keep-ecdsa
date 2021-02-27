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

func (lc *localChain) GetKeepAtIndex(
	keepIndex *big.Int,
) (common.Address, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	index := int(keepIndex.Uint64())

	if index > len(lc.keepAddresses) {
		return common.HexToAddress("0x0"), fmt.Errorf("out of bounds")
	}

	return lc.keepAddresses[index], nil
}

// OnSignatureRequested is a callback that is invoked on-chain
// when a keep's signature is requested.
func (lc *localChain) OnSignatureRequested(
	keepAddress common.Address,
	handler func(event *chain.SignatureRequestedEvent),
) (subscription.EventSubscription, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	handlerID := generateHandlerID()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return nil, fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	keep.signatureRequestedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lc.localChainMutex.Lock()
		defer lc.localChainMutex.Unlock()

		delete(keep.signatureRequestedHandlers, handlerID)
	}), nil
}

func (lc *localChain) OnConflictingPublicKeySubmitted(
	keepAddress common.Address,
	handler func(event *chain.ConflictingPublicKeySubmittedEvent),
) (subscription.EventSubscription, error) {
	panic("implement")
}

func (lc *localChain) OnPublicKeyPublished(
	keepAddress common.Address,
	handler func(event *chain.PublicKeyPublishedEvent),
) (subscription.EventSubscription, error) {
	panic("implement")
}

// SubmitKeepPublicKey checks if public key has been already submitted for given
// keep address, if not it stores the key in a map.
func (lc *localChain) SubmitKeepPublicKey(
	keepAddress common.Address,
	publicKey [64]byte,
) error {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	if keep.publicKey != [64]byte{} {
		return fmt.Errorf(
			"public key already submitted for keep [%s]",
			keepAddress.String(),
		)
	}

	keep.publicKey = publicKey

	return nil
}

// SubmitSignature submits a signature to a keep contract deployed under a
// given address.
func (lc *localChain) SubmitSignature(
	keepAddress common.Address,
	signature *ecdsa.Signature,
) error {
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
	if keep.latestDigest == [32]byte{} {
		return fmt.Errorf(
			"keep [%s] is not awaiting for a signature",
			keepAddress.String(),
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

	keep.signatureSubmittedEvents = append(
		keep.signatureSubmittedEvents,
		&chain.SignatureSubmittedEvent{
			Digest:     keep.latestDigest,
			R:          rBytes,
			S:          sBytes,
			RecoveryID: uint8(signature.RecoveryID),
		},
	)

	return nil
}

func (lc *localChain) OnKeepClosed(
	keepAddress common.Address,
	handler func(event *chain.KeepClosedEvent),
) (subscription.EventSubscription, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	handlerID := generateHandlerID()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return nil, fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	keep.keepClosedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lc.localChainMutex.Lock()
		defer lc.localChainMutex.Unlock()

		delete(keep.keepClosedHandlers, handlerID)
	}), nil

}

func (lc *localChain) OnKeepTerminated(
	keepAddress common.Address,
	handler func(event *chain.KeepTerminatedEvent),
) (subscription.EventSubscription, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	handlerID := generateHandlerID()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return nil, fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	keep.keepTerminatedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lc.localChainMutex.Lock()
		defer lc.localChainMutex.Unlock()

		delete(keep.keepTerminatedHandlers, handlerID)
	}), nil
}

// IsAwaitingSignature checks if the keep is waiting for a signature to be
// calculated for the given digest.
func (lc *localChain) IsAwaitingSignature(
	keepAddress common.Address,
	digest [32]byte,
) (bool, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return false, fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	return keep.latestDigest != [32]byte{}, nil
}

// IsActive checks for current state of a keep on-chain.
func (lc *localChain) IsActive(keepAddress common.Address) (bool, error) {

	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return false, fmt.Errorf("no keep with address [%v]", keepAddress)
	}

	return keep.status == active, nil

}

func (lc *localChain) LatestDigest(keepAddress common.Address) ([32]byte, error) {
	panic("implement")
}

func (lc *localChain) SignatureRequestedBlock(
	keepAddress common.Address,
	digest [32]byte,
) (uint64, error) {
	panic("implement")
}

func (lc *localChain) GetPublicKey(keepAddress common.Address) ([]uint8, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return nil, fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	return keep.publicKey[:], nil

}

func (lc *localChain) GetMembers(
	keepAddress common.Address,
) ([]common.Address, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return nil, fmt.Errorf("no keep with address [%v]", keepAddress)

	}
	return keep.members, nil
}

func (lc *localChain) GetHonestThreshold(
	keepAddress common.Address,
) (uint64, error) {
	panic("implement")
}

func (lc *localChain) GetOpenedTimestamp(keepAddress common.Address) (time.Time, error) {
	panic("implement")
}

func (lc *localChain) PastSignatureSubmittedEvents(
	keepAddress string,
	startBlock uint64,
) ([]*chain.SignatureSubmittedEvent, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	keep, ok := lc.keeps[common.HexToAddress(keepAddress)]
	if !ok {
		return nil, fmt.Errorf("no keep with address [%v]", keepAddress)

	}

	return keep.signatureSubmittedEvents, nil
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
