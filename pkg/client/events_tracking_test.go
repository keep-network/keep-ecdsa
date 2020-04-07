package client

import (
	"encoding/hex"
	"fmt"
	"reflect"
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

	if err := rs.add(keepAddress1); err != nil {
		t.Errorf("unexpected error: [%v]", err)
	}

	if !rs.data[keepAddress1.String()] {
		t.Errorf(
			"unexpected value for keep [%s]\nexpected: [%v]\nactual:   [%v]",
			keepAddress1.String(),
			true,
			rs.data[keepAddress1.String()],
		)
	}

	if err := rs.add(keepAddress2); err != nil {
		t.Errorf("unexpected error: [%v]", err)
	}

	if !rs.data[keepAddress2.String()] {
		t.Errorf(
			"unexpected value for keep [%s]\nexpected: [%v]\nactual:   [%v]",
			keepAddress2.String(),
			true,
			rs.data[keepAddress1.String()],
		)
	}
}

func TestRequestedSignersTrackAdd_Duplicate(t *testing.T) {
	keepAddress1 := common.BytesToAddress([]byte{1})
	expectedError := fmt.Errorf("signer generation already requested for keep: [0x0000000000000000000000000000000000000001]")

	rs := &requestedSignersTrack{
		data:  make(map[string]bool),
		mutex: &sync.Mutex{},
	}

	if err := rs.add(keepAddress1); err != nil {
		t.Errorf("unexpected error: [%v]", err)
	}

	err := rs.add(keepAddress1)
	if !reflect.DeepEqual(expectedError, err) {
		t.Errorf(
			"unexpected error\nexpected: [%v]\nactual:   [%v]",
			expectedError,
			err,
		)
	}

	if !rs.data[keepAddress1.String()] {
		t.Errorf(
			"unexpected value for keep [%s]\nexpected: [%v]\nactual:   [%v]",
			keepAddress1.String(),
			true,
			rs.data[keepAddress1.String()],
		)
	}
}

func TestRequestedSignersTrackRemove(t *testing.T) {
	keepAddress1 := common.BytesToAddress([]byte{1})

	rs := &requestedSignersTrack{
		data:  make(map[string]bool),
		mutex: &sync.Mutex{},
	}

	if err := rs.add(keepAddress1); err != nil {
		t.Errorf("unexpected error: [%v]", err)
	}

	rs.remove(keepAddress1)

	if rs.data[keepAddress1.String()] {
		t.Errorf(
			"unexpected value for keep [%s]\nexpected: [%v]\nactual:   [%v]",
			keepAddress1.String(),
			false,
			rs.data[keepAddress1.String()],
		)
	}
}

func TestRequestedSignersTrackRemove_WhenEmpty(t *testing.T) {
	keepAddress1 := common.BytesToAddress([]byte{1})

	rs := &requestedSignersTrack{
		data:  make(map[string]bool),
		mutex: &sync.Mutex{},
	}

	rs.remove(keepAddress1)

	if rs.data[keepAddress1.String()] {
		t.Errorf(
			"unexpected value for keep [%s]\nexpected: [%v]\nactual:   [%v]",
			keepAddress1.String(),
			false,
			rs.data[keepAddress1.String()],
		)
	}
}

func TestRequestedSignaturesTrackAdd_SameKeep(t *testing.T) {
	keepAddress1 := common.BytesToAddress([]byte{1})

	digest1 := [32]byte{9}
	digest2 := [32]byte{8}

	digest1String := hex.EncodeToString(digest1[:])
	digest2String := hex.EncodeToString(digest2[:])

	rs := &requestedSignaturesTrack{
		data:  make(map[string]map[string]bool),
		mutex: &sync.Mutex{},
	}

	expectedMap := map[string]map[string]bool{
		keepAddress1.String(): map[string]bool{
			digest1String: true,
		},
	}
	if err := rs.add(keepAddress1, digest1); err != nil {
		t.Errorf("unexpected error: [%v]", err)
	}
	if !reflect.DeepEqual(expectedMap, rs.data) {
		t.Errorf(
			"unexpected map content\nexpected: [%v]\nactual:   [%v]",
			expectedMap,
			rs.data,
		)
	}

	expectedMap = map[string]map[string]bool{
		keepAddress1.String(): map[string]bool{
			digest1String: true,
			digest2String: true,
		},
	}
	if err := rs.add(keepAddress1, digest2); err != nil {
		t.Errorf("unexpected error: [%v]", err)
	}
	if !reflect.DeepEqual(expectedMap, rs.data) {
		t.Errorf(
			"unexpected map content\nexpected: [%v]\nactual:   [%v]",
			expectedMap,
			rs.data,
		)
	}
}

func TestRequestedSignaturesTrackAdd_DifferentKeeps(t *testing.T) {
	keepAddress1 := common.BytesToAddress([]byte{1})
	keepAddress2 := common.BytesToAddress([]byte{2})

	digest1 := [32]byte{9}
	digest2 := [32]byte{8}

	digest1String := hex.EncodeToString(digest1[:])
	digest2String := hex.EncodeToString(digest2[:])

	rs := &requestedSignaturesTrack{
		data:  make(map[string]map[string]bool),
		mutex: &sync.Mutex{},
	}

	expectedMap := map[string]map[string]bool{
		keepAddress1.String(): map[string]bool{
			digest1String: true,
		},
	}
	if err := rs.add(keepAddress1, digest1); err != nil {
		t.Errorf("unexpected error: [%v]", err)
	}
	if !reflect.DeepEqual(expectedMap, rs.data) {
		t.Errorf(
			"unexpected map content\nexpected: [%v]\nactual:   [%v]",
			expectedMap,
			rs.data,
		)
	}

	expectedMap = map[string]map[string]bool{
		keepAddress1.String(): map[string]bool{
			digest1String: true,
		},
		keepAddress2.String(): map[string]bool{
			digest2String: true,
		},
	}
	if err := rs.add(keepAddress2, digest2); err != nil {
		t.Errorf("unexpected error: [%v]", err)
	}
	if !reflect.DeepEqual(expectedMap, rs.data) {
		t.Errorf(
			"unexpected map content\nexpected: [%v]\nactual:   [%v]",
			expectedMap,
			rs.data,
		)
	}
}

func TestRequestedSignaturesTrackAdd_Duplicate(t *testing.T) {
	keepAddress1 := common.BytesToAddress([]byte{1})
	digest1 := [32]byte{9}
	digest1String := hex.EncodeToString(digest1[:])

	expectedError := fmt.Errorf("signature for digest [0900000000000000000000000000000000000000000000000000000000000000] already requested from keep: [0x0000000000000000000000000000000000000001]")

	rs := &requestedSignaturesTrack{
		data:  make(map[string]map[string]bool),
		mutex: &sync.Mutex{},
	}

	expectedMap := map[string]map[string]bool{
		keepAddress1.String(): map[string]bool{
			digest1String: true,
		},
	}
	if err := rs.add(keepAddress1, digest1); err != nil {
		t.Errorf("unexpected error: [%v]", err)
	}
	if !reflect.DeepEqual(expectedMap, rs.data) {
		t.Errorf(
			"unexpected map content\nexpected: [%v]\nactual:   [%v]",
			expectedMap,
			rs.data,
		)
	}

	err := rs.add(keepAddress1, digest1)
	if !reflect.DeepEqual(expectedError, err) {
		t.Errorf(
			"unexpected error\nexpected: [%v]\nactual:   [%v]",
			expectedError,
			err,
		)
	}
	if !reflect.DeepEqual(expectedMap, rs.data) {
		t.Errorf(
			"unexpected map content\nexpected: [%v]\nactual:   [%v]",
			expectedMap,
			rs.data,
		)
	}
}

func TestRequestedSignaturesTrackRemove(t *testing.T) {
	keepAddress1 := common.BytesToAddress([]byte{1})
	keepAddress2 := common.BytesToAddress([]byte{2})

	digest1 := [32]byte{9}
	digest2 := [32]byte{8}
	digest3 := [32]byte{7}

	digest1String := hex.EncodeToString(digest1[:])
	digest2String := hex.EncodeToString(digest2[:])
	digest3String := hex.EncodeToString(digest3[:])

	rs := &requestedSignaturesTrack{
		data:  make(map[string]map[string]bool),
		mutex: &sync.Mutex{},
	}

	rs.data = map[string]map[string]bool{
		keepAddress1.String(): map[string]bool{
			digest1String: true,
			digest2String: true,
		},
		keepAddress2.String(): map[string]bool{
			digest3String: true,
		},
	}

	// Remove keep1 : digest1
	expectedMap := map[string]map[string]bool{
		keepAddress1.String(): map[string]bool{
			digest2String: true,
		},
		keepAddress2.String(): map[string]bool{
			digest3String: true,
		},
	}

	rs.remove(keepAddress1, digest1)

	if !reflect.DeepEqual(expectedMap, rs.data) {
		t.Errorf(
			"unexpected map content\nexpected: [%v]\nactual:   [%v]",
			expectedMap,
			rs.data,
		)
	}

	// Remove keep2 : digest3
	expectedMap = map[string]map[string]bool{
		keepAddress1.String(): map[string]bool{
			digest2String: true,
		},
	}

	rs.remove(keepAddress2, digest3)

	if !reflect.DeepEqual(expectedMap, rs.data) {
		t.Errorf(
			"unexpected map content\nexpected: [%v]\nactual:   [%v]",
			expectedMap,
			rs.data,
		)
	}

	// Remove keep1 : digest1
	expectedMap = map[string]map[string]bool{}

	rs.remove(keepAddress1, digest2)

	if !reflect.DeepEqual(expectedMap, rs.data) {
		t.Errorf(
			"unexpected map content\nexpected: [%v]\nactual:   [%v]",
			expectedMap,
			rs.data,
		)
	}
}

func TestRequestedSignaturesTrackRemove_WhenEmpty(t *testing.T) {
	keepAddress1 := common.BytesToAddress([]byte{1})
	digest1 := [32]byte{9}

	rs := &requestedSignaturesTrack{
		data:  make(map[string]map[string]bool),
		mutex: &sync.Mutex{},
	}

	expectedMap := map[string]map[string]bool{}

	rs.remove(keepAddress1, digest1)

	if !reflect.DeepEqual(expectedMap, rs.data) {
		t.Errorf(
			"unexpected map content\nexpected: [%v]\nactual:   [%v]",
			expectedMap,
			rs.data,
		)
	}
}
