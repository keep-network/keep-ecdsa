// Package client defines ECDSA keep client.
package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-core/pkg/net"
	"github.com/keep-network/keep-core/pkg/operator"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss"
	"github.com/keep-network/keep-ecdsa/pkg/node"
	"github.com/keep-network/keep-ecdsa/pkg/registry"
)

var logger = log.Logger("keep-ecdsa")

const (
	keyGenerationTimeout = 150 * time.Minute
	signingTimeout       = 90 * time.Minute
	blockConfirmations   = 12
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
	tssConfig *tss.Config,
) {
	keepsRegistry := registry.NewKeepsRegistry(persistence)

	tssNode := node.NewNode(ethereumChain, networkProvider, tssConfig)

	tssNode.InitializeTSSPreParamsPool()

	requestedSigners := &requestedSignersTrack{
		data:  make(map[string]bool),
		mutex: &sync.Mutex{},
	}
	requestedSignatures := &requestedSignaturesTrack{
		data:  make(map[string]map[string]bool),
		mutex: &sync.Mutex{},
	}

	// Load current keeps' signers from storage and register for signing events.
	keepsRegistry.LoadExistingKeeps()

	confirmIsInactive := func(keepAddress common.Address) bool {
		currentBlock, err := ethereumChain.BlockCounter().CurrentBlock()
		if err != nil {
			logger.Errorf("failed to get current block height [%v]", err)
			return false
		}

		isKeepActive, err := waitForChainConfirmation(
			ethereumChain,
			currentBlock,
			func() (bool, error) {
				return ethereumChain.IsActive(keepAddress)
			},
		)
		if err != nil {
			logger.Errorf(
				"failed to confirm that keep [%s] is inactive: [%v]",
				keepAddress.String(),
				err,
			)
			return false
		}

		return !isKeepActive
	}

	for _, keepAddress := range keepsRegistry.GetKeepsAddresses() {
		go func(keepAddress common.Address) {
			isActive, err := ethereumChain.IsActive(keepAddress)
			if err != nil {
				logger.Errorf(
					"failed to verify if keep is still active: [%v]; "+
						"subscriptions for keep signing and closing events are skipped",
					err,
				)
				return
			}

			if !isActive {
				logger.Infof(
					"keep [%s] seems no longer active; confirming",
					keepAddress.String(),
				)
				if isInactivityConfirmed := confirmIsInactive(keepAddress); isInactivityConfirmed {
					logger.Infof(
						"confirmed that keep [%s] is no longer active; archiving",
						keepAddress.String(),
					)
					keepsRegistry.UnregisterKeep(keepAddress)
					return
				}
				logger.Warningf("keep [%s] is still active", keepAddress.String())
			}

			signers, err := keepsRegistry.GetSigners(keepAddress)
			if err != nil {
				// If there are no signers for loaded keep that something is clearly
				// wrong. We don't want to continue processing for this keep.
				logger.Errorf(
					"no signers for keep [%s]: [%v]",
					keepAddress.String(),
					err,
				)
				return
			}

			for _, signer := range signers {
				subscriptionOnSignatureRequested, err := monitorSigningRequests(
					ethereumChain,
					tssNode,
					keepAddress,
					signer,
					requestedSignatures,
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
					return
				}
				go monitorKeepClosedEvents(
					ethereumChain,
					keepAddress,
					keepsRegistry,
					subscriptionOnSignatureRequested,
				)
				go monitorKeepTerminatedEvent(
					ethereumChain,
					keepAddress,
					keepsRegistry,
					subscriptionOnSignatureRequested,
				)
			}
		}(keepAddress)
	}

	// Watch for new keeps creation.
	ethereumChain.OnBondedECDSAKeepCreated(func(event *eth.BondedECDSAKeepCreatedEvent) {
		logger.Infof(
			"new keep [%s] created with members: [%x]\n",
			event.KeepAddress.String(),
			event.Members,
		)

		if event.IsMember(ethereumChain.Address()) {
			go func(event *eth.BondedECDSAKeepCreatedEvent) {
				if ok := requestedSigners.add(event.KeepAddress); !ok {
					logger.Errorf(
						"keep creation event for keep [%s] already registered",
						event.KeepAddress.String(),
					)
					return
				}
				defer requestedSigners.remove(event.KeepAddress)

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
					requestedSignatures,
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
				go monitorKeepTerminatedEvent(
					ethereumChain,
					event.KeepAddress,
					keepsRegistry,
					subscriptionOnSignatureRequested,
				)
			}(event)
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
	keygenCtx, cancel := context.WithTimeout(ctx, keyGenerationTimeout)
	defer cancel()

	return tssNode.GenerateSignerForKeep(
		keygenCtx,
		operatorPublicKey,
		event.KeepAddress,
		event.Members,
	)
}

// monitorSigningRequests registers for signature requested events emitted by
// specific keep contract.
func monitorSigningRequests(
	ethereumChain eth.Handle,
	tssNode *node.Node,
	keepAddress common.Address,
	signer *tss.ThresholdSigner,
	requestedSignatures *requestedSignaturesTrack,
) (subscription.EventSubscription, error) {
	go checkAwaitingSignature(
		ethereumChain,
		tssNode,
		keepAddress,
		signer,
		requestedSignatures,
	)

	return ethereumChain.OnSignatureRequested(
		keepAddress,
		func(event *eth.SignatureRequestedEvent) {
			logger.Infof(
				"new signature requested from keep [%s] for digest [%+x] at block [%d]",
				keepAddress.String(),
				event.Digest,
				event.BlockNumber,
			)

			go func(event *eth.SignatureRequestedEvent) {
				if ok := requestedSignatures.add(keepAddress, event.Digest); !ok {
					logger.Errorf(
						"signature requested event for keep [%s] and digest [%x] already registered",
						keepAddress.String(),
						event.Digest,
					)
					return
				}
				defer requestedSignatures.remove(keepAddress, event.Digest)

				isAwaitingSignature, err := waitForChainConfirmation(
					ethereumChain,
					event.BlockNumber,
					func() (bool, error) {
						return ethereumChain.IsAwaitingSignature(keepAddress, event.Digest)
					},
				)
				if err != nil {
					logger.Errorf(
						"failed to confirm signing request for digest [%+x] and keep [%s]: [%v]",
						event.Digest,
						keepAddress.String(),
						err,
					)
					return
				}

				if !isAwaitingSignature {
					logger.Warningf(
						"keep [%s] is not awaiting a signature for digest [%+x]",
						keepAddress.String(),
						event.Digest,
					)
					return
				}

				generateSignatureForKeep(tssNode, keepAddress, signer, event.Digest)
			}(event)
		},
	)
}

func checkAwaitingSignature(
	ethereumChain eth.Handle,
	tssNode *node.Node,
	keepAddress common.Address,
	signer *tss.ThresholdSigner,
	requestedSignatures *requestedSignaturesTrack,
) {
	logger.Debugf("checking awaiting signature for keep [%s]", keepAddress.String())

	latestDigest, err := ethereumChain.LatestDigest(keepAddress)
	if err != nil {
		logger.Errorf("could not get latest digest for keep [%s]", keepAddress.String())
		return
	}

	isAwaitingDigest, err := ethereumChain.IsAwaitingSignature(keepAddress, latestDigest)
	if err != nil {
		logger.Errorf(
			"could not check awaiting signature of "+
				"digest [%+x] for keep [%s]",
			latestDigest,
			keepAddress.String(),
		)
		return
	}

	if isAwaitingDigest {
		if ok := requestedSignatures.add(keepAddress, latestDigest); !ok {
			logger.Errorf(
				"signature requested event for keep [%s] and digest [%x] already registered",
				keepAddress.String(),
				latestDigest,
			)
			return
		}
		defer requestedSignatures.remove(keepAddress, latestDigest)

		currentBlock, err := ethereumChain.BlockCounter().CurrentBlock()
		if err != nil {
			logger.Errorf("failed to get current block height: [%v]", err)
			return
		}

		isStillAwaitingSignature, err := waitForChainConfirmation(
			ethereumChain,
			currentBlock,
			func() (bool, error) {
				return ethereumChain.IsAwaitingSignature(keepAddress, latestDigest)
			},
		)
		if err != nil {
			logger.Errorf(
				"failed to confirm signing request for digest [%+x] and keep [%s]: [%v]",
				latestDigest,
				keepAddress.String(),
				err,
			)
			return
		}

		if !isStillAwaitingSignature {
			logger.Warningf(
				"keep [%s] is not awaiting a signature for digest [%+x]",
				keepAddress.String(),
				latestDigest,
			)
			return
		}

		generateSignatureForKeep(tssNode, keepAddress, signer, latestDigest)
	}
}

func generateSignatureForKeep(
	tssNode *node.Node,
	keepAddress common.Address,
	signer *tss.ThresholdSigner,
	digest [32]byte,
) {
	signingCtx, cancel := context.WithTimeout(context.Background(), signingTimeout)
	defer cancel()

	if err := tssNode.CalculateSignature(
		signingCtx,
		signer,
		digest,
	); err != nil {
		logger.Errorf("signature calculation failed: [%v]", err)
	}
}

// monitorKeepClosedEvent monitors KeepClosed event and if that event happens
// unsubscribes from signing event for the given keep and unregisters it from
// the keep registry.
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
			logger.Infof(
				"keep [%s] closed event received at block [%d]",
				keepAddress.String(),
				event.BlockNumber,
			)

			isKeepActive, err := waitForChainConfirmation(
				ethereumChain,
				event.BlockNumber,
				func() (bool, error) {
					return ethereumChain.IsActive(keepAddress)
				},
			)
			if err != nil {
				logger.Errorf(
					"failed to confirm keep [%s] closure: [%v]",
					keepAddress.String(),
					err,
				)
				return
			}

			if isKeepActive {
				logger.Warningf("keep [%s] has not been closed", keepAddress.String())
				return
			}

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

	logger.Info("unsubscribing from events on keep closed")
}

// monitorKeepTerminatedEvent monitors KeepTerminated event and if that event
// happens unsubscribes from signing event for the given keep and unregisters it
// from the keep registry.
func monitorKeepTerminatedEvent(
	ethereumChain eth.Handle,
	keepAddress common.Address,
	keepsRegistry *registry.Keeps,
	subscriptionOnSignatureRequested subscription.EventSubscription,
) {
	keepTerminated := make(chan *eth.KeepTerminatedEvent)

	subscriptionOnKeepTerminated, err := ethereumChain.OnKeepTerminated(
		keepAddress,
		func(event *eth.KeepTerminatedEvent) {
			logger.Warningf(
				"keep [%s] terminated event received at block [%d]",
				keepAddress.String(),
				event.BlockNumber,
			)

			isKeepActive, err := waitForChainConfirmation(
				ethereumChain,
				event.BlockNumber,
				func() (bool, error) {
					return ethereumChain.IsActive(keepAddress)
				},
			)
			if err != nil {
				logger.Errorf(
					"failed to confirm keep [%s] termination: [%v]",
					keepAddress.String(),
					err,
				)
				return
			}

			if isKeepActive {
				logger.Warningf("keep [%s] has not been terminated", keepAddress.String())
				return
			}

			keepsRegistry.UnregisterKeep(keepAddress)
			keepTerminated <- event
		},
	)
	if err != nil {
		logger.Errorf(
			"failed on registering for keep terminated event: [%v]",
			err,
		)

		return
	}

	defer subscriptionOnKeepTerminated.Unsubscribe()
	defer subscriptionOnSignatureRequested.Unsubscribe()

	<-keepTerminated

	logger.Info("unsubscribing from events on keep terminated")
}

// waitForChainConfirmation ensures that after receiving specific number of block
// confirmations the state of the chain is actually as expected. It waits for
// predefined number of blocks since the start block number provided. After the
// required block number is reached it performs a check of the chain state with
// a provided function returning a boolean value.
func waitForChainConfirmation(
	ethereumChain eth.Handle,
	startBlockNumber uint64,
	stateCheck func() (bool, error),
) (bool, error) {
	blockHeight := startBlockNumber + blockConfirmations
	logger.Infof("waiting for block [%d] to confirm chain state", blockHeight)

	err := ethereumChain.BlockCounter().WaitForBlockHeight(blockHeight)
	if err != nil {
		return false, fmt.Errorf("failed to wait for block height: [%v]", err)
	}

	result, err := stateCheck()
	if err != nil {
		return false, fmt.Errorf("failed to get chain state confirmation: [%v]", err)
	}

	return result, nil
}
