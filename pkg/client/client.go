// Package client defines ECDSA keep client.
package client

import (
	"context"
	"time"

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
// Expects a slice of sanctioned applications selected by the operator for which
// operator will be registered as a member candidate.
func Initialize(
	ctx context.Context,
	ethereumChain eth.Handle,
	networkProvider net.Provider,
	persistence persistence.Handle,
	sanctionedApplications []common.Address,
	registrationRetryTicker time.Duration,
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
	ethereumChain.OnECDSAKeepCreated(func(event *eth.ECDSAKeepCreatedEvent) {
		logger.Infof(
			"new keep [%s] created with members: [%x]\n",
			event.KeepAddress.String(),
			event.Members,
		)

		if event.IsMember(ethereumChain.Address()) {
			signer, err := tssNode.GenerateSignerForKeep(
				event.KeepAddress,
				event.Members,
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
	// is already registered. If not checks client's stake and retries registration
	// until the stake allows to complete it.
	for _, application := range sanctionedApplications {
		go func(application common.Address) {
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
		}(application)
	}
}

// registerAsMemberCandidate checks current operator's stake balance and if it's
// positive registers the operator as a member candidate for the given application.
// It retries the operation after each new mined block until the operator is
// successfully registered.
func registerAsMemberCandidate(parentCtx context.Context, ethereumChain eth.Handle, application common.Address) {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	newBlockChan := ethereumChain.WatchBlocks(ctx)

	for {
		select {
		case <-newBlockChan:
			currentStake, err := ethereumChain.EligibleStake()
			if err != nil {
				logger.Warningf(
					"failed to get eligible stake [%s]: [%v]",
					application.String(),
					err,
				)
				continue
			}

			// TODO: Valide if stake is greater than minimum required stake.
			if currentStake.Sign() <= 0 {
				logger.Warningf("operator doesn't have enough stake")
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
