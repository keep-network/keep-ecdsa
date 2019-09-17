// Package byteutils provides helper utilities for working with bytes
package byteutils

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestLeftPadTo32Bytes(t *testing.T) {
	value32 := []byte{89, 178, 1, 195, 69, 50, 81, 72, 28, 86, 134, 50, 240, 65, 54, 243,
		62, 118, 195, 210, 184, 57, 206, 33, 48, 3, 81, 73, 125, 189, 39, 147,
	}

	tests := map[string]struct {
		value          []byte
		expectedResult []byte
		expectedError  error
	}{
		"valid case of 32-bytes value": {
			value:          value32,
			expectedResult: value32,
		},
		"valid case of 31-bytes value": {
			value:          value32[:31],
			expectedResult: append([]byte{0}, value32[:31]...),
		},
		"invalid case of 33-bytes value": {
			value:          append([]byte{16}, value32...),
			expectedResult: []byte{},
			expectedError:  fmt.Errorf("cannot pad 33 byte array to 32 bytes"),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			actualResult, err := LeftPadTo32Bytes(test.value)

			if !bytes.Equal(test.expectedResult, actualResult) {
				t.Errorf(
					"unexpected result\nexpected: [%+v]\nactual:   [%+v]",
					test.expectedResult,
					actualResult,
				)
			}

			if !reflect.DeepEqual(test.expectedError, err) {
				t.Errorf(
					"unexpected error\nexpected: [%+v]\nactual:   [%+v]",
					test.expectedError,
					err,
				)
			}
		})
	}
}

func TestBytesTo32Byte(t *testing.T) {
	value32 := []byte{89, 178, 1, 195, 69, 50, 81, 72, 28, 86, 134, 50, 240, 65, 54, 243,
		62, 118, 195, 210, 184, 57, 206, 33, 48, 3, 81, 73, 125, 189, 39, 147,
	}

	value32Array := [32]byte{89, 178, 1, 195, 69, 50, 81, 72, 28, 86, 134, 50, 240, 65, 54, 243,
		62, 118, 195, 210, 184, 57, 206, 33, 48, 3, 81, 73, 125, 189, 39, 147,
	}

	value31Array := [32]byte{00, 89, 178, 1, 195, 69, 50, 81, 72, 28, 86, 134, 50, 240, 65, 54, 243,
		62, 118, 195, 210, 184, 57, 206, 33, 48, 3, 81, 73, 125, 189, 39,
	}

	tests := map[string]struct {
		value          []byte
		expectedResult [32]byte
		expectedError  error
	}{
		"valid case of 32-bytes value": {
			value:          value32,
			expectedResult: value32Array,
		},
		"valid case of 31-bytes value": {
			value:          value32[:31],
			expectedResult: value31Array,
		},
		"invalid case of 33-bytes value": {
			value:          append([]byte{16}, value32...),
			expectedResult: [32]byte{},
			expectedError:  fmt.Errorf("cannot pad 33 byte array to 32 bytes"),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			actualResult, err := BytesTo32Byte(test.value)

			if !bytes.Equal(test.expectedResult[:], actualResult[:]) {
				t.Errorf(
					"unexpected result\nexpected: [%+v]\nactual:   [%+v]",
					test.expectedResult,
					actualResult,
				)
			}

			if !reflect.DeepEqual(test.expectedError, err) {
				t.Errorf(
					"unexpected error\nexpected: [%+v]\nactual:   [%+v]",
					test.expectedError,
					err,
				)
			}
		})
	}
}
