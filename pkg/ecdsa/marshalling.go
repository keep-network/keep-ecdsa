package ecdsa

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/keep-network/keep-tecdsa/pkg/registry/gen/pb"
)

// Marshal converts Signer to byte array.
func (s *Signer) Marshal() ([]byte, error) {
	return (&pb.Signer{
		PrivateKey: crypto.FromECDSA(s.privateKey),
	}).Marshal()
}

// Unmarshal converts a byte array back to Signer.
func (s *Signer) Unmarshal(bytes []byte) error {
	pbSigner := pb.Signer{}
	if err := pbSigner.Unmarshal(bytes); err != nil {
		return err
	}

	privateKey, err := crypto.ToECDSA(pbSigner.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to decode private key: [%v]", err)
	}

	s.privateKey = privateKey

	return nil
}
