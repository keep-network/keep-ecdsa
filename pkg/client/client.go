// Package client defines ECDSA keep client.
package client

import (
	"context"
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
		}
	})

	// Register client as a candidate member for keep. Validates if the client
	// is already registered. If not checks client's stake and retries
	// registration until the stake allows to complete it.
	for _, application := range sanctionedApplications {
		go tryRegisterAsMemberCandidate(ctx, ethereumChain, application)
	}

	logger.Infof("client initialized")
}

// tryRegisterAsMemberCandidate checks if current operator is registered
// as a member candidate for the given application. If not, it triggers
// registration process.
func tryRegisterAsMemberCandidate(
	ctx context.Context,
	ethereumChain eth.Handle,
	application common.Address,
) {
	isRegistered, err := ethereumChain.IsRegistered(application)
	if err != nil {
		logger.Errorf(
			"failed to check if member is registered for application [%s]: [%v]",
			application.String(),
			err,
		)
		return
	}

	if !isRegistered {
		registerAsMemberCandidate(ctx, ethereumChain, application)

		logger.Debugf(
			"client registered as member candidate for application: [%s]",
			application.String(),
		)
	} else {
		logger.Debugf(
			"client is already registered as member candidate for application: [%s]",
			application.String(),
		)
	}

	// TODO: implementation of monitorStake
	// When the operator is registered monitor it's stake.
	// go monitorStake(ctx, ethereumChain, application)
}

// registerAsMemberCandidate checks current operator's stake balance and if it's
// positive registers the operator as a member candidate for the given application.
// It retries the operation after each new mined block until the operator is
// successfully registered.
func registerAsMemberCandidate(
	parentCtx context.Context,
	ethereumChain eth.Handle,
	application common.Address,
) {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	newBlockChan := ethereumChain.WatchBlocks(ctx)

	for {
		select {
		case <-newBlockChan:
			isEligible, err := ethereumChain.IsEligible(application)
			if err != nil {
				logger.Warningf(
					"failed to check operator eligibility for application [%s]: [%v]",
					application.String(),
					err,
				)
				continue
			}

			if !isEligible {
				logger.Warningf(
					"operator isn't eligible for application [%s]",
					application.String(),
				)
				continue
			}

			if err := ethereumChain.RegisterAsMemberCandidate(application); err != nil {
				logger.Warningf(
					"failed to register member for application [%s]: [%v]",
					application.String(),
					err,
				)
				continue
			}

			cancel()
		case <-ctx.Done():
			return
		}
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
