package sign

import (
	"crypto/ecdsa"
	"testing"
)

func TestSign(t *testing.T) {
	privateKey,err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	signer, err := NewSigner(privateKey)
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
