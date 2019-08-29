package local

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/sign"
)

// localChain is an implementation of ethereum blockchain interface.
//
// It mocks the behaviour of a real blockchain, without the complexity of deployments,
// accounts, async transactions and so on. For use in tests ONLY.
type localChain struct {
	handlerMutex sync.Mutex

	keeps map[eth.KeepAddress]*localKeep

	keepCreatedHandlers map[int]func(event *eth.ECDSAKeepCreatedEvent)
}

// Connect performs initialization for communication with Ethereum blockchain
// based on provided config.
func Connect() eth.Interface {
	return &localChain{
		keeps:               make(map[eth.KeepAddress]*localKeep),
		keepCreatedHandlers: make(map[int]func(event *eth.ECDSAKeepCreatedEvent)),
	}
}

// OnECDSAKeepCreated is a callback that is invoked when an on-chain
// notification of a new ECDSA keep creation is seen.
func (lc *localChain) OnECDSAKeepCreated(
	handler func(event *eth.ECDSAKeepCreatedEvent),
) (subscription.EventSubscription, error) {
	lc.handlerMutex.Lock()
	defer lc.handlerMutex.Unlock()

	handlerID := rand.Int()

	lc.keepCreatedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lc.handlerMutex.Lock()
		defer lc.handlerMutex.Unlock()

		delete(lc.keepCreatedHandlers, handlerID)
	}), nil
}

// OnSignatureRequested is a callback that is invoked on-chain
// when a keep's signature is requested.
func (lc *localChain) OnSignatureRequested(
	keepAddress eth.KeepAddress,
	handler func(event *eth.SignatureRequestedEvent),
) (subscription.EventSubscription, error) {
	lc.handlerMutex.Lock()
	defer lc.handlerMutex.Unlock()

	handlerID := rand.Int()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return nil, fmt.Errorf(
			"keep not found for address [%s]",
			keepAddress.String(),
		)
	}

	keep.signatureRequestedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lc.handlerMutex.Lock()
		defer lc.handlerMutex.Unlock()

		delete(keep.signatureRequestedHandlers, handlerID)
	}), nil
}

// SubmitKeepPublicKey checks if public key has been already submitted for given
// keep address, if not it stores the key in a map.
func (lc *localChain) SubmitKeepPublicKey(
	keepAddress eth.KeepAddress,
	publicKey [64]byte,
) error {
	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return fmt.Errorf(
			"keep not found for address [%s]",
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
	keepAddress eth.KeepAddress,
	digest [32]byte,
	signature *sign.Signature,
) error {
	return nil
}
