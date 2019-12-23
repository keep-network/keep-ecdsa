package tss

import (
	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	tssLib "github.com/binance-chain/tss-lib/tss"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

// ThresholdSigner is a threshold signer who completed key generation stage.
type ThresholdSigner struct {
	*groupInfo

	// keygenData contains output of key generation stage. This data should be
	// persisted to local storage.
	keygenData keygen.LocalPartySaveData
}

// MemberID returns member's unique identifer.
func (s *ThresholdSigner) MemberID() MemberID {
	return s.memberID
}

// GroupID return signing group unique identifer.
func (s *ThresholdSigner) GroupID() string {
	return s.groupID
}

// TODO: Publisher Selection: Temp solution only the first member in the group
// publishes. We need to replace it with proper publisher selection.
func (s *ThresholdSigner) MemberIndex() int {
	for i, memberID := range s.groupMemberIDs {
		if memberID.Equal(s.MemberID()) {
			return i
		}
	}

	return -1
}

// PublicKey returns signer's ECDSA public key which is also the signing group's
// public key.
func (s *ThresholdSigner) PublicKey() *ecdsa.PublicKey {
	pkX, pkY := s.keygenData.ECDSAPub.X(), s.keygenData.ECDSAPub.Y()

	curve := tssLib.EC()
	publicKey := ecdsa.PublicKey{
		Curve: curve,
		X:     pkX,
		Y:     pkY,
	}

	return (*ecdsa.PublicKey)(&publicKey)
}
