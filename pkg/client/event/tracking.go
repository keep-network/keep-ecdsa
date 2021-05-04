package event

import (
	"encoding/hex"
	"sync"

	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

// uniqueEventTrack is a simple event track implementation allowing to track
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
type uniqueEventTrack struct {
	data  map[string]bool // <keep, bool>
	mutex sync.Mutex
}

func (uet *uniqueEventTrack) add(keepID chain.ID) bool {
	uet.mutex.Lock()
	defer uet.mutex.Unlock()

	if uet.data[keepID.String()] == true {
		return false
	}

	uet.data[keepID.String()] = true

	return true
}

func (uet *uniqueEventTrack) has(keepID chain.ID) bool {
	uet.mutex.Lock()
	defer uet.mutex.Unlock()

	return uet.data[keepID.String()]
}

func (uet *uniqueEventTrack) remove(keepID chain.ID) {
	uet.mutex.Lock()
	defer uet.mutex.Unlock()

	delete(uet.data, keepID.String())
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

func (rst *requestedSignaturesTrack) add(keepID chain.ID, digest [32]byte) bool {
	rst.mutex.Lock()
	defer rst.mutex.Unlock()

	digestString := hex.EncodeToString(digest[:])

	keepSignaturesRequests, ok := rst.data[keepID.String()]
	if !ok {
		rst.data[keepID.String()] = map[string]bool{digestString: true}
		return true
	}
	if keepSignaturesRequests[digestString] == true {
		return false
	}

	keepSignaturesRequests[digestString] = true
	return true

}

func (rst *requestedSignaturesTrack) has(keepID chain.ID, digest [32]byte) bool {
	rst.mutex.Lock()
	defer rst.mutex.Unlock()

	keepSignaturesRequests, ok := rst.data[keepID.String()]
	if !ok {
		return false
	}

	digestString := hex.EncodeToString(digest[:])
	return keepSignaturesRequests[digestString]
}

func (rst *requestedSignaturesTrack) remove(keepID chain.ID, digest [32]byte) {
	rst.mutex.Lock()
	defer rst.mutex.Unlock()

	if keepSignatures, ok := rst.data[keepID.String()]; ok {
		digestString := hex.EncodeToString(digest[:])
		delete(keepSignatures, digestString)

		if len(keepSignatures) == 0 {
			delete(rst.data, keepID.String())
		}
	}
}
