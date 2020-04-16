package pbutils

import (
	"github.com/gogo/protobuf/proto"
	fuzz "github.com/google/gofuzz"
)

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

// FuzzUnmarshaler tests given unmarshaler with random bytes.
func FuzzUnmarshaler(unmarshaler proto.Unmarshaler) {
	for i := 0; i < 100; i++ {
		var messageBytes []byte

		f := fuzz.New().NilChance(0.01).NumElements(0, 512)
		f.Fuzz(&messageBytes)

		_ = unmarshaler.Unmarshal(messageBytes)
	}
}
