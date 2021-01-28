package event

import (
	"encoding/hex"
	"sync"
)

// keepEventTrack is a simple event track implementation allowing to track
// events that are happening once in the entire keep history, for example:
// - keep opened, generating signer
// - keep getting closed
// - keep getting archived
//
// Those events do not require any special tracking or additional fields to
// distinguish them. It is used to ensure that the process execution is not
// duplicated, e.g. when the client receives the same event multiple times. When
// event is received, it should be noted in this struct. When the signer
// generation process completes (no matter if it succeeded or failed), it should
// be removed from this struct.
type keepEventTrack struct {
	data  map[string]bool // <keep, bool>
	mutex sync.Mutex
}

func (ket *keepEventTrack) add(keepAddress string) bool {
	ket.mutex.Lock()
	defer ket.mutex.Unlock()

	if ket.data[keepAddress] == true {
		return false
	}

	ket.data[keepAddress] = true

	return true
}

func (ket *keepEventTrack) has(keepAddress string) bool {
	ket.mutex.Lock()
	defer ket.mutex.Unlock()

	return ket.data[keepAddress]
}

func (ket *keepEventTrack) remove(keepAddress string) {
	ket.mutex.Lock()
	defer ket.mutex.Unlock()

	delete(ket.data, keepAddress)
}

// requestedSignaturesTrack is used to track signature calculation started after
// signature request event is received. It is used to ensure that the process execution
// is not duplicated, e.g. when the client receives the same event multiple times.
// When event is received, it should be noted in this struct. When the signature
// calculation process completes (no matter if it succeeded or failed), it should
// be removed from this struct.
type requestedSignaturesTrack struct {
	data  map[string]map[string]bool // <keep, <digest, bool>>
	mutex sync.Mutex
}

func (rst *requestedSignaturesTrack) add(keepAddress string, digest [32]byte) bool {
	rst.mutex.Lock()
	defer rst.mutex.Unlock()

	digestString := hex.EncodeToString(digest[:])

	keepSignaturesRequests, ok := rst.data[keepAddress]
	if !ok {
		rst.data[keepAddress] = map[string]bool{digestString: true}
		return true
	}
	if keepSignaturesRequests[digestString] == true {
		return false
	}

	keepSignaturesRequests[digestString] = true
	return true

}

func (rst *requestedSignaturesTrack) has(keepAddress string, digest [32]byte) bool {
	rst.mutex.Lock()
	defer rst.mutex.Unlock()

	keepSignaturesRequests, ok := rst.data[keepAddress]
	if !ok {
		return false
	}

	digestString := hex.EncodeToString(digest[:])
	return keepSignaturesRequests[digestString]
}

func (rst *requestedSignaturesTrack) remove(keepAddress string, digest [32]byte) {
	rst.mutex.Lock()
	defer rst.mutex.Unlock()

	if keepSignatures, ok := rst.data[keepAddress]; ok {
		digestString := hex.EncodeToString(digest[:])
		delete(keepSignatures, digestString)

		if len(keepSignatures) == 0 {
			delete(rst.data, keepAddress)
		}
	}
}
