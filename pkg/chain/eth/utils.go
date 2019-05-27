package eth

import (
	"crypto/ecdsa"

	"github.com/keep-network/keep-tecdsa/pkg/utils/byteutils"
)

// ECDSAPublicKeyToKeepPublicKey serializes ECDSA public key to a format required
// for public key by a keep contract. It pads each of X, Y coordinates to 32-bytes
// and returns a 64-byte long serialized public key.
func ECDSAPublicKeyToKeepPublicKey(
	publicKey *ecdsa.PublicKey) (KeepPublicKey, error) {
	var serialized KeepPublicKey

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
