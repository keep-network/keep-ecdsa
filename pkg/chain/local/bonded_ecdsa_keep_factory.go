package local

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

func (c *localChain) createKeep(keepAddress common.Address) error {
	return c.createKeepWithMembers(keepAddress, []common.Address{})
}

func (c *localChain) createKeepWithMembers(
	keepAddress common.Address,
	members []common.Address,
) error {
	c.localChainMutex.Lock()
	defer c.localChainMutex.Unlock()

	if _, ok := c.keeps[keepAddress]; ok {
		return fmt.Errorf(
			"keep already exists for address [%s]",
			keepAddress.String(),
		)
	}

	localKeep := &localKeep{
		chain:                      c,
		keepID:                     keepAddress,
		publicKey:                  [64]byte{},
		members:                    members,
		signatureRequestedHandlers: make(map[int]func(event *chain.SignatureRequestedEvent)),
		keepClosedHandlers:         make(map[int]func(event *chain.KeepClosedEvent)),
		keepTerminatedHandlers:     make(map[int]func(event *chain.KeepTerminatedEvent)),
		signatureSubmittedEvents:   make([]*chain.SignatureSubmittedEvent, 0),
	}

	c.keeps[keepAddress] = localKeep
	c.keepAddresses = append(c.keepAddresses, keepAddress)

	// Ignore errors as the local version never errors.
	operatorIndex := localKeep.unsafeOperatorIndex()

	keepCreatedEvent := &chain.BondedECDSAKeepCreatedEvent{
		Keep:                 localKeep,
		ThisOperatorIsMember: operatorIndex > -1,
	}

	for _, handler := range c.keepCreatedHandlers {
		go func(
			handler func(event *chain.BondedECDSAKeepCreatedEvent),
			keepCreatedEvent *chain.BondedECDSAKeepCreatedEvent,
		) {
			handler(keepCreatedEvent)
		}(handler, keepCreatedEvent)
	}

	return nil
}
