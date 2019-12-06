package tss

import (
	"fmt"
	"math/big"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/ecdsa/signing"
	"github.com/binance-chain/tss-lib/tss"
	tssLib "github.com/binance-chain/tss-lib/tss"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-tecdsa/pkg/net"
)

// InitializeSigning initializes a member to run a threshold multi-party signature
// calculation protocol. Signature will be calculated for provided digest.
//
// Network channel should support broadcast and unicast messages transport.
func (s *Signer) InitializeSigning(
	digest []byte,
	networkChannel net.NetworkChannel,
) *SigningSigner {
	digestInt := new(big.Int).SetBytes(digest)

	errChan := make(chan error)
	doneChan := make(chan struct{})

	party, endChan := initializeSigningParty(
		digestInt,
		s.tssParameters,
		s.keygenData,
		networkChannel,
		errChan,
	)

	return &SigningSigner{
		signingParty:   party,
		signingEndChan: endChan,
		signingErrChan: errChan,
		doneChan:       doneChan,
	}, nil
}

// SigningSigner represents Signer who initialized signing stage and is ready to
// start signature calculation.
type SigningSigner struct {
	// Signing
	signingParty tssLib.Party
	// Channels where results of the signing protocol execution will be written to.
	signingEndChan <-chan signing.SignatureData // data from a successful execution
	signingErrChan <-chan error                 // errors emitted during the protocol execution

	doneChan chan struct{}
}

// Sign executes the protocol to calculate a signature. This function needs to be
// executed only after all members finished the initialization stage. As a result
// the calculated ECDSA signature will be returned.
func (s *SigningSigner) Sign() (*ecdsa.Signature, error) {
	defer close(s.doneChan)

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

func initializeSigningParty(
	digest *big.Int,
	tssParameters *tss.Parameters,
	keygenData keygen.LocalPartySaveData,
	networkChannel net.NetworkChannel,
	errChan chan error,
) (tss.Party, <-chan signing.SignatureData) {
	outChan := make(chan tssLib.Message)
	endChan := make(chan signing.SignatureData)

	recvChan := make(chan net.Message, tssParameters.PartyCount())
	networkChannel.Receive(func(message net.Message) error {
		recvChan <- message
		return nil
	})

	party := signing.NewLocalParty(digest, tssParameters, keygenData, outChan, endChan)

	go handleMessages(party, tssParameters, networkChannel, outChan, recvChan, errChan)

	return party, endChan
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
