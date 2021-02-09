package local

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
	chain "github.com/keep-network/keep-ecdsa/pkg/chain"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
)

func (lc *localChain) BondedECDSAKeepManager() (chain.BondedECDSAKeepManager, error) {
	return lc, nil
}

// OnBondedECDSAKeepCreated is a callback that is invoked when an on-chain
// notification of a new ECDSA keep creation is seen.
func (lc *localChain) OnBondedECDSAKeepCreated(
	handler func(event *eth.BondedECDSAKeepCreatedEvent),
) subscription.EventSubscription {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	handlerID := generateHandlerID()

	lc.keepCreatedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lc.localChainMutex.Lock()
		defer lc.localChainMutex.Unlock()

		delete(lc.keepCreatedHandlers, handlerID)
	})
}

func (lc *localChain) GetKeepCount() (*big.Int, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	return big.NewInt(int64(len(lc.keeps))), nil
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

	return lc.GetKeepWithID(lc.keepAddresses[index])
}

// GetKeepWithID returns a handle to the BondedECDSAKeep with the provided id.
func (lc *localChain) GetKeepWithID(
	keepID chain.KeepID,
) (chain.BondedECDSAKeepHandle, error) {
	// Inside the Ethereum chain, keep ids are always addresses.
	keepAddressString := keepID.String()
	if !common.IsHexAddress(keepAddressString) {
		return nil, fmt.Errorf("incorrect keep address [%s]", keepAddressString)
	}

	return lc.keeps[common.HexToAddress(keepAddressString)], nil
}

func (lc *localChain) createKeep(keepAddress common.Address) (*localKeep, error) {
	return lc.createKeepWithMembers(keepAddress, []common.Address{})
}

func (lc *localChain) createKeepWithMembers(
	keepAddress common.Address,
	members []common.Address,
) (*localKeep, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	if _, ok := lc.keeps[keepAddress]; ok {
		return nil, fmt.Errorf(
			"keep already exists for address [%s]",
			keepAddress.String(),
		)
	}

	localKeep := &localKeep{
		handle:                     lc,
		address:                    keepAddress,
		publicKey:                  [64]byte{},
		members:                    members,
		signatureRequestedHandlers: make(map[int]func(event *chain.SignatureRequestedEvent)),
		keepClosedHandlers:         make(map[int]func(event *chain.KeepClosedEvent)),
		keepTerminatedHandlers:     make(map[int]func(event *chain.KeepTerminatedEvent)),
		signatureSubmittedEvents:   make([]*chain.SignatureSubmittedEvent, 0),
	}

	lc.keeps[keepAddress] = localKeep
	lc.keepAddresses = append(lc.keepAddresses, keepAddress)

	keepCreatedEvent := &chain.BondedECDSAKeepCreatedEvent{
		KeepAddress: keepAddress,
	}

	for _, handler := range lc.keepCreatedHandlers {
		go func(
			handler func(event *chain.BondedECDSAKeepCreatedEvent),
			keepCreatedEvent *chain.BondedECDSAKeepCreatedEvent,
		) {
			handler(keepCreatedEvent)
		}(handler, keepCreatedEvent)
	}

	return localKeep, nil
}
