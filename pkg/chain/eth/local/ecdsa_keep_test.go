package local

import (
	"testing"

	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
)

func TestRequestSignatureNonexistentKeep(t *testing.T) {
	chain := initializeLocalChain()
	keepAddress := eth.KeepAddress([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	digest := []byte{1}

	expectedError := "keep not found for address [0x0000000000000000000000000000000000000001]"
	err := chain.requestSignature(keepAddress, digest)

	if err.Error() != expectedError {
		t.Fatalf(
			"unexpected error\nexpected: [%v]\nactual:   [%v]",
			expectedError,
			err.Error(),
		)
	}
}

// func TestRequestSignatureNoHandler(t *testing.T) {
// 	// handler := func(event *eth.SignatureRequestedEvent)
// }

// func TestRequestSignature(t *testing.T) {

// }
