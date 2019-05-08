package sign

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"testing"
)

func TestSign(t *testing.T) {
	privateKey, err := GenerateKey(crand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	signer := NewSigner(privateKey)

	hash := []byte("test hash")

	signature, err := signer.CalculateSignature(crand.Reader, hash)
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
