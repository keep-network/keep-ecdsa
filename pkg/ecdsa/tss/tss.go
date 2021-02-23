// Package tss contains implementation of Threshold Multi-Party ECDSA Signature
// Scheme. This package uses [tss-lib] protocol implementation based on [GG19].
//
// [tss-lib]: https://github.com/binance-chain/tss-lib.
// [GG19]: Fast Multiparty Threshold ECDSA with Fast Trustless Setup, Rosario
// Gennaro and Steven Goldfeder, 2019, https://eprint.iacr.org/2019/114.pdf.
package tss

import (
	"context"
	cecdsa "crypto/ecdsa"
	"fmt"
	"time"

	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-core/pkg/net"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss/params"
)

const (
	// KeyGenerationProtocolTimeout represents the amount of time before we give up trying to communicate key generation
	KeyGenerationProtocolTimeout = 8 * time.Minute
	// SigningProtocolTimeout represents the amount of time before we give up trying to communicate signing
	SigningProtocolTimeout = 10 * time.Minute
)

var logger = log.Logger("keep-tss")

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
// TSS protocol requires pre-parameters such as safe primes to be generated for
// execution. The parameters should be generated prior to running this function.
// If not provided they will be generated.
//
// As a result a signer will be returned or an error, if key generation failed.
func GenerateThresholdSigner(
	parentCtx context.Context,
	groupID string,
	memberID MemberID,
	groupMemberIDs []MemberID,
	dishonestThreshold uint,
	networkProvider net.Provider,
	pubKeyToAddressFn func(cecdsa.PublicKey) []byte,
	paramsBox *params.Box,
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

	netBridge, err := newNetworkBridge(group, networkProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize network bridge: [%v]", err)
	}

	ctx, cancel := context.WithTimeout(parentCtx, KeyGenerationProtocolTimeout)
	defer cancel()

	preParams, err := paramsBox.Content()
	if err != nil {
		return nil, fmt.Errorf("failed to get pre-parameters: [%v]", err)
	}

	keyGenSigner, err := initializeKeyGeneration(
		ctx,
		group,
		preParams,
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

	if err := readyProtocol(
		ctx,
		group,
		broadcastChannel,
		pubKeyToAddressFn,
	); err != nil {
		return nil, fmt.Errorf("readiness signaling protocol failed: [%v]", err)
	}

	// We are begining the communication with other members using pre-parameters
	// provided inside of this box. It's time to destroy box content so that the
	// pre-parameters cannot be later reused.
	paramsBox.DestroyContent()

	logger.Infof("[party:%s]: starting key generation", keyGenSigner.keygenParty.PartyID())

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
	pubKeyToAddressFn func(cecdsa.PublicKey) []byte,
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

	if err := readyProtocol(
		ctx,
		s.groupInfo,
		broadcastChannel,
		pubKeyToAddressFn,
	); err != nil {
		return nil, fmt.Errorf("readiness signaling protocol failed: [%v]", err)
	}

	signature, err := signingSigner.sign(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: [%v]", err)
	}

	return signature, err
}
