package tss

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/binance-chain/tss-lib/ecdsa/signing"
	"github.com/binance-chain/tss-lib/tss"
	tssLib "github.com/binance-chain/tss-lib/tss"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

const signingTimeout = 120 * time.Second

// initializeSigning initializes a member to run a threshold multi-party signature
// calculation protocol. Signature will be calculated for provided digest.
func (s *ThresholdSigner) initializeSigning(
	digest []byte,
	netBridge *networkBridge,
) (*signingSigner, error) {
	ctx, cancel := context.WithTimeout(context.Background(), signingTimeout)

	digestInt := new(big.Int).SetBytes(digest)

	party, endChan, err := s.initializeSigningParty(
		ctx,
		digestInt,
		netBridge,
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize signing party: [%v]", err)
	}

	return &signingSigner{
		groupInfo:      s.groupInfo,
		context:        ctx,
		contextCancel:  cancel,
		networkBridge:  netBridge,
		signingParty:   party,
		signingEndChan: endChan,
	}, nil
}

// signingSigner represents Signer who initialized signing stage and is ready to
// start signature calculation.
type signingSigner struct {
	*groupInfo

	// Context for protocol execution with a timeout.
	context       context.Context
	contextCancel context.CancelFunc
	networkBridge *networkBridge
	// Signing
	signingParty tssLib.Party
	// Channels where results of the signing protocol execution will be written to.
	signingEndChan <-chan signing.SignatureData // data from a successful execution
}

// sign executes the protocol to calculate a signature. This function needs to be
// executed only after all members finished the initialization stage. As a result
// the calculated ECDSA signature will be returned or an error, if the signature
// generation failed.
func (s *signingSigner) sign() (*ecdsa.Signature, error) {
	defer s.networkBridge.close()
	
	defer s.contextCancel()

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
		case <-s.context.Done():
			memberIDs := []MemberID{}

			if s.signingParty.WaitingFor() != nil {
				for _, partyID := range s.signingParty.WaitingFor() {
					memberIDs = append(memberIDs, MemberID(partyID.GetId()))
				}
			}

			return nil, timeoutError{signingTimeout, "signing", memberIDs}
		}
	}
}

func (s *ThresholdSigner) initializeSigningParty(
	ctx context.Context,
	digest *big.Int,
	netBridge *networkBridge,
) (
	tssLib.Party,
	<-chan signing.SignatureData,
	error,
) {
	tssMessageChan := make(chan tss.Message, len(s.groupMemberIDs))
	endChan := make(chan signing.SignatureData)

	currentPartyID, groupPartiesIDs, err := generatePartiesIDs(
		s.memberID,
		s.groupMemberIDs,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate parties IDs: [%v]", err)
	}

	params := tss.NewParameters(
		tss.NewPeerContext(tss.SortPartyIDs(groupPartiesIDs)),
		currentPartyID,
		len(groupPartiesIDs),
		s.dishonestThreshold,
	)

	party := signing.NewLocalParty(
		digest,
		params,
		s.keygenData,
		tssMessageChan,
		endChan,
	)

	if err := netBridge.connect(
		ctx,
		tssMessageChan,
		party,
		params.Parties().IDs(),
	); err != nil {
		return nil, nil, fmt.Errorf("failed to connect bridge network: [%v]", err)
	}

	return party, endChan, nil
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
