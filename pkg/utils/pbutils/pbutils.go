package pbutils

import "github.com/gogo/protobuf/proto"

// RoundTrip is code borrowed from keep-core/pkg/internal/pbutils.
// TODO: Extract this code from `keep-core` to `keep-common`.
func RoundTrip(
	marshaler proto.Marshaler,
	unmarshaler proto.Unmarshaler,
) error {
	bytes, err := marshaler.Marshal()
	if err != nil {
		return err
	}

	err = unmarshaler.Unmarshal(bytes)
	if err != nil {
		return err
	}

	return nil
}
