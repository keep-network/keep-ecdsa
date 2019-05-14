package ethereum

import (
	"fmt"
	"log"
	"sync"

	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-tecdsa/pkg/eth/chain/gen/abi"
)

func (ec *EthereumChain) watchGroupRequested(
	success func(event *abi.KeepTECDSAGroupGroupRequested),
	fail func(err error) error,
) (subscription.EventSubscription, error) {
	eventChan := make(chan *abi.KeepTECDSAGroupGroupRequested)

	eventSubscription, err := ec.keepTECDSAGroupContract.WatchGroupRequested(
		nil,
		eventChan,
	)
	if err != nil {
		close(eventChan)
		return nil, fmt.Errorf(
			"could not create watch for KeepTECDSAGroupGroupRequested event: [%v]",
			err,
		)
	}

	var subscriptionMutex = &sync.Mutex{}

	go func() {
		for {
			select {
			case event, subscribed := <-eventChan:
				log.Println("GOT EVENT")
				subscriptionMutex.Lock()
				// if eventChan has been closed, it means we have unsubscribed
				if !subscribed {
					subscriptionMutex.Unlock()
					return
				}
				success(event)
				subscriptionMutex.Unlock()
			case err := <-eventSubscription.Err():
				log.Println("GOT FAILURE")
				fail(err)
				return
			}
		}
	}()
	log.Println("registered watchGroupRequested")
	return subscription.NewEventSubscription(func() {
		log.Println("callback in subscription")
		subscriptionMutex.Lock()
		defer subscriptionMutex.Unlock()

		eventSubscription.Unsubscribe()
		close(eventChan)
	}), nil
}
