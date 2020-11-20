package client

import (
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestRequestedSignersTrackAdd(t *testing.T) {
	keepAddress1 := common.BytesToAddress([]byte{1})
	keepAddress2 := common.BytesToAddress([]byte{2})

	rs := &requestedSignersTrack{
		data:  make(map[string]bool),
		mutex: &sync.Mutex{},
	}

	if !rs.add(keepAddress1) {
		t.Error("event wasn't emitted before; should be added successfully")
	}

	if !rs.add(keepAddress2) {
		t.Error("event wasn't emitted before; should be added successfully")
	}
}

func TestRequestedSignersTrackAdd_Duplicate(t *testing.T) {
	keepAddress := common.BytesToAddress([]byte{1})

	rs := &requestedSignersTrack{
		data:  make(map[string]bool),
		mutex: &sync.Mutex{},
	}

	if !rs.add(keepAddress) {
		t.Error("event wasn't emitted before; should be added successfully")
	}

	if rs.add(keepAddress) {
		t.Error("event was emitted before; it should not be added")
	}
}

func TestRequestedSignersTrackRemove(t *testing.T) {
	keepAddress := common.BytesToAddress([]byte{1})

	rs := &requestedSignersTrack{
		data:  make(map[string]bool),
		mutex: &sync.Mutex{},
	}

	if !rs.add(keepAddress) {
		t.Error("event wasn't emitted before; should be added successfully")
	}

	rs.remove(keepAddress)

	if !rs.add(keepAddress) {
		t.Error("event was removed from tracking; should be added successfully")
	}
}

func TestRequestedSignersTrackRemove_WhenEmpty(t *testing.T) {
	keepAddress := common.BytesToAddress([]byte{1})

	rs := &requestedSignersTrack{
		data:  make(map[string]bool),
		mutex: &sync.Mutex{},
	}

	rs.remove(keepAddress)

	if !rs.add(keepAddress) {
		t.Error("event wasn't emitted before; should be added successfully")
	}
}

func TestRequestedSignaturesTrackAdd_SameKeep(t *testing.T) {
	keepAddress := common.BytesToAddress([]byte{1})

	digest1 := [32]byte{9}
	digest2 := [32]byte{8}

	rs := &requestedSignaturesTrack{
		data:  make(map[string]map[string]bool),
		mutex: &sync.Mutex{},
	}

	if !rs.add(keepAddress, digest1) {
		t.Error(
			"signature for the first digest wasn't requested before; " +
				"event should be added successfully",
		)
	}
	if !rs.add(keepAddress, digest2) {
		t.Error(
			"signature for the second digest wasn't requested before; " +
				"event should be added successfully",
		)
	}
}

func TestRequestedSignaturesTrackAdd_DifferentKeeps(t *testing.T) {
	keepAddress1 := common.BytesToAddress([]byte{1})
	keepAddress2 := common.BytesToAddress([]byte{2})

	digest1 := [32]byte{9}
	digest2 := [32]byte{8}

	rs := &requestedSignaturesTrack{
		data:  make(map[string]map[string]bool),
		mutex: &sync.Mutex{},
	}

	if !rs.add(keepAddress1, digest1) {
		t.Error(
			"signature from the first keep wasn't requested before; " +
				"event should be added successfully",
		)
	}

	if !rs.add(keepAddress2, digest2) {
		t.Error(
			"signature from the second keep wasn't requested before; " +
				"event should be added successfully",
		)
	}
}

func TestRequestedSignaturesTrackAdd_Duplicate(t *testing.T) {
	keepAddress := common.BytesToAddress([]byte{1})
	digest := [32]byte{9}

	rs := &requestedSignaturesTrack{
		data:  make(map[string]map[string]bool),
		mutex: &sync.Mutex{},
	}

	if !rs.add(keepAddress, digest) {
		t.Error(
			"signature wasn't requested before; event should be added",
		)
	}

	if rs.add(keepAddress, digest) {
		t.Error("signature was requested before; event should not be added")
	}
}

func TestRequestedSignaturesTrackRemove(t *testing.T) {
	keepAddress1 := common.BytesToAddress([]byte{1})
	keepAddress2 := common.BytesToAddress([]byte{2})

	digest := [32]byte{9}

	rs := &requestedSignaturesTrack{
		data:  make(map[string]map[string]bool),
		mutex: &sync.Mutex{},
	}

	if !rs.add(keepAddress1, digest) {
		t.Error(
			"signature from the first keep wasn't requested before; " +
				"event should be added successfully",
		)
	}

	if !rs.add(keepAddress2, digest) {
		t.Error(
			"signature from the second keep wasn't requested before; " +
				"event should be added successfully",
		)
	}

	rs.remove(keepAddress1, digest)

	if !rs.add(keepAddress1, digest) {
		t.Error(
			"signature event for the first keep was removed from tracking; " +
				"event should be added successfully",
		)
	}

	if rs.add(keepAddress2, digest) {
		t.Error(
			"signature event for the second keep was not removed from tracking; " +
				"event should not be added",
		)
	}
}

func TestRequestedSignaturesTrackRemove_WhenEmpty(t *testing.T) {
	keepAddress := common.BytesToAddress([]byte{1})
	digest := [32]byte{9}

	rs := &requestedSignaturesTrack{
		data:  make(map[string]map[string]bool),
		mutex: &sync.Mutex{},
	}

	rs.remove(keepAddress, digest)

	if !rs.add(keepAddress, digest) {
		t.Error(
			"signature from the first keep wasn't requested before; " +
				"event should be added successfully",
		)
	}
}

func TestKeepClosedTrackAdd(t *testing.T) {
	keepAddress1 := common.BytesToAddress([]byte{1})
	keepAddress2 := common.BytesToAddress([]byte{2})

	kct := getKeepClosedTrackInstance()

	if !kct.add(keepAddress1) {
		t.Error("event wasn't emitted before; should be added successfully")
	}

	if !kct.add(keepAddress2) {
		t.Error("event wasn't emitted before; should be added successfully")
	}

	keepClosedTrackCleanup([]common.Address{keepAddress1, keepAddress2})
}

func TestKeepClosedTrackAdd_Duplicate(t *testing.T) {
	keepAddress := common.BytesToAddress([]byte{1})

	kct := getKeepClosedTrackInstance()

	if !kct.add(keepAddress) {
		t.Error("event wasn't emitted before; should be added successfully")
	}

	if kct.add(keepAddress) {
		t.Error("event was emitted before; it should not be added")
	}

	keepClosedTrackCleanup([]common.Address{keepAddress})
}

func TestKeepClosedTrackRemove(t *testing.T) {
	keepAddress := common.BytesToAddress([]byte{1})

	kct := getKeepClosedTrackInstance()

	if !kct.add(keepAddress) {
		t.Error("event wasn't emitted before; should be added successfully")
	}

	kct.remove(keepAddress)

	if !kct.add(keepAddress) {
		t.Error("event was removed from tracking; should be added successfully")
	}

	keepClosedTrackCleanup([]common.Address{keepAddress})
}

func TestKeepClosedTrackRemove_WhenEmpty(t *testing.T) {
	keepAddress := common.BytesToAddress([]byte{1})

	kct := getKeepClosedTrackInstance()

	kct.remove(keepAddress)

	if !kct.add(keepAddress) {
		t.Error("event wasn't emitted before; should be added successfully")
	}

	keepClosedTrackCleanup([]common.Address{keepAddress})
}

func TestKeepClosedTrack_GetOneInstance(t *testing.T) {
	kct1 := getKeepClosedTrackInstance()
	kct2 := getKeepClosedTrackInstance()

	if kct1 != kct2 {
		t.Error("should be only one instance for keep closed tracking")
	}
}

func keepClosedTrackCleanup(addresses []common.Address) {
	kct := getKeepClosedTrackInstance()
	for _, addr := range addresses {
		kct.remove(addr)
	}
}