package tbtc

import "testing"

func TestMonitoringLock_TryLock(t *testing.T) {
	monitoringLock := newMonitoringLock()

	if !monitoringLock.tryLock("0xAA", "monitoring one") {
		t.Errorf("monitoring wasn't started before; should be locked successfully")
	}

	if !monitoringLock.tryLock("0xBB", "monitoring one") {
		t.Errorf("monitoring wasn't started before; should be locked successfully")
	}

	if !monitoringLock.tryLock("0xAA", "monitoring two") {
		t.Errorf("monitoring wasn't started before; should be locked successfully")
	}

	if !monitoringLock.tryLock("0xBB", "monitoring two") {
		t.Errorf("monitoring wasn't started before; should be locked successfully")
	}
}

func TestMonitoringLock_TryLock_Duplicate(t *testing.T) {
	monitoringLock := newMonitoringLock()

	if !monitoringLock.tryLock("0xAA", "monitoring one") {
		t.Errorf("monitoring wasn't started before; should be locked successfully")
	}

	if monitoringLock.tryLock("0xAA", "monitoring one") {
		t.Errorf("monitoring was started before; lock attempt should be rejected")
	}
}

func TestMonitoringLock_Release(t *testing.T) {
	monitoringLock := newMonitoringLock()

	if !monitoringLock.tryLock("0xAA", "monitoring one") {
		t.Errorf("monitoring wasn't started before; should be locked successfully")
	}

	monitoringLock.release("0xAA", "monitoring one")

	if !monitoringLock.tryLock("0xAA", "monitoring one") {
		t.Errorf("monitoring lock has been released; should be locked successfully")
	}
}

func TestMonitoringLock_Release_WhenEmpty(t *testing.T) {
	monitoringLock := newMonitoringLock()

	monitoringLock.release("0xAA", "monitoring one")

	if !monitoringLock.tryLock("0xAA", "monitoring one") {
		t.Errorf("monitoring wasn't started before; should be locked successfully")
	}
}
