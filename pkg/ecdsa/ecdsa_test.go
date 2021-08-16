package ecdsa

import (
	"math/big"
	"testing"
)

func TestSignatureString(t *testing.T) {
	signature := &Signature{
		R:          big.NewInt(1234567890),
		S:          big.NewInt(963852741),
		RecoveryID: 1,
	}
	expectedString := "R: 0x499602d2, S: 0x397339c5, RecoveryID: 1"

	if signature.String() != expectedString {
		t.Errorf(
			"unexpected signature.String() result\n"+
				"expected: [%s]\n"+
				"actual:   [%s]",
			expectedString,
			signature.String(),
		)
	}
}
