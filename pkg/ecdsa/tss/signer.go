package tss

import (
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-tecdsa/pkg/net"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	tssLib "github.com/binance-chain/tss-lib/tss"
)

// Signer is threshold signer.
type Signer struct {
	tssParameters *tssLib.Parameters
	keygenParty   tssLib.Party
	// Network channels used for messages transport.
	networkProvider net.Provider
	// Channels where results of the key generation protocol execution will be written to.
	keygenEndChan <-chan keygen.LocalPartySaveData // data from a successful execution
	keygenErrChan chan error                       // errors emitted during the protocol execution
	// keygenData contains output of key generation stage. This data should be
	// persisted to local storage.
	keygenData keygen.LocalPartySaveData
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
