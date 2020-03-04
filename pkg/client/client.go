// Package client defines ECDSA keep client.
package client

import (
	"context"
	"time"

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

const (
	KeyGenerationTimeout = 150 * time.Minute
	SigningTimeout       = 90 * time.Minute
)

// Initialize initializes the ECDSA client with rules related to events handling.
// Expects a slice of sanctioned applications selected by the operator for which
// operator will be registered as a member candidate.
func Initialize(
	ctx context.Context,
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
				go monitorKeepClosedEvents(
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

			keygenCtx, cancel := context.WithTimeout(ctx, KeyGenerationTimeout)
			defer cancel()

			var signer *tss.ThresholdSigner
			var err error

			for {
				if keygenCtx.Err() != nil {
					logger.Errorf("key generation timeout exceeded")
					return
				}

				memberIDs, err := tssNode.AnnounceSignerPresence(
					keygenCtx,
					operatorPublicKey,
					event.KeepAddress,
					event.Members,
				)
				if err != nil {
					logger.Errorf("announce signer presence failed: [%v]", err)
					continue
				}

				signer, err = tssNode.GenerateSignerForKeep(
					keygenCtx,
					operatorPublicKey,
					event.KeepAddress,
					memberIDs,
				)
				if err != nil {
					logger.Errorf("signer generation failed: [%v]", err)
					continue
				}

				break
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

			go monitorKeepClosedEvents(
				ethereumChain,
				event.KeepAddress,
				keepsRegistry,
			)
		}
	})

	for _, application := range sanctionedApplications {
		go checkStatusAndRegisterForApplication(ctx, ethereumChain, application)
	}
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
				signingCtx, cancel := context.WithTimeout(context.Background(), SigningTimeout)
				defer cancel()

				for {
					if signingCtx.Err() != nil {
						logger.Errorf("signing timeout exceeded")
						return
					}

					err := tssNode.CalculateSignature(
						signingCtx,
						signer,
						signatureRequestedEvent.Digest,
					)

					if err != nil {
						logger.Errorf("signature calculation failed: [%v]", err)
					}

					break
				}
			}()
		},
	)
}

// monitorKeepClosedEvents subscribes and unsubscribes for keep closed events.
func monitorKeepClosedEvents(
	ethereumChain eth.Handle,
	keepAddress common.Address,
	keepsRegistry *registry.Keeps,
) {
	keepClosed := make(chan *eth.KeepClosedEvent)

	subscriptionOnKeepClosed, err := ethereumChain.OnKeepClosed(
		keepAddress,
		func(event *eth.KeepClosedEvent) {
			logger.Infof("keep [%s] is being closed", keepAddress.String())
			keepsRegistry.UnregisterKeep(keepAddress)
			keepClosed <- event
		},
	)
	if err != nil {
		logger.Errorf(
			"failed on registering for keep closed event: [%v]",
			err,
		)
	}

	defer subscriptionOnKeepClosed.Unsubscribe()

	<-keepClosed

	logger.Info(
		"unsubscribing from KeepClosed event",
	)

}
