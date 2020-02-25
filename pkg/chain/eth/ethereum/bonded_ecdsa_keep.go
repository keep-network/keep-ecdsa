package ethereum

import (
	"fmt"
	"sync"

	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/gen/abi"
)

func (ec *EthereumChain) watchSignatureRequested(
	keepContract *abi.BondedECDSAKeep,
	success func(event *abi.BondedECDSAKeepSignatureRequested),
	fail func(err error) error,
) (subscription.EventSubscription, error) {
	eventChan := make(chan *abi.BondedECDSAKeepSignatureRequested)

	eventSubscription, err := keepContract.WatchSignatureRequested(
		nil,
		eventChan,
	)
	if err != nil {
		close(eventChan)
		return nil, fmt.Errorf(
			"failed to create watch for BondedECDSAKeepSignatureRequested event: [%v]",
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

func (ec *EthereumChain) watchKeepClosed(
	keepContract *abi.BondedECDSAKeep,
	success func(event *abi.BondedECDSAKeepKeepClosed),
	fail func(err error) error,
) (subscription.EventSubscription, error) {
	eventChan := make(chan *abi.BondedECDSAKeepKeepClosed)

	eventSubscription, err := keepContract.WatchKeepClosed(
		nil,
		eventChan,
	)
	if err != nil {
		close(eventChan)
		return nil, fmt.Errorf(
			"failed to create watch for BondedECDSAKeepClosed event: [%v]",
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
