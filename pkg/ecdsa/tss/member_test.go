package tss

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMemberIDFromHex(t *testing.T) {
	var tests = map[string]struct {
		inputHex       string
		expectedString string
		expectedError  error
	}{
		"converts mixed case hex string": {
			inputHex:       "0123456789aAbBcCdDeEfF",
			expectedString: "0123456789aabbccddeeff",
		},
		"converts 0x prefixed string": {
			inputHex:       "0x0123456789aAbBcCdDeEfF",
			expectedString: "0123456789aabbccddeeff",
		},
		"converts 0X prefixed string": {
			inputHex:       "0X0123456789aAbBcCdDeEfF",
			expectedString: "0123456789aabbccddeeff",
		},
		"fails for empty string": {
			inputHex:      "",
			expectedError: fmt.Errorf("empty string"),
		},
		"fails for odd length": {
			inputHex:      "123456789aAbBcCdDeEfF",
			expectedError: fmt.Errorf("failed to decode string: [encoding/hex: odd length hex string]"),
		},
		"fails for non hex character": {
			inputHex:      "0123456789aAbBcCdDeEfZ",
			expectedError: fmt.Errorf("failed to decode string: [encoding/hex: invalid byte: U+005A 'Z']"),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			memberID, err := MemberIDFromHex(test.inputHex)
			if !reflect.DeepEqual(err, test.expectedError) {
				t.Errorf(
					"invalid error\nexpected: %v\nactual:   %v\n",
					test.expectedError,
					err,
				)
			}

			if memberID.String() != test.expectedString {
				t.Errorf(
					"invalid converted memberID\nexpected: %v\nactual:   %v\n",
					test.expectedString,
					memberID.String(),
				)
			}
		})
	}
}
