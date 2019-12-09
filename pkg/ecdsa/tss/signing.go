package tss

import (
	"fmt"
	"math/big"

	"github.com/binance-chain/tss-lib/ecdsa/signing"
	tssLib "github.com/binance-chain/tss-lib/tss"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

// InitializeSigning initializes a member to run a threshold multi-party signature
// calculation protocol. Signature will be calculated for provided digest.
func (s *Signer) InitializeSigning(
	digest []byte,
	networkBridge *NetworkBridge,
) (*SigningSigner, error) {
	digestInt := new(big.Int).SetBytes(digest)

	party, endChan, errChan, err := s.initializeSigningParty(
		digestInt,
		networkBridge,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize signing party: [%v]", err)
	}

	return &SigningSigner{
		GroupInfo:      s.GroupInfo,
		networkBridge:  networkBridge,
		signingParty:   party,
		signingEndChan: endChan,
		signingErrChan: errChan,
	}, nil
}

// SigningSigner represents Signer who initialized signing stage and is ready to
// start signature calculation.
type SigningSigner struct {
	GroupInfo

	networkBridge *NetworkBridge
	// Signing
	signingParty tssLib.Party
	// Channels where results of the signing protocol execution will be written to.
	signingEndChan <-chan signing.SignatureData // data from a successful execution
	signingErrChan <-chan error                 // error from a failed execution
}

// Sign executes the protocol to calculate a signature. This function needs to be
// executed only after all members finished the initialization stage. As a result
// the calculated ECDSA signature will be returned or error, if the signature
// generation failed.
func (s *SigningSigner) Sign() (*ecdsa.Signature, error) {
	defer s.networkBridge.close()

	if s.signingParty == nil {
		return nil, fmt.Errorf("failed to get initialized signing party")
	}

	if err := s.signingParty.Start(); err != nil {
		return nil, fmt.Errorf(
			"failed to start signing: [%v]",
			s.signingParty.WrapError(err),
		)
	}

	for {
		select {
		case signature := <-s.signingEndChan:
			ecdsaSignature := convertSignatureTSStoECDSA(signature)

			return &ecdsaSignature, nil
		case err := <-s.signingErrChan:
			return nil,
				fmt.Errorf(
					"failed to sign: [%v]",
					s.signingParty.WrapError(err),
				)
		}
	}
}

func (s *Signer) initializeSigningParty(
	digest *big.Int,
	networkBridge *NetworkBridge,
) (
	tssLib.Party,
	<-chan signing.SignatureData,
	chan error,
	error,
) {
	tssMessageChan := make(chan tssLib.Message, s.tssParameters.PartyCount())
	endChan := make(chan signing.SignatureData)
	errChan := make(chan error)

	party := signing.NewLocalParty(digest, s.tssParameters, s.keygenData, tssMessageChan, endChan)

	if err := networkBridge.connect(
		s.GroupInfo.groupID,
		party,
		s.tssParameters,
		tssMessageChan,
		errChan,
	); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect bridge network: [%v]", err)
	}

	return party, endChan, errChan, nil
}

func convertSignatureTSStoECDSA(tssSignature signing.SignatureData) ecdsa.Signature {
	// `SignatureData` contains recovery ID as a byte slice. Only the first byte
	// is relevant and is converted to `int`.
	recoveryBytes := tssSignature.GetSignatureRecovery()
	recoveryInt := int(0)
	recoveryInt = (recoveryInt << 8) | int(recoveryBytes[0])

	return ecdsa.Signature{
		R:          new(big.Int).SetBytes(tssSignature.GetR()),
		S:          new(big.Int).SetBytes(tssSignature.GetS()),
		RecoveryID: recoveryInt,
	}
}
