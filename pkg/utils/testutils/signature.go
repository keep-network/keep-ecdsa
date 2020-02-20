package testutils

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	cecdsa "github.com/keep-network/keep-tecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-tecdsa/pkg/utils/byteutils"
	"testing"
)

// VerifyEthereumSignature validates that signature in form (r, s, recoveryID)
// is a valid ethereum signature. `SigToPub` is a wrapper on `Ecrecover` that
// allows us to validate a signature in the same way as it's done on-chain, we
// extract public key from the signature and compare it with signer's public key.
func VerifyEthereumSignature(
	t *testing.T,
	hash []byte,
	signature *cecdsa.Signature,
	expectedPublicKey *ecdsa.PublicKey,
) {

	serializedSignature, err := serializeSignature(signature)
	if err != nil {
		t.Fatalf("failed to serialize signature: [%v]", err)
	}

	publicKey, err := crypto.SigToPub(hash, serializedSignature)
	if err != nil {
		t.Fatalf("failed to get public key from signature: [%v]", err)
	}

	if expectedPublicKey.X.Cmp(publicKey.X) != 0 ||
		expectedPublicKey.Y.Cmp(publicKey.Y) != 0 {
		t.Errorf(
			"invalid public key:\nexpected: [%x]\nactual:   [%x]\n",
			expectedPublicKey,
			publicKey,
		)
	}
}

func serializeSignature(signature *cecdsa.Signature) ([]byte, error) {
	var serializedBytes []byte

	r, err := byteutils.LeftPadTo32Bytes(signature.R.Bytes())
	if err != nil {
		return nil, err
	}

	s, err := byteutils.LeftPadTo32Bytes(signature.S.Bytes())
	if err != nil {
		return nil, err
	}

	recoveryID := byte(signature.RecoveryID)

	serializedBytes = append(serializedBytes, r...)
	serializedBytes = append(serializedBytes, s...)
	serializedBytes = append(serializedBytes, recoveryID)

	return serializedBytes, nil
}
