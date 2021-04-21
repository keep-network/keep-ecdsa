package event

import (
	"context"
	"testing"

	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/local"
)

var (
	keepID1String = "0x4e09cadc7037afa36603138d1c0b76fe2aa5039c"
	keepID2String = "0x0000000000000000000000000000000000000002"

	keepID1, keepID2 chain.ID
)

func init() {
	localChain := local.Connect(context.Background())

	keepID1, _ = localChain.UnmarshalID(keepID1String)
	keepID2, _ = localChain.UnmarshalID(keepID2String)
}

func TestUniqueEventTrackAdd(t *testing.T) {
	rs := &uniqueEventTrack{
		data: make(map[string]bool),
	}

	if !rs.add(keepID1) {
		t.Error("event wasn't emitted before; should be added successfully")
	}

	if !rs.add(keepID2) {
		t.Error("event wasn't emitted before; should be added successfully")
	}
}

func TestUniqueEventTrackAdd_Duplicate(t *testing.T) {
	rs := &uniqueEventTrack{
		data: make(map[string]bool),
	}

	if !rs.add(keepID1) {
		t.Error("event wasn't emitted before; should be added successfully")
	}

	if rs.add(keepID1) {
		t.Error("event was emitted before; it should not be added")
	}
}

func TestUniqueEventTrackRemove(t *testing.T) {
	rs := &uniqueEventTrack{
		data: make(map[string]bool),
	}

	if !rs.add(keepID1) {
		t.Error("event wasn't emitted before; should be added successfully")
	}

	rs.remove(keepID1)

	if !rs.add(keepID1) {
		t.Error("event was removed from tracking; should be added successfully")
	}
}

func TestUniqueEventTrackRemove_WhenEmpty(t *testing.T) {
	rs := &uniqueEventTrack{
		data: make(map[string]bool),
	}

	rs.remove(keepID1)

	if !rs.add(keepID1) {
		t.Error("event wasn't emitted before; should be added successfully")
	}
}

func TestUniqueEventTrackHas(t *testing.T) {

	rs := &uniqueEventTrack{
		data: make(map[string]bool),
	}

	rs.add(keepID1)

	if !rs.has(keepID1) {
		t.Error("event was emitted and should be tracked")
	}
	if rs.has(keepID2) {
		t.Error("event was not emitted and should not be tracked")
	}

	rs.remove(keepID1)
	if rs.has(keepID1) {
		t.Error("event was removed and should no longer be tracked")
	}
}

func TestRequestedSignaturesTrackAdd_SameKeep(t *testing.T) {
	digest1 := [32]byte{9}
	digest2 := [32]byte{8}

	rs := &requestedSignaturesTrack{
		data: make(map[string]map[string]bool),
	}

	if !rs.add(keepID1, digest1) {
		t.Error(
			"signature for the first digest wasn't requested before; " +
				"event should be added successfully",
		)
	}
	if !rs.add(keepID1, digest2) {
		t.Error(
			"signature for the second digest wasn't requested before; " +
				"event should be added successfully",
		)
	}
}

func TestRequestedSignaturesTrackAdd_DifferentKeeps(t *testing.T) {

	digest1 := [32]byte{9}
	digest2 := [32]byte{8}

	rs := &requestedSignaturesTrack{
		data: make(map[string]map[string]bool),
	}

	if !rs.add(keepID1, digest1) {
		t.Error(
			"signature from the first keep wasn't requested before; " +
				"event should be added successfully",
		)
	}

	if !rs.add(keepID2, digest2) {
		t.Error(
			"signature from the second keep wasn't requested before; " +
				"event should be added successfully",
		)
	}
}

func TestRequestedSignaturesTrackAdd_Duplicate(t *testing.T) {

	rs := &requestedSignaturesTrack{
		data: make(map[string]map[string]bool),
	}

	if !rs.add(keepID1, digest) {
		t.Error(
			"signature wasn't requested before; event should be added",
		)
	}

	if rs.add(keepID1, digest) {
		t.Error("signature was requested before; event should not be added")
	}
}

func TestRequestedSignaturesTrackRemove(t *testing.T) {

	digest := [32]byte{9}

	rs := &requestedSignaturesTrack{
		data: make(map[string]map[string]bool),
	}

	if !rs.add(keepID1, digest) {
		t.Error(
			"signature from the first keep wasn't requested before; " +
				"event should be added successfully",
		)
	}

	if !rs.add(keepID2, digest) {
		t.Error(
			"signature from the second keep wasn't requested before; " +
				"event should be added successfully",
		)
	}

	rs.remove(keepID1, digest)

	if !rs.add(keepID1, digest) {
		t.Error(
			"signature event for the first keep was removed from tracking; " +
				"event should be added successfully",
		)
	}

	if rs.add(keepID2, digest) {
		t.Error(
			"signature event for the second keep was not removed from tracking; " +
				"event should not be added",
		)
	}
}

func TestRequestedSignaturesTrackRemove_WhenEmpty(t *testing.T) {

	rs := &requestedSignaturesTrack{
		data: make(map[string]map[string]bool),
	}

	rs.remove(keepID1, digest)

	if !rs.add(keepID1, digest) {
		t.Error(
			"signature from the first keep wasn't requested before; " +
				"event should be added successfully",
		)
	}
}

func TestRequestedSignaturesTrackHas(t *testing.T) {

	digest1 := [32]byte{9}
	digest2 := [32]byte{10}

	rs := &requestedSignaturesTrack{
		data: make(map[string]map[string]bool),
	}

	rs.add(keepID1, digest1)

	if !rs.has(keepID1, digest1) {
		t.Errorf("event was emitted and should be tracked")
	}
	if rs.has(keepID1, digest2) {
		t.Errorf("event with this digest was not emitted and should not be tracked")
	}
	if rs.has(keepID2, digest1) {
		t.Errorf("event for this keep was not emitted and should not be tracked")
	}

	rs.remove(keepID1, digest1)
	if rs.has(keepID1, digest1) {
		t.Errorf("event was removed and should no longer be tracked")
	}
}
