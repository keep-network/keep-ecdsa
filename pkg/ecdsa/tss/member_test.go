package tss

import (
	"math/big"
	"testing"
)

func TestMemberIDConversions(t *testing.T) {
	memberID := MemberID(1234567890)

	expectedString := "1234567890"
	expectedBigInt := big.NewInt(1234567890)

	if memberID.String() != expectedString {
		t.Errorf(
			"invalid string\nexpected: %v\nactual:   %v\n",
			expectedString,
			memberID.String(),
		)
	}

	if memberID.bigInt().Cmp(expectedBigInt) != 0 {
		t.Errorf(
			"invalid big int\nexpected: %v\nactual:   %v\n",
			expectedBigInt,
			memberID.bigInt(),
		)
	}

	bytes := memberID.bytes()
	fromBytes := memberIDFromBytes(bytes)
	if memberID != fromBytes {
		t.Errorf(
			"invalid member id from bytes\nexpected: %v\nactual:   %v\n",
			memberID,
			fromBytes,
		)
	}
}
