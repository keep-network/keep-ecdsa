package local

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
)

// localChain is an implementation of ethereum blockchain interface.
type localChain struct {
	handlerMutex sync.Mutex

	keeps map[string][64]byte

	keepCreatedHandlers map[int]func(groupRegistration *eth.ECDSAKeepCreatedEvent)
}

// Connect performs initialization for communication with Ethereum blockchain
// based on provided config.
func Connect() eth.Interface {
	return &localChain{
		keeps: make(map[string][64]byte),
	}
}

// OnECDSAKeepCreated is a callback that is invoked when an on-chain
// notification of a new ECDSA keep creation is seen.
func (lc *localChain) OnECDSAKeepCreated(
	handle func(groupRequested *eth.ECDSAKeepCreatedEvent),
) (subscription.EventSubscription, error) {
	lc.handlerMutex.Lock()
	defer lc.handlerMutex.Unlock()

	handlerID := rand.Int()

	lc.keepCreatedHandlers[handlerID] = handle

	return subscription.NewEventSubscription(func() {
		lc.handlerMutex.Lock()
		defer lc.handlerMutex.Unlock()

		delete(lc.keepCreatedHandlers, handlerID)
	}), nil
}

// OnSignatureRequested is a callback that is invoked on-chain
// when a keep's signature is requested.
func (lc *localChain) OnSignatureRequested(
	keepAddress common.Address,
	handle func(event *eth.SignatureRequestedEvent),
) (subscription.EventSubscription, error) {
	return nil, fmt.Errorf("unimplemented: localChain.OnSignatureRequested")
}

// SubmitKeepPublicKey checks if public key has been already submitted for given
// keep address, if not it stores the key in a map.
func (lc *localChain) SubmitKeepPublicKey(
	address eth.KeepAddress,
	publicKey [64]byte,
) error {
	if _, ok := lc.keeps[address.Hex()]; ok {
		return fmt.Errorf(
			"public key already submitted for keep [%s]",
			address.String(),
		)
	}

	lc.keeps[address.Hex()] = publicKey

	return nil
}
