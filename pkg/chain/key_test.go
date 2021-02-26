package chain

import (
	"bytes"
	"math/big"
	"testing"

	cecdsa "crypto/ecdsa"
)

func TestSerializePublicKey(t *testing.T) {
	bytes32 := []byte{207, 73, 229, 19, 136, 216, 125, 157, 135, 142, 67, 130,
		136, 13, 76, 188, 32, 218, 243, 134, 95, 73, 155, 24, 38, 73, 117, 90,
		215, 95, 216, 19}
	bytes31 := []byte{182, 142, 176, 51, 131, 130, 111, 197, 191, 103, 180, 137,
		171, 101, 34, 78, 251, 234, 118, 184, 16, 116, 238, 82, 131, 153, 134,
		17, 46, 158, 94}

	expectedResult := [64]byte{
		// bytes32
		207, 73, 229, 19, 136, 216, 125, 157, 135, 142, 67, 130, 136, 13, 76,
		188, 32, 218, 243, 134, 95, 73, 155, 24, 38, 73, 117, 90, 215, 95, 216,
		19,
		// padding
		00,
		// bytes31
		182, 142, 176, 51, 131, 130, 111, 197, 191, 103, 180, 137, 171, 101, 34,
		78, 251, 234, 118, 184, 16, 116, 238, 82, 131, 153, 134, 17, 46, 158, 94,
	}

	actualResult, err := SerializePublicKey(
		&cecdsa.PublicKey{
			X: new(big.Int).SetBytes(bytes32),
			Y: new(big.Int).SetBytes(bytes31),
		},
	)

	if !bytes.Equal(expectedResult[:], actualResult[:]) {
		t.Errorf(
			"unexpected result\nexpected: [%+v]\nactual:   [%+v]",
			expectedResult,
			actualResult,
		)
	}

	if err != nil {
		t.Errorf("unexpected error [%+v]", err)

	}
}
