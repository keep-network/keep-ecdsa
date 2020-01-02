package local

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

var keepsMutex = &sync.RWMutex{}
var keeps = make(map[common.Address]*localKeep)

// LocalChain is an implementation of ethereum blockchain interface.
//
// It mocks the behaviour of a real blockchain, without the complexity of deployments,
// accounts, async transactions and so on. For use in tests ONLY.
type LocalChain struct {
	handlerMutex sync.Mutex

	keepCreatedHandlers map[int]func(event *eth.ECDSAKeepCreatedEvent)

	clientAddress common.Address
}

// Connect performs initialization for communication with Ethereum blockchain
// based on provided config.
func Connect() eth.Handle {
	return &LocalChain{
		keepCreatedHandlers: make(map[int]func(event *eth.ECDSAKeepCreatedEvent)),
		clientAddress:       common.HexToAddress("6299496199d99941193Fdd2d717ef585F431eA05"),
	}
}

// Connect performs initialization for communication with Ethereum blockchain
// based on provided config.
func ConnectWithKey(publicKey *key.NetworkPublic) eth.Handle {
	return &LocalChain{
		keepCreatedHandlers: make(map[int]func(event *eth.ECDSAKeepCreatedEvent)),
		clientAddress:       common.HexToAddress(key.NetworkPubKeyToEthAddress(publicKey)),
	}
}

// Address returns client's ethereum address.
func (lc *LocalChain) Address() common.Address {
	return lc.clientAddress
}

// RegisterAsMemberCandidate registers client as a candidate to be selected
// to a keep.
func (lc *LocalChain) RegisterAsMemberCandidate() error {
	return nil
}

// OnECDSAKeepCreated is a callback that is invoked when an on-chain
// notification of a new ECDSA keep creation is seen.
func (lc *LocalChain) OnECDSAKeepCreated(
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
func (lc *LocalChain) OnSignatureRequested(
	keepAddress common.Address,
	handler func(event *eth.SignatureRequestedEvent),
) (subscription.EventSubscription, error) {
	lc.handlerMutex.Lock()
	defer lc.handlerMutex.Unlock()

	handlerID := rand.Int()

	keep, ok := keeps[keepAddress]
	if !ok {
		return nil, fmt.Errorf(
			"failed to find keep with address: [%s]",
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

// SubmitKeepPublicKey stores a public key for given keep.
func (lc *LocalChain) SubmitKeepPublicKey(
	keepAddress common.Address,
	publicKey [64]byte,
) error {
	keepsMutex.Lock()
	defer keepsMutex.Unlock()

	keep, ok := keeps[keepAddress]
	if !ok {
		return fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	keep.publicKey = publicKey

	return nil
}

// SubmitSignature submits a signature to a keep contract deployed under a
// given address.
func (lc *LocalChain) SubmitSignature(
	keepAddress common.Address,
	signature *ecdsa.Signature,
) error {
	keepsMutex.Lock()
	defer keepsMutex.Unlock()

	keep, ok := keeps[keepAddress]
	if !ok {
		return fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	keep.signaturesMutex.Lock()
	keep.signatures = append(keep.signatures, signature)
	keep.signaturesMutex.Unlock()

	return nil
}
