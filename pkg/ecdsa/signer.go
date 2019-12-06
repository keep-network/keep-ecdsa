package ecdsa

import (
	cecdsa "crypto/ecdsa"
	"io"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// Signer is used to calculate a signature. It holds an ECDSA private key.
type Signer struct {
	privateKey *cecdsa.PrivateKey
}

// PublicKey holds a public key in a form of X and Y coordinates of a point on
// an elliptic curve.
type PublicKey cecdsa.PublicKey

// NewSigner creates a new Signer and initializes it with a provided ECDSA
// private key.
func NewSigner(privateKey *cecdsa.PrivateKey) *Signer {
	return &Signer{privateKey: privateKey}
}

// GenerateKey generates an ECDSA private key. It utilizes go-ethereum's secp256k1
// elliptic curve implementation.
func GenerateKey(rand io.Reader) (*cecdsa.PrivateKey, error) {
	return cecdsa.GenerateKey(secp256k1.S256(), rand)
}

// PublicKey returns Signer's ECDSA public key.
func (s *Signer) PublicKey() *PublicKey {
	return (*PublicKey)(&s.privateKey.PublicKey)
}

// Curve returns elliptic curve instance used by the Signer.
func (s *Signer) Curve() *secp256k1.BitCurve {
	return s.privateKey.Curve.(*secp256k1.BitCurve)
}

// Marshal serializes Public Key to bytes in uncompressed form as described in
// [SEC 1] section 2.3.3: `04 + <x coordinate> + <y coordinate>`
func (pk *PublicKey) Marshal() []byte {
	return crypto.FromECDSAPub((*cecdsa.PublicKey)(pk))
}
