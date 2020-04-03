// Package tss contains implementation of Threshold Multi-Party ECDSA Signature
// Scheme. This package uses [tss-lib] protocol implementation based on [GG19].
//
// [tss-lib]: https://github.com/binance-chain/tss-lib.
// [GG19]: Fast Multiparty Threshold ECDSA with Fast Trustless Setup, Rosario
// Gennaro and Steven Goldfeder, 2019, https://eprint.iacr.org/2019/114.pdf.
package tss

import (
	"context"
	"fmt"
	"time"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-core/pkg/net"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
)

const (
	KeyGenerationProtocolTimeout = 8 * time.Minute
	SigningProtocolTimeout       = 10 * time.Minute
)

var logger = log.Logger("keep-tss")

// Protocol is a structure containing data required to run the TSS protocol.
type Protocol struct {
	// TSS protocol requires pre-parameters such as safe primes to be generated
	// for signer generation. The parameters should be generated and supplied
	// prior to executing the protocol.
	preParams *keygen.LocalPreParams
}

// ProvidePreParams sets protocol pre parameters to the provided ones. It fails
// if the paramters are already set.
func (p *Protocol) ProvidePreParams(preParams *keygen.LocalPreParams) error {
	if p.PreParamsExist() {
		return fmt.Errorf("pre parameters already configured")
	}
	p.preParams = preParams
	return nil
}

// PreParamsExist verifies if pre parameters are already set.
func (p *Protocol) PreParamsExist() bool {
	return p.preParams != nil
}

// flushPreParams removes pre parameters.
func (p *Protocol) flushPreParams() {
	p.preParams = nil
}

// GenerateThresholdSigner executes a threshold multi-party key generation protocol.
//
// It expects unique identifiers of the current member as well as identifiers of
// all members of the signing group. Group ID should be unique for each concurrent
// execution.
//
// Dishonest threshold `t` defines a maximum number of signers controlled by the
// adversary such that the adversary still cannot produce a signature. Any subset
// of `t + 1` players can jointly sign, but any smaller subset cannot.
//
// As a result a signer will be returned or an error, if key generation failed.
func (p *Protocol) GenerateThresholdSigner(
	parentCtx context.Context,
	groupID string,
	memberID MemberID,
	groupMemberIDs []MemberID,
	dishonestThreshold uint,
	networkProvider net.Provider,
) (*ThresholdSigner, error) {
	if len(groupMemberIDs) < 2 {
		return nil, fmt.Errorf(
			"group should have at least 2 members but got: [%d]",
			len(groupMemberIDs),
		)
	}

	if len(groupMemberIDs) <= int(dishonestThreshold) {
		return nil, fmt.Errorf(
			"group size [%d], should be greater than dishonest threshold [%d]",
			len(groupMemberIDs),
			dishonestThreshold,
		)
	}

	group := &groupInfo{
		groupID:            groupID,
		memberID:           memberID,
		groupMemberIDs:     groupMemberIDs,
		dishonestThreshold: int(dishonestThreshold),
	}

	if !p.PreParamsExist() {
		// We expect the params to be pre-generated and provided for the protocol
		// execution.
		return nil, fmt.Errorf("tss pre parameters were not provided")
	}

	netBridge, err := newNetworkBridge(group, networkProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize network bridge: [%v]", err)
	}

	ctx, cancel := context.WithTimeout(parentCtx, KeyGenerationProtocolTimeout)
	defer cancel()

	keyGenSigner, err := initializeKeyGeneration(
		ctx,
		group,
		p.preParams,
		netBridge,
	)
	if err != nil {
		return nil, err
	}
	logger.Infof("[party:%s]: initialized key generation", keyGenSigner.keygenParty.PartyID())

	broadcastChannel, err := netBridge.getBroadcastChannel()
	if err != nil {
		return nil, err
	}

	if err := readyProtocol(ctx, group, broadcastChannel); err != nil {
		return nil, fmt.Errorf("readiness signaling protocol failed: [%v]", err)
	}

	logger.Infof("[party:%s]: starting key generation", keyGenSigner.keygenParty.PartyID())

	// In next steps the protocol members will start exchanging messages based
	// on the generated pre parameters. We flush the pre parameters provided to
	// this function that in case of a failure and retry of the protocol execution
	// new set of parameters is supplied.
	p.flushPreParams()

	signer, err := keyGenSigner.generateKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: [%v]", err)
	}
	logger.Infof("[party:%s]: completed key generation", keyGenSigner.keygenParty.PartyID())

	return signer, nil
}

// CalculateSignature executes a threshold multi-party signature calculation
// protocol for the given digest. As a result the calculated ECDSA signature will
// be returned or an error, if the signature generation failed.
func (s *ThresholdSigner) CalculateSignature(
	parentCtx context.Context,
	digest []byte,
	networkProvider net.Provider,
) (*ecdsa.Signature, error) {
	netBridge, err := newNetworkBridge(s.groupInfo, networkProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize network bridge: [%v]", err)
	}

	ctx, cancel := context.WithTimeout(parentCtx, SigningProtocolTimeout)
	defer cancel()

	signingSigner, err := s.initializeSigning(ctx, digest[:], netBridge)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize signing: [%v]", err)
	}

	broadcastChannel, err := netBridge.getBroadcastChannel()
	if err != nil {
		return nil, err
	}

	if err := readyProtocol(ctx, s.groupInfo, broadcastChannel); err != nil {
		return nil, fmt.Errorf("readiness signaling protocol failed: [%v]", err)
	}

	signature, err := signingSigner.sign(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: [%v]", err)
	}

	return signature, err
}
