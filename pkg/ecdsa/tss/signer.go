package tss

import (
	cecdsa "crypto/ecdsa"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	tssLib "github.com/binance-chain/tss-lib/tss"
)

// ThresholdSigner is a threshold signer who completed key generation stage.
type ThresholdSigner struct {
	*groupInfo

	// thresholdKey contains a signer's key generated for a threshold signing
	// scheme. This data should be persisted to a local storage.
	thresholdKey ThresholdKey
}

// ThresholdKey contains data of signer's threshold key.
type ThresholdKey keygen.LocalPartySaveData

// MemberID returns member's unique identifer.
func (s *ThresholdSigner) MemberID() MemberID {
	return s.memberID
}

// GroupID return signing group unique identifer.
func (s *ThresholdSigner) GroupID() string {
	return s.groupID
}

// PublicKey returns signer's ECDSA public key which is also the signing group's
// public key.
func (s *ThresholdSigner) PublicKey() *cecdsa.PublicKey {
	pkX, pkY := s.thresholdKey.ECDSAPub.X(), s.thresholdKey.ECDSAPub.Y()

	curve := tssLib.EC()
	publicKey := cecdsa.PublicKey{
		Curve: curve,
		X:     pkX,
		Y:     pkY,
	}

	return &publicKey
}
