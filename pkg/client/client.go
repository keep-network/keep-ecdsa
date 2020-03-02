// Package client defines ECDSA keep client.
package client

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-core/pkg/net"
	"github.com/keep-network/keep-core/pkg/operator"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss"
	"github.com/keep-network/keep-tecdsa/pkg/node"
	"github.com/keep-network/keep-tecdsa/pkg/registry"
)

var logger = log.Logger("keep-tecdsa")

// Initialize initializes the ECDSA client with rules related to events handling.
// Expects a slice of sanctioned applications selected by the operator for which
// operator will be registered as a member candidate.
func Initialize(
	operatorPublicKey *operator.PublicKey,
	ethereumChain eth.Handle,
	networkProvider net.Provider,
	persistence persistence.Handle,
	sanctionedApplications []common.Address,
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
				registerForKeepClosedEvents(
					ethereumChain,
					keepAddress,
					keepsRegistry,
				)
				logger.Debugf(
					"signer registered for events from keep: [%s]",
					keepAddress.String(),
				)
			}
		},
	)

	// Watch for new keeps creation.
	ethereumChain.OnBondedECDSAKeepCreated(func(event *eth.BondedECDSAKeepCreatedEvent) {
		logger.Infof(
			"new keep [%s] created with members: [%x]\n",
			event.KeepAddress.String(),
			event.Members,
		)

		if event.IsMember(ethereumChain.Address()) {
			logger.Infof(
				"member [%s] is starting signer generation for keep [%s]...",
				ethereumChain.Address().String(),
				event.KeepAddress.String(),
			)

			memberIDs, err := tssNode.AnnounceSignerPresence(
				operatorPublicKey,
				event.KeepAddress,
				event.Members,
			)

			signer, err := tssNode.GenerateSignerForKeep(
				operatorPublicKey,
				event.KeepAddress,
				memberIDs,
			)
			if err != nil {
				logger.Errorf("signer generation failed: [%v]", err)
				return
			}

			logger.Infof("initialized signer for keep [%s]", event.KeepAddress.String())

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

			registerForKeepClosedEvents(
				ethereumChain,
				event.KeepAddress,
				keepsRegistry,
			)
		}
	})

	// Register client as a candidate member for keep.
	for _, application := range sanctionedApplications {
		// TODO: Validate if client is already registered and can be registered.
		// If can register but it is not registered, it is registering. If can't
		// be registered yet (stake maturation period), waits some time and tries again
		if err := ethereumChain.RegisterAsMemberCandidate(application); err != nil {
			logger.Errorf(
				"failed to register member for application [%s]: [%v]",
				application.String(),
				err,
			)
			continue
		}
		logger.Debugf(
			"client registered as member candidate for application: [%s]",
			application.String(),
		)
	}

	logger.Infof("client initialized")
}


// registerForKeepClosedEvents registers for keep closing events.
func registerForKeepClosedEvents(
	ethereumChain eth.Handle,
	keepAddress common.Address,
	keepsRegistry *registry.Keeps,
) {
	ethereumChain.OnKeepClosed(
		keepAddress,
		func(event *eth.KeepClosedEvent) {
			logger.Infof("keep [%v] is being closed", keepAddress)
			keepsRegistry.UnregisterKeep(keepAddress)
		})
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
