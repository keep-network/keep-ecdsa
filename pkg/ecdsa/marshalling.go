package ecdsa

import (
	cecdsa "crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"

	"github.com/keep-network/keep-tecdsa/pkg/registry/gen/pb"
)

// Marshal converts Signer to byte array.
func (s *Signer) Marshal() ([]byte, error) {
	return (&pb.Signer{
		PrivateKey: s.privateKey.D.Text(16),
	}).Marshal()
}

// Unmarshal converts a byte array back to Signer.
func (s *Signer) Unmarshal(bytes []byte) error {
	pbSigner := pb.Signer{}
	if err := pbSigner.Unmarshal(bytes); err != nil {
		return err
	}

	privateKeyD, ok := new(big.Int).SetString(pbSigner.PrivateKey, 16)
	if !ok {
		return fmt.Errorf("failed to set private key from string")
	}

	privateKey := &cecdsa.PrivateKey{}
	curve := secp256k1.S256()

	privateKey.PublicKey.Curve = curve
	privateKey.D = privateKeyD
	privateKey.PublicKey.X, privateKey.PublicKey.Y = curve.ScalarBaseMult(privateKeyD.Bytes())

	s.privateKey = privateKey

	return nil
}
