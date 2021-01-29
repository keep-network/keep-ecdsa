package event

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestUniqueEventTrackAdd(t *testing.T) {
	keepAddress1 := common.BytesToAddress([]byte{1})
	keepAddress2 := common.BytesToAddress([]byte{2})

	rs := &uniqueEventTrack{
		data: make(map[string]bool),
	}

	if !rs.add(keepAddress1) {
		t.Error("event wasn't emitted before; should be added successfully")
	}

	if !rs.add(keepAddress2) {
		t.Error("event wasn't emitted before; should be added successfully")
	}
}

func TestUniqueEventTrackAdd_Duplicate(t *testing.T) {
	keepAddress := common.BytesToAddress([]byte{1})

	rs := &uniqueEventTrack{
		data: make(map[string]bool),
	}

	if !rs.add(keepAddress) {
		t.Error("event wasn't emitted before; should be added successfully")
	}

	if rs.add(keepAddress) {
		t.Error("event was emitted before; it should not be added")
	}
}

func TestUniqueEventTrackRemove(t *testing.T) {
	keepAddress := common.BytesToAddress([]byte{1})

	rs := &uniqueEventTrack{
		data: make(map[string]bool),
	}

	if !rs.add(keepAddress) {
		t.Error("event wasn't emitted before; should be added successfully")
	}

	rs.remove(keepAddress)

	if !rs.add(keepAddress) {
		t.Error("event was removed from tracking; should be added successfully")
	}
}

func TestUniqueEventTrackRemove_WhenEmpty(t *testing.T) {
	keepAddress := common.BytesToAddress([]byte{1})

	rs := &uniqueEventTrack{
		data: make(map[string]bool),
	}

	rs.remove(keepAddress)

	if !rs.add(keepAddress) {
		t.Error("event wasn't emitted before; should be added successfully")
	}
}

func TestUniqueEventTrackHas(t *testing.T) {
	keepAddress1 := common.BytesToAddress([]byte{1})
	keepAddress2 := common.BytesToAddress([]byte{2})

	rs := &uniqueEventTrack{
		data: make(map[string]bool),
	}

	rs.add(keepAddress1)

	if !rs.has(keepAddress1) {
		t.Error("event was emitted and should be tracked")
	}
	if rs.has(keepAddress2) {
		t.Error("event was not emitted and should not be tracked")
	}

	rs.remove(keepAddress1)
	if rs.has(keepAddress1) {
		t.Error("event was removed and should no longer be tracked")
	}
}

func TestRequestedSignaturesTrackAdd_SameKeep(t *testing.T) {
	keepAddress := common.BytesToAddress([]byte{1})

	digest1 := [32]byte{9}
	digest2 := [32]byte{8}

	rs := &requestedSignaturesTrack{
		data: make(map[string]map[string]bool),
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
		data: make(map[string]map[string]bool),
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
		data: make(map[string]map[string]bool),
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
		data: make(map[string]map[string]bool),
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
		data: make(map[string]map[string]bool),
	}

	rs.remove(keepAddress, digest)

	if !rs.add(keepAddress, digest) {
		t.Error(
			"signature from the first keep wasn't requested before; " +
				"event should be added successfully",
		)
	}
}

func TestRequestedSignaturesTrackHas(t *testing.T) {
	keepAddress1 := common.BytesToAddress([]byte{1})
	keepAddress2 := common.BytesToAddress([]byte{2})

	digest1 := [32]byte{9}
	digest2 := [32]byte{10}

	rs := &requestedSignaturesTrack{
		data: make(map[string]map[string]bool),
	}

	rs.add(keepAddress1, digest1)

	if !rs.has(keepAddress1, digest1) {
		t.Errorf("event was emitted and should be tracked")
	}
	if rs.has(keepAddress1, digest2) {
		t.Errorf("event with this digest was not emitted and should not be tracked")
	}
	if rs.has(keepAddress2, digest1) {
		t.Errorf("event for this keep was not emitted and should not be tracked")
	}

	rs.remove(keepAddress1, digest1)
	if rs.has(keepAddress1, digest1) {
		t.Errorf("event was removed and should no longer be tracked")
	}
}
