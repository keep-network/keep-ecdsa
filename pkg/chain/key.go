package chain

import (
	cecdsa "crypto/ecdsa"

	"github.com/keep-network/keep-ecdsa/pkg/utils/byteutils"
)

// SerializePublicKey takes X and Y coordinates of a signer's public key and
// concatenates it to a 64-byte long array. If any of coordinates is shorter
// than 32-byte it is preceded with zeros.
func SerializePublicKey(publicKey *cecdsa.PublicKey) ([64]byte, error) {
	var serialized [64]byte

	x, err := byteutils.LeftPadTo32Bytes(publicKey.X.Bytes())
	if err != nil {
		return serialized, err
	}

	y, err := byteutils.LeftPadTo32Bytes(publicKey.Y.Bytes())
	if err != nil {
		return serialized, err
	}

	serializedBytes := append(x, y...)

	copy(serialized[:], serializedBytes)

	return serialized, nil
}
