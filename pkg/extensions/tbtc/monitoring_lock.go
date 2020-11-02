package tbtc

import (
	"fmt"
	"strings"
	"sync"
)

type monitoringLockKey string

func newMonitoringLockKey(
	depositAddress string,
	monitoringName string,
) monitoringLockKey {
	return monitoringLockKey(fmt.Sprintf(
		"%v-%v",
		depositAddress,
		strings.ReplaceAll(monitoringName, " ", ""),
	))
}

type monitoringLock struct {
	locks map[monitoringLockKey]bool
	mutex sync.Mutex
}

func newMonitoringLock() *monitoringLock {
	return &monitoringLock{
		locks: make(map[monitoringLockKey]bool),
	}
}

func (ml *monitoringLock) tryLock(depositAddress, monitoringName string) bool {
	ml.mutex.Lock()
	defer ml.mutex.Unlock()

	key := newMonitoringLockKey(depositAddress, monitoringName)

	if ml.locks[key] == true {
		return false
	}

	ml.locks[key] = true

	return true
}

func (ml *monitoringLock) release(depositAddress, monitoringName string) {
	ml.mutex.Lock()
	defer ml.mutex.Unlock()

	key := newMonitoringLockKey(depositAddress, monitoringName)

	delete(ml.locks, key)
}
