package client

import (
	"encoding/hex"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

var once sync.Once

// requestedSignersTrack is used to track signers generation started after keep
// creation event is received. It is used to ensure that the process execution
// is not duplicated, e.g. when the client receives the same event multiple times.
// When event is received, it should be noted in this struct. When the signer
// generation process completes (no matter if it succeeded or failed), it should
// be removed from this struct.
type requestedSignersTrack struct {
	data  map[string]bool // <keep, bool>
	mutex *sync.Mutex
}

func (rst *requestedSignersTrack) add(keepAddress common.Address) bool {
	rst.mutex.Lock()
	defer rst.mutex.Unlock()

	if rst.data[keepAddress.String()] == true {
		return false
	}

	rst.data[keepAddress.String()] = true

	return true
}

func (rst *requestedSignersTrack) remove(keepAddress common.Address) {
	rst.mutex.Lock()
	defer rst.mutex.Unlock()

	delete(rst.data, keepAddress.String())
}

// requestedSignaturesTrack is used to track signature calculation started after
// signature request event is received. It is used to ensure that the process execution
// is not duplicated, e.g. when the client receives the same event multiple times.
// When event is received, it should be noted in this struct. When the signature
// calculation process completes (no matter if it succeeded or failed), it should
// be removed from this struct.
type requestedSignaturesTrack struct {
	data  map[string]map[string]bool // <keep, <digest, bool>>
	mutex *sync.Mutex
}

func (rst *requestedSignaturesTrack) add(keepAddress common.Address, digest [32]byte) bool {
	rst.mutex.Lock()
	defer rst.mutex.Unlock()

	digestString := hex.EncodeToString(digest[:])

	keepSignaturesRequests, ok := rst.data[keepAddress.String()]
	if !ok {
		rst.data[keepAddress.String()] = map[string]bool{digestString: true}
		return true
	}
	if keepSignaturesRequests[digestString] == true {
		return false
	}

	keepSignaturesRequests[digestString] = true
	return true

}

func (rst *requestedSignaturesTrack) remove(keepAddress common.Address, digest [32]byte) {
	rst.mutex.Lock()
	defer rst.mutex.Unlock()

	if keepSignatures, ok := rst.data[keepAddress.String()]; ok {
		digestString := hex.EncodeToString(digest[:])
		delete(keepSignatures, digestString)

		if len(keepSignatures) == 0 {
			delete(rst.data, keepAddress.String())
		}
	}
}

type keepClosedTrack struct {
	data  map[string]bool // <keep, bool>
	mutex *sync.Mutex
}

var keepClosedTrackInstance *keepClosedTrack

func getKeepClosedTrackInstance() *keepClosedTrack {
	if keepClosedTrackInstance == nil {
		once.Do(
			func() {
				keepClosedTrackInstance = &keepClosedTrack{
					data:  make(map[string]bool),
					mutex: &sync.Mutex{},
				}
			})
	}

	return keepClosedTrackInstance
}

func (kct *keepClosedTrack) add(keepAddress common.Address) bool {
	kct.mutex.Lock()
	defer kct.mutex.Unlock()

	if kct.data[keepAddress.String()] {
		return false
	}

	kct.data[keepAddress.String()] = true

	return true
}

func (kct *keepClosedTrack) remove(keepAddress common.Address) {
	kct.mutex.Lock()
	defer kct.mutex.Unlock()

	delete(kct.data, keepAddress.String())
}
