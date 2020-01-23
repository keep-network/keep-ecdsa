// Package client defines ECDSA keep client.
package client

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss"
	"github.com/keep-network/keep-tecdsa/pkg/net"
	"github.com/keep-network/keep-tecdsa/pkg/node"
	"github.com/keep-network/keep-tecdsa/pkg/registry"
)

var logger = log.Logger("keep-tecdsa")

// Initialize initializes the ECDSA client with rules related to events handling.
func Initialize(
	ethereumChain eth.Handle,
	networkProvider net.Provider,
	persistence persistence.Handle,
) {
	keepsRegistry := registry.NewKeepsRegistry(persistence)

	tssNode := node.NewNode(ethereumChain, networkProvider)

	tssNode.InitializeTSSPreParamsPool()

	// Load current keeps' signers from storage and register for signing events.
	keepsRegistry.LoadExistingKeeps()

	keepsRegistry.ForEachKeep(
		func(keepAddress common.Address, signer []*tss.ThresholdSigner) {
			for _, signer := range signer {
				registerForSignEvents(
					ethereumChain,
					tssNode,
					keepAddress,
					signer,
				)
				logger.Debugf(
					"registered for events from keep [%s] with signer: [%s]",
					keepAddress.String(),
					signer.MemberID(),
				)
			}
		},
	)

	// Watch for new keeps creation.
	ethereumChain.OnECDSAKeepCreated(func(event *eth.ECDSAKeepCreatedEvent) {
		logger.Infof(
			"new keep [%s] created with members: [%x]\n",
			event.KeepAddress.String(),
			event.Members,
		)

		for memberIndex, memberAddress := range event.Members {
			if memberAddress == ethereumChain.Address() {
				go func(keepMemberIndex uint) {
					logger.Infof(
						"generating signer for keep [%s] with member index [%d]",
						event.KeepAddress.String(),
						keepMemberIndex,
					)

					signer, err := tssNode.GenerateSignerForKeep(
						event.KeepAddress,
						event.Members,
						keepMemberIndex,
					)
					if err != nil {
						logger.Errorf("signer generation failed: [%v]", err)
						return
					}

					logger.Infof(
						"initialized signer for keep [%s] with member index [%d]",
						event.KeepAddress.String(),
						keepMemberIndex,
					)

					err = keepsRegistry.RegisterSigner(event.KeepAddress, signer)
					if err != nil {
						logger.Errorf(
							"failed to register threshold signer for keep [%s]: [%v]",
							event.KeepAddress.String(),
							err,
						)
					}

					registerForSignEvents(
						ethereumChain,
						tssNode,
						event.KeepAddress,
						signer,
					)
				}(uint(memberIndex))
			}
		}
	})

	// Register client as a candidate member for keep.
	if err := ethereumChain.RegisterAsMemberCandidate(); err != nil {
		logger.Errorf("failed to register member: [%v]", err)
	}

	logger.Infof("client registered as member candidate in keep factory")
}

// registerForSignEvents registers for signature requested events emitted by
// specific keep contract.
func registerForSignEvents(
	ethereumChain eth.Handle,
	tssNode *node.Node,
	keepAddress common.Address,
	signer *tss.ThresholdSigner,
) {
	ethereumChain.OnSignatureRequested(
		keepAddress,
		func(signatureRequestedEvent *eth.SignatureRequestedEvent) {
			logger.Infof(
				"new signature requested from keep [%s] for digest: [%+x]",
				keepAddress.String(),
				signatureRequestedEvent.Digest,
			)

			go func() {
				err := tssNode.CalculateSignature(
					signer,
					signatureRequestedEvent.Digest,
				)

				if err != nil {
					logger.Errorf("signature calculation failed: [%v]", err)
				}
			}()
		},
	)
}
