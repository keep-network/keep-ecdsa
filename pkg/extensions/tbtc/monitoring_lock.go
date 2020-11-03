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
	locks sync.Map
}

func newMonitoringLock() *monitoringLock {
	return &monitoringLock{}
}

func (ml *monitoringLock) tryLock(depositAddress, monitoringName string) bool {
	_, isExistingKey := ml.locks.LoadOrStore(
		newMonitoringLockKey(depositAddress, monitoringName),
		true,
	)

	return !isExistingKey
}

func (ml *monitoringLock) release(depositAddress, monitoringName string) {
	ml.locks.Delete(newMonitoringLockKey(depositAddress, monitoringName))
}