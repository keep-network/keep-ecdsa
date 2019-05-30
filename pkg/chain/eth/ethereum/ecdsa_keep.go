package ethereum

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/gen/abi"
)

func (ec *EthereumChain) registerKeepContract(address common.Address) error {
	ecdsaKeepContract, err := abi.NewECDSAKeep(
		address,
		ec.client,
	)
	if err != nil {
		return err
	}

	if _, ok := ec.keepContracts[address]; ok {
		return fmt.Errorf("keep already registered: [%v]", address)
	}
	ec.keepContracts[address] = ecdsaKeepContract

	return nil
}

func (ec *EthereumChain) watchECDSAKeepSignatureRequested(
	keepContract *abi.ECDSAKeep,
	success func(event *abi.ECDSAKeepSignatureRequested),
	fail func(err error) error,
) (subscription.EventSubscription, error) {
	eventChan := make(chan *abi.ECDSAKeepSignatureRequested)

	eventSubscription, err := keepContract.WatchSignatureRequested(
		nil,
		eventChan,
	)
	if err != nil {
		close(eventChan)
		return nil, fmt.Errorf(
			"could not create watch for ECDSAKeepSignatureRequested event: [%v]",
			err,
		)
	}

	var subscriptionMutex = &sync.Mutex{}

	go func() {
		for {
			select {
			case event, subscribed := <-eventChan:
				subscriptionMutex.Lock()
				// if eventChan has been closed, it means we have unsubscribed
				if !subscribed {
					subscriptionMutex.Unlock()
					return
				}
				success(event)
				subscriptionMutex.Unlock()
			case err := <-eventSubscription.Err():
				fail(err)
				return
			}
		}
	}()

	return subscription.NewEventSubscription(func() {
		subscriptionMutex.Lock()
		defer subscriptionMutex.Unlock()

		eventSubscription.Unsubscribe()
		close(eventChan)
	}), nil
}
