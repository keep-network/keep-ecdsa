// Package byteutils provides helper utilities for working with bytes
package byteutils

import (
	"fmt"
)

// LeftPadTo32Bytes appends zeros to bytes slice to make it exactly 32 bytes long.
// TODO: This is copied from keep-core. Consider extracting utils to a separate repo.
func LeftPadTo32Bytes(bytes []byte) ([]byte, error) {
	expectedByteLen := 32
	if len(bytes) > expectedByteLen {
		return nil, fmt.Errorf(
			"cannot pad %v byte array to %v bytes", len(bytes), expectedByteLen,
		)
	}

	result := make([]byte, 0)
	if len(bytes) < expectedByteLen {
		result = make([]byte, expectedByteLen-len(bytes))
	}
	result = append(result, bytes...)

	return result, nil
}

// BytesTo32Byte converts bytes slice to a 32-byte array. It left pads the array
// with zeros in case of a slice shorter than 32-byte.
func BytesTo32Byte(bytes []byte) ([32]byte, error) {
	var result [32]byte

	paddedBytes, err := LeftPadTo32Bytes(bytes)
	if err != nil {
		return result, err
	}

	copy(result[:], paddedBytes[:32])

	return result, nil
}
