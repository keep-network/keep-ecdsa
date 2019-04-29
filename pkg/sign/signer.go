package sign

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"math/big"

	"github.com/keep-network/go-ethereum/crypto/secp256k1"
)

// Signer is used to calculate a signature. It holds an ECDSA private key.
type Signer struct {
	privateKey *ecdsa.PrivateKey
}

// PublicKey holds a public key in a form of two big.Int X and Y values.
type PublicKey struct {
	X *big.Int
	Y *big.Int
}

// NewSigner creates a new Signer and initializes it with random private and
// public keys. It utilizes go-ethereum's secp256k1 elliptic curve implementation.
func NewSigner() (*Signer, error) {
	signer := &Signer{}

	err := signer.generateKey()
	if err != nil {
		return nil, err
	}

	return signer, nil
}

func (s *Signer) generateKey() error {
	privateKey, err := ecdsa.GenerateKey(secp256k1.S256(), crand.Reader)
	if err != nil {
		return err
	}

	s.privateKey = privateKey

	return nil
}

// PublicKey returns Signer's public key as a pair of X and Y coordinates.
func (s *Signer) PublicKey() *PublicKey {
	return &PublicKey{
		X: s.privateKey.PublicKey.X,
		Y: s.privateKey.PublicKey.Y,
	}
}
