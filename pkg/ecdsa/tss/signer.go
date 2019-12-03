package tss

import (
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	tssLib "github.com/binance-chain/tss-lib/tss"
)

// Signer is a threshold signer who completed key generation stage.
type Signer struct {
	BaseMember

	tssParameters *tssParameters
	// keygenData contains output of key generation stage. This data should be
	// persisted to local storage.
	keygenData keygen.LocalPartySaveData
}

type tssParameters struct {
	currentPartyID *tssLib.PartyID
	sortedPartyIDs []*tssLib.PartyID
	threshold      int
}

// PublicKey returns Signer's ECDSA public key.
func (s *Signer) PublicKey() *ecdsa.PublicKey {
	pkX, pkY := s.keygenData.ECDSAPub.X(), s.keygenData.ECDSAPub.Y()

	curve := tssLib.EC()
	publicKey := ecdsa.PublicKey{
		Curve: curve,
		X:     pkX,
		Y:     pkY,
	}

	return (*ecdsa.PublicKey)(&publicKey)
}
