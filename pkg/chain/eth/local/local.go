package local

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
)

// LocalChain is an implementation of ethereum blockchain interface.
type LocalChain struct {
	handlerMutex sync.Mutex

	keeps map[string]*eth.KeepPublicKey

	keepCreatedHandler map[int]func(groupRegistration *eth.ECDSAKeepCreatedEvent)
}

// Connect performs initialization for communication with Ethereum blockchain
// based on provided config.
func Connect() eth.Interface {
	return &LocalChain{
		keeps: make(map[string]*eth.KeepPublicKey),
	}
}

// OnECDSAKeepCreated is a callback that is invoked when an on-chain
// notification of a new ECDSA keep creation is seen.
func (lc *LocalChain) OnECDSAKeepCreated(
	handle func(groupRequested *eth.ECDSAKeepCreatedEvent),
) (subscription.EventSubscription, error) {
	lc.handlerMutex.Lock()
	defer lc.handlerMutex.Unlock()

	handlerID := rand.Int()

	lc.keepCreatedHandler[handlerID] = handle

	return subscription.NewEventSubscription(func() {
		lc.handlerMutex.Lock()
		defer lc.handlerMutex.Unlock()

		delete(lc.keepCreatedHandler, handlerID)
	}), nil
}

// SubmitKeepPublicKey checks if public key has been already submitted for given
// keep address, if not it stores the key in a map.
func (lc *LocalChain) SubmitKeepPublicKey(
	address eth.KeepAddress,
	publicKey eth.KeepPublicKey,
) error {
	if _, ok := lc.keeps[address.Hex()]; ok {
		return fmt.Errorf(
			"public key already submitted for keep [%s]",
			address.String(),
		)
	}

	lc.keeps[address.Hex()] = &publicKey

	return nil
}
