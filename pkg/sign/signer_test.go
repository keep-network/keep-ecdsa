package sign

import (
	"crypto/ecdsa"
	"testing"
)

func TestSign(t *testing.T) {
	signer, err := NewSigner()
	if err != nil {
		t.Fatal(err)
	}

	hash := []byte("test hash")

	signature, err := signer.CalculateSignature(hash)
	if err != nil {
		t.Fatal(err)
	}

	if !ecdsa.Verify(
		&signer.privateKey.PublicKey,
		hash,
		signature.R,
		signature.S,
	) {
		t.Errorf("signature is invalid")
	}
}
