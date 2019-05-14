package sign

import (
	"crypto/ecdsa"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// Signer is used to calculate a signature. It holds an ECDSA private key.
type Signer struct {
	privateKey *ecdsa.PrivateKey
}

// PublicKey holds a public key in a form of X and Y coordinates of a point on
// an elliptic curve.
type PublicKey ecdsa.PublicKey

// Signature holds a signature in a form of two big.Int R and S values.
type Signature struct {
	R *big.Int
	S *big.Int
}

// NewSigner creates a new Signer and initializes it with a provided ECDSA
// private key.
func NewSigner(privateKey *ecdsa.PrivateKey) *Signer {
	return &Signer{privateKey: privateKey}
}

// GenerateKey generates an ECDSA private key. It utilizes go-ethereum's secp256k1
// elliptic curve implementation.
func GenerateKey(rand io.Reader) (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(secp256k1.S256(), rand)
}

// PublicKey returns Signer's public key as a pair of X and Y coordinates.
func (s *Signer) PublicKey() *PublicKey {
	return (*PublicKey)(&s.privateKey.PublicKey)
}

// CalculateSignature returns an ECDSA Signature over provided hash, calculated
// with Signer's private key.
func (s *Signer) CalculateSignature(rand io.Reader, hash []byte) (*Signature, error) {
	sigR, sigS, err := ecdsa.Sign(rand, s.privateKey, hash)
	if err != nil {
		return nil, err
	}

	return &Signature{R: sigR, S: sigS}, nil
}
