package tbtc

import (
	"sync"
)

type tbtcEventsLog struct {
	depositRedemptionRequestedEventStorage *eventsStorage
}

func newTBTCEventsLog() *tbtcEventsLog {
	return &tbtcEventsLog{
		depositRedemptionRequestedEventStorage: newEventsStorage(),
	}
}

func (tel *tbtcEventsLog) logDepositRedemptionRequestedEvent(
	deposit string,
	event *depositRedemptionRequestedEvent,
) {
	tel.depositRedemptionRequestedEventStorage.storeEvent(deposit, event)
}

type eventsStorage struct {
	storage      map[string][]interface{} // <deposit address, events slice>
	storageMutex sync.Mutex
}

func newEventsStorage() *eventsStorage {
	return &eventsStorage{
		storage: make(map[string][]interface{}),
	}
}

func (ms *eventsStorage) storeEvent(deposit string, event interface{}) {
	ms.storageMutex.Lock()
	defer ms.storageMutex.Unlock()

	ms.storage[deposit] = append(ms.storage[deposit], event)
}

func (ms *eventsStorage) getEvents(deposit string) []interface{} {
	ms.storageMutex.Lock()
	defer ms.storageMutex.Unlock()

	events := ms.storage[deposit]

	return events
}
