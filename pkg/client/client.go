// Package client defines ECDSA keep client.
package client

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-common/pkg/subscription"
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

	for _, keepAddress := range keepsRegistry.GetKeepsAddresses() {
		isActive, err := ethereumChain.IsActive(keepAddress)
		if err != nil {
			logger.Errorf(
				"failed to verify if keep is still active: [%v]; "+
					"subscriptions for keep signing and closing events are skipped",
				err,
			)

			// If there are no signers for loaded keep that something is clearly
			// wrong. We don't want to continue processing for this keep.
			continue
		}

		if isActive {
			signers, err := keepsRegistry.GetSigners(keepAddress)
			if err != nil {
				logger.Errorf("no signers for keep [%s]", keepAddress.String())
				continue
			}

			for _, signer := range signers {
				subscriptionOnSignatureRequested, err := monitorSigningRequests(
					ethereumChain,
					tssNode,
					keepAddress,
					signer,
				)
				if err != nil {
					logger.Errorf(
						"failed registering for requested signature event for keep [%s]: [%v]",
						keepAddress.String(),
						err,
					)
					// In case of an error we want to avoid subscribing to keep
					// closed events. Something is wrong and we should stop
					// further processing.
					continue
				}
				go monitorKeepClosedEvents(
					ethereumChain,
					keepAddress,
					keepsRegistry,
					subscriptionOnSignatureRequested,
				)
			}
		} else {
			logger.Infof(
				"keep [%s] is no longer active; archiving",
				keepAddress.String(),
			)
			keepsRegistry.UnregisterKeep(keepAddress)
		}
	}

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

			signer, err := generateSignerForKeep(ctx, tssNode, operatorPublicKey, event)
			if err != nil {
				logger.Errorf(
					"failed to generate signer for keep [%s]: [%v]",
					event.KeepAddress.String(),
					err,
				)
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

			subscriptionOnSignatureRequested, err := monitorSigningRequests(
				ethereumChain,
				tssNode,
				event.KeepAddress,
				signer,
			)
			if err != nil {
				logger.Errorf(
					"failed on registering for requested signature event for keep [%s]: [%v]",
					event.KeepAddress.String(),
					err,
				)

				// In case of an error we want to avoid subscribing to keep
				// closed events. Something is wrong and we should stop
				// further processing.
				return
			}

			go monitorKeepClosedEvents(
				ethereumChain,
				event.KeepAddress,
				keepsRegistry,
				subscriptionOnSignatureRequested,
			)
		}
	})

	for _, application := range sanctionedApplications {
		go checkStatusAndRegisterForApplication(ctx, ethereumChain, application)
	}
}

func generateSignerForKeep(
	ctx context.Context,
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	event *eth.BondedECDSAKeepCreatedEvent,
) (*tss.ThresholdSigner, error) {
	keygenCtx, cancel := context.WithTimeout(ctx, KeyGenerationTimeout)
	defer cancel()

	attemptCounter := 0

	for {
		attemptCounter++

		logger.Infof(
			"signer generation for keep [%s]; attempt [%v]",
			event.KeepAddress.String(),
			attemptCounter,
		)

		if keygenCtx.Err() != nil {
			return nil, fmt.Errorf("key generation timeout exceeded")
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

		signer, err := tssNode.GenerateSignerForKeep(
			keygenCtx,
			operatorPublicKey,
			event.KeepAddress,
			memberIDs,
		)
		if err != nil {
			logger.Errorf("signer generation failed: [%v]", err)
			continue
		}

		return signer, nil // key generation succeeded.
	}
}

// monitorSigningRequests registers for signature requested events emitted by
// specific keep contract.
func monitorSigningRequests(
	ethereumChain eth.Handle,
	tssNode *node.Node,
	keepAddress common.Address,
	signer *tss.ThresholdSigner,
) (subscription.EventSubscription, error) {
	return ethereumChain.OnSignatureRequested(
		keepAddress,
		func(event *eth.SignatureRequestedEvent) {
			logger.Infof(
				"new signature requested from keep [%s] for digest: [%+x]",
				keepAddress.String(),
				event.Digest,
			)

			go generateSignatureForKeep(
				tssNode,
				keepAddress,
				signer,
				event,
			)
		},
	)
}

func generateSignatureForKeep(
	tssNode *node.Node,
	keepAddress common.Address,
	signer *tss.ThresholdSigner,
	event *eth.SignatureRequestedEvent,
) {
	signingCtx, cancel := context.WithTimeout(context.Background(), SigningTimeout)
	defer cancel()

	attemptCounter := 0

	for {
		attemptCounter++

		logger.Infof(
			"calculate signature for keep [%s]; attempt [%v]",
			keepAddress.String(),
			attemptCounter,
		)

		if signingCtx.Err() != nil {
			logger.Errorf("signing timeout exceeded")
			return
		}

		err := tssNode.CalculateSignature(
			signingCtx,
			signer,
			event.Digest,
		)
		if err != nil {
			logger.Errorf("signature calculation failed: [%v]", err)
			continue
		}

		return // signature generation succeeded.
	}
}

// monitorKeepClosedEvents subscribes and unsubscribes for keep closed events.
func monitorKeepClosedEvents(
	ethereumChain eth.Handle,
	keepAddress common.Address,
	keepsRegistry *registry.Keeps,
	subscriptionOnSignatureRequested subscription.EventSubscription,
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

		return
	}

	defer subscriptionOnKeepClosed.Unsubscribe()
	defer subscriptionOnSignatureRequested.Unsubscribe()

	<-keepClosed

	logger.Info(
		"unsubscribing from events on keep closure",
	)

}
