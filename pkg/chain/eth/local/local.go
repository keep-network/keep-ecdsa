package local

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

// LocalChain is an implementation of ethereum blockchain interface.
//
// It mocks the behavior of a real blockchain, without the complexity of deployments,
// accounts, async transactions and so on. For use in tests ONLY.
type LocalChain struct {
	handlerMutex sync.Mutex

	keeps            map[eth.KeepAddress]*localKeep
	memberCandidates []common.Address
	signatures       map[string]*ecdsa.Signature

	keepCreatedHandlers map[int]func(event *eth.ECDSAKeepCreatedEvent)

	clientAddress common.Address
}

// Connect performs initialization for communication with Ethereum blockchain
// based on provided config.
func Connect() eth.Handle {
	return &LocalChain{
		keeps:               make(map[eth.KeepAddress]*localKeep),
		memberCandidates:    []common.Address{},
		signatures:          make(map[string]*ecdsa.Signature),
		keepCreatedHandlers: make(map[int]func(event *eth.ECDSAKeepCreatedEvent)),
		clientAddress:       common.HexToAddress("6299496199d99941193Fdd2d717ef585F431eA05"),
	}
}

// Address returns client's ethereum address.
func (lc *LocalChain) Address() common.Address {
	return lc.clientAddress
}

// RegisterAsMemberCandidate registers client as a candidate to be selected
// to a keep.
func (lc *LocalChain) RegisterAsMemberCandidate() error {
	lc.memberCandidates = append(lc.memberCandidates, lc.Address())
	return nil
}

// GetMemberCandidates returns list of registered candidates for keep members
// selection.
func (lc *LocalChain) GetMemberCandidates() []common.Address {
	return lc.memberCandidates
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
	keepAddress eth.KeepAddress,
	handler func(event *eth.SignatureRequestedEvent),
) (subscription.EventSubscription, error) {
	lc.handlerMutex.Lock()
	defer lc.handlerMutex.Unlock()

	handlerID := rand.Int()

	keep, ok := lc.keeps[keepAddress]
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

// SubmitKeepPublicKey checks if public key has been already submitted for given
// keep address, if not it stores the key in a map.
func (lc *LocalChain) SubmitKeepPublicKey(
	keepAddress eth.KeepAddress,
	publicKey [64]byte,
) error {
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

// GetKeepPublicKey returns a public key submitted for the keep.
func (lc *LocalChain) GetKeepPublicKey(keepAddress eth.KeepAddress) ([64]byte, error) {
	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return [64]byte{}, fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	return keep.publicKey, nil
}

// SubmitSignature submits a signature to a keep contract deployed under a
// given address.
func (lc *LocalChain) SubmitSignature(
	keepAddress eth.KeepAddress,
	digest [32]byte,
	signature *ecdsa.Signature,
) error {
	key := keepAddress.String() + string(digest[:])
	lc.signatures[key] = signature
	return nil
}

// GetSignature returns a signature submitted to keep for given digest.
func (lc *LocalChain) GetSignature(
	keepAddress eth.KeepAddress,
	digest [32]byte,
) (*ecdsa.Signature, error) {
	key := keepAddress.Hex() + string(digest[:])
	return lc.signatures[key], nil
}
