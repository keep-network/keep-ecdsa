// Package client defines ECDSA keep client.
package client

import (
	"context"
	"math/big"
	"time"

	"github.com/keep-network/keep-common/pkg/chain/ethlike"

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-common/pkg/subscription"
	corechain "github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-core/pkg/net"
	"github.com/keep-network/keep-core/pkg/operator"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/client/event"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss"
	"github.com/keep-network/keep-ecdsa/pkg/extensions/tbtc"
	"github.com/keep-network/keep-ecdsa/pkg/node"
	"github.com/keep-network/keep-ecdsa/pkg/registry"
	"github.com/keep-network/keep-ecdsa/pkg/utils"
)

var logger = log.Logger("keep-ecdsa")

// The number of block confirmations the client waits until it starts the
// requested signing process. This value prevents from reporting unauthorized
// signings by adversaries in case of a chain fork.
const blockConfirmations = 12

// The timeout for executing repeated on-chain check for a keep awaiting
// a signature. Once the client receives a signature requested event, it needs
// to deduplicate it and execute on-chain check. This action is repeated with
// a timeout to address problems with minor chain re-orgs and chain clients not
// being perfectly in sync yet.
const awaitingSignatureEventCheckTimeout = 60 * time.Second

// Handle represents a handle to the ECDSA client.
type Handle struct {
	tssNode *node.Node
}

// TSSPreParamsPoolSize returns the current size of the TSS params pool.
func (h *Handle) TSSPreParamsPoolSize() int {
	return h.tssNode.TSSPreParamsPoolSize()
}

// Initialize initializes the ECDSA client with rules related to events handling.
// Expects a slice of sanctioned applications selected by the operator for which
// operator will be registered as a member candidate.
func Initialize(
	ctx context.Context,
	operatorPublicKey *operator.PublicKey,
	hostChain chain.Handle,
	networkProvider net.Provider,
	persistence persistence.Handle,
	clientConfig *Config,
	tssConfig *tss.Config,
) *Handle {
	keepsRegistry := registry.NewKeepsRegistry(
		persistence,
		hostChain.UnmarshalID,
	)

	tssNode := node.NewNode(hostChain, networkProvider, tssConfig)

	tssNode.InitializeTSSPreParamsPool()

	eventDeduplicator := event.NewDeduplicator(
		keepsRegistry,
		hostChain,
	)

	// Load current keeps' signers from storage and register for signing events.
	keepsRegistry.LoadExistingKeeps()

	confirmIsInactive := func(keep chain.BondedECDSAKeepHandle) bool {
		currentBlock, err := hostChain.BlockCounter().CurrentBlock()
		if err != nil {
			logger.Errorf("failed to get current block height [%v]", err)
			return false
		}

		isKeepActive, err := ethlike.WaitForBlockConfirmations(
			hostChain.BlockCounter(),
			currentBlock,
			blockConfirmations,
			keep.IsActive,
		)
		if err != nil {
			logger.Errorf(
				"failed to confirm that keep [%s] is inactive: [%v]",
				keep.ID(),
				err,
			)
			return false
		}

		return !isKeepActive
	}

	for _, keepID := range keepsRegistry.GetKeepsIDs() {
		go func(keepID chain.ID) {
			keep, err := hostChain.GetKeepWithID(keepID)
			if err != nil {
				logger.Errorf(
					"failed to look up keep [%s] for active check: [%v]; "+
						"subscriptions for keep signing and closing events are skipped",
					keep.ID(),
					err,
				)
				return
			}

			isActive, err := keep.IsActive()
			if err != nil {
				logger.Errorf(
					"failed to verify if keep [%s] is still active: [%v]; "+
						"subscriptions for keep signing and closing events are skipped",
					keep.ID(),
					err,
				)
				return
			}

			if !isActive {
				logger.Infof(
					"keep [%s] seems no longer active; confirming",
					keep.ID(),
				)
				if isInactivityConfirmed := confirmIsInactive(keep); isInactivityConfirmed {
					logger.Infof(
						"confirmed that keep [%s] is no longer active; archiving",
						keep.ID(),
					)
					keepsRegistry.UnregisterKeep(keepID)
					return
				}
				logger.Warningf("keep [%s] is still active", keep.ID())
			}

			signer, err := keepsRegistry.GetSigner(keepID)
			if err != nil {
				// If there are no signer for loaded keep that something is clearly
				// wrong. We don't want to continue processing for this keep.
				logger.Errorf(
					"no signer for keep [%s]: [%v]",
					keep.ID(),
					err,
				)
				return
			}

			subscriptionOnSignatureRequested, err := monitorSigningRequests(
				hostChain,
				clientConfig,
				tssNode,
				keep,
				signer,
				eventDeduplicator,
			)
			if err != nil {
				logger.Errorf(
					"failed registering for requested signature event for keep [%s]: [%v]",
					keep.ID(),
					err,
				)
				// In case of an error we want to avoid subscribing to keep
				// closed events. Something is wrong and we should stop
				// further processing.
				return
			}
			go monitorKeepClosedEvents(
				hostChain,
				keep,
				keepsRegistry,
				subscriptionOnSignatureRequested,
				eventDeduplicator,
			)
			go monitorKeepTerminatedEvent(
				hostChain,
				keep,
				keepsRegistry,
				subscriptionOnSignatureRequested,
				eventDeduplicator,
			)

		}(keepID)
	}

	go checkAwaitingKeyGeneration(
		ctx,
		hostChain,
		clientConfig,
		tssNode,
		operatorPublicKey,
		keepsRegistry,
		eventDeduplicator,
	)

	// Watch for new keeps creation.
	_ = hostChain.OnBondedECDSAKeepCreated(func(event *chain.BondedECDSAKeepCreatedEvent) {
		logger.Infof(
			"new keep [%s] created with members: [%s] at block [%d]",
			event.Keep.ID(),
			event.MemberIDs,
			event.BlockNumber,
		)

		if event.ThisOperatorIsMember {
			go func(event *chain.BondedECDSAKeepCreatedEvent) {
				if shouldHandle := eventDeduplicator.NotifyKeyGenStarted(event.Keep.ID()); !shouldHandle {
					logger.Infof(
						"key generation request for keep [%s] already handled",
						event.Keep.ID(),
					)

					// currently handling or already handled in the past
					// in case this event is a duplicate.
					return
				}
				defer eventDeduplicator.NotifyKeyGenCompleted(event.Keep.ID())

				keep, err := hostChain.GetKeepWithID(event.Keep.ID())
				if err != nil {
					logger.Errorf(
						"failed to resolve keep with address [%v] for created event: [%v]",
						keep.ID(),
						err,
					)
				}

				generateKeyForKeep(
					ctx,
					hostChain,
					clientConfig,
					tssNode,
					operatorPublicKey,
					keepsRegistry,
					eventDeduplicator,
					keep,
					event.MemberIDs,
					event.HonestThreshold,
				)
			}(event)
		} else {
			logger.Infof(
				"not a signing group member in keep [%s], skipping",
				event.Keep.ID(),
			)
		}
	})

	blockCounter := hostChain.BlockCounter()

	tbtcApplicationHandle, err := hostChain.TBTCApplicationHandle()
	if err != nil {
		logger.Errorf(
			"failed to look up on-chain tBTC application information for "+
				"chain [%s]; this client WILL NOT ATTEMPT TO OPERATE "+
				"on the tBTC system",
			hostChain,
		)
	} else {
		go checkStatusAndRegisterForApplication(ctx, blockCounter, tbtcApplicationHandle)
	}

	initializeExtensions(
		ctx,
		tbtcApplicationHandle,
		blockCounter,
		hostChain.BlockTimestamp,
	)

	return &Handle{
		tssNode: tssNode,
	}
}

func initializeExtensions(
	ctx context.Context,
	tbtcHandle chain.TBTCHandle,
	blockCounter corechain.BlockCounter,
	blockTimestamp func(blockNumber *big.Int) (uint64, error),
) {
	if tbtcHandle != nil {
		tbtc.Initialize(
			ctx,
			tbtcHandle,
			blockCounter,
			blockTimestamp,
		)
	} else {
		logger.Errorf(
			"could not initialize tbtc chain extension",
		)
	}
}

func checkAwaitingKeyGeneration(
	ctx context.Context,
	ethereumChain chain.Handle,
	clientConfig *Config,
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	keepsRegistry *registry.Keeps,
	eventDeduplicator *event.Deduplicator,
) {
	keepCount, err := ethereumChain.GetKeepCount()
	if err != nil {
		logger.Warningf("could not get keep count: [%v]", err)
		return
	}

	lookbackPeriod := clientConfig.GetAwaitingKeyGenerationLookback()

	zero := big.NewInt(0)
	one := big.NewInt(1)

	lastIndex := new(big.Int).Sub(keepCount, one)

	// Iterate through keeps starting from the end.
	for keepIndex := new(big.Int).Set(lastIndex); keepIndex.Cmp(zero) != -1; keepIndex.Sub(keepIndex, one) {
		logger.Debugf(
			"checking awaiting key generation for keep at index [%v]",
			keepIndex.String(),
		)

		keep, err := ethereumChain.GetKeepAtIndex(keepIndex)
		if err != nil {
			logger.Warningf(
				"could not get keep at index [%v]: [%v]",
				keepIndex,
				err,
			)
			continue
		}

		keepOpenedTimestamp, err := keep.GetOpenedTimestamp()
		if err != nil {
			logger.Warningf(
				"could not check opening timestamp for keep [%s]: [%v]",
				keep.ID(),
				err,
			)
			continue
		}

		// If a keep was opened before the defined lookback duration there is no
		// sense to continue because the next keep was created earlier.
		if keepOpenedTimestamp.Add(lookbackPeriod).Before(time.Now()) {
			logger.Debugf(
				"stopping awaiting key generation check with keep at index [%s] opened at [%s]",
				keepIndex,
				keepOpenedTimestamp,
			)
			break
		}

		err = checkAwaitingKeyGenerationForKeep(
			ctx,
			ethereumChain,
			clientConfig,
			tssNode,
			operatorPublicKey,
			keepsRegistry,
			eventDeduplicator,
			keep,
		)
		if err != nil {
			logger.Warningf(
				"could not check awaiting key generation for keep [%s]: [%v]",
				keep.ID(),
				err,
			)
		}
	}
}

func checkAwaitingKeyGenerationForKeep(
	ctx context.Context,
	ethereumChain chain.Handle,
	clientConfig *Config,
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	keepsRegistry *registry.Keeps,
	eventDeduplicator *event.Deduplicator,
	keep chain.BondedECDSAKeepHandle,
) error {
	publicKey, err := keep.GetPublicKey()
	if err != nil {
		return err
	}

	if len(publicKey) != 0 {
		return nil
	}

	// If the key material is stored in the registry it means that the key
	// generation succeeded and public key transaction has been submitted.
	// There are two scenarios possible:
	// - public key submission transactions are still mining,
	// - conflicting public key has been submitted.
	// In both cases, the client should not attempt to generate the key again.
	if keepsRegistry.HasSigner(keep.ID()) {
		logger.Warningf(
			"keep public key is not registered on-chain but key material "+
				"is stored on disk; skipping key generation; PLEASE INSPECT "+
				"PUBLIC KEY SUBMISSION TRANSACTION FOR KEEP [%v]",
			keep.ID(),
		)
		return nil
	}

	members, err := keep.GetMembers()
	if err != nil {
		return err
	}

	isThisOperatorMember, err := keep.IsThisOperatorMember()
	if err != nil {
		return err
	}

	honestThreshold, err := keep.GetHonestThreshold()
	if err != nil {
		return err
	}

	if isThisOperatorMember {
		go generateKeyForKeep(
			ctx,
			ethereumChain,
			clientConfig,
			tssNode,
			operatorPublicKey,
			keepsRegistry,
			eventDeduplicator,
			keep,
			members,
			honestThreshold,
		)
	}

	return nil
}

func generateKeyForKeep(
	ctx context.Context,
	ethereumChain chain.Handle,
	clientConfig *Config,
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	keepsRegistry *registry.Keeps,
	eventDeduplicator *event.Deduplicator,
	keep chain.BondedECDSAKeepHandle,
	members []chain.ID,
	honestThreshold uint64,
) {
	if len(members) < 2 {
		// TODO: #408 Implement single signer support.
		logger.Errorf(
			"keep [%s] has [%d] members; only keeps with at least 2 members are supported",
			keep.ID(),
			len(members),
		)
		return
	}

	if honestThreshold != uint64(len(members)) {
		// TODO: #325 Implement threshold support.
		logger.Errorf(
			"keep [%s] has honest threshold [%s] and [%d] members; "+
				"only keeps with honest threshold same as group size are supported",
			keep.ID(),
			honestThreshold,
			len(members),
		)
		return
	}

	logger.Infof(
		"member [%s] is starting signer generation for keep [%s]...",
		ethereumChain.OperatorID(),
		keep.ID(),
	)

	signer, err := generateSignerForKeep(
		ctx,
		clientConfig,
		tssNode,
		operatorPublicKey,
		keep,
		members,
		keepsRegistry,
	)
	if err != nil {
		logger.Errorf(
			"failed to generate signer for keep [%s]: [%v]",
			keep.ID(),
			err,
		)
		return
	}

	logger.Infof("initialized signer for keep [%s]", keep.ID())

	err = keepsRegistry.RegisterSigner(keep.ID(), signer)
	if err != nil {
		logger.Errorf(
			"failed to register threshold signer for keep [%s]: [%v]",
			keep.ID(),
			err,
		)

		// In case of an error during signer registration, we want to avoid
		// subscribing to the events emitted by the keep. The signer is not
		// operating so we should stop further processing.
		return
	}

	subscriptionOnSignatureRequested, err := monitorSigningRequests(
		ethereumChain,
		clientConfig,
		tssNode,
		keep,
		signer,
		eventDeduplicator,
	)
	if err != nil {
		logger.Errorf(
			"failed on registering for requested signature event "+
				"for keep [%s]: [%v]",
			keep.ID(),
			err,
		)

		// In case of an error we want to avoid subscribing to keep
		// closed events. Something is wrong and we should stop
		// further processing.
		return
	}

	go monitorKeepClosedEvents(
		ethereumChain,
		keep,
		keepsRegistry,
		subscriptionOnSignatureRequested,
		eventDeduplicator,
	)
	go monitorKeepTerminatedEvent(
		ethereumChain,
		keep,
		keepsRegistry,
		subscriptionOnSignatureRequested,
		eventDeduplicator,
	)
}

func generateSignerForKeep(
	ctx context.Context,
	clientConfig *Config,
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	keep chain.BondedECDSAKeepHandle,
	members []chain.ID,
	keepsRegistry *registry.Keeps,
) (*tss.ThresholdSigner, error) {
	keygenCtx, cancel := context.WithTimeout(ctx, clientConfig.GetKeyGenerationTimeout())
	defer cancel()

	return tssNode.GenerateSignerForKeep(
		keygenCtx,
		operatorPublicKey,
		keep,
		members,
		keepsRegistry,
	)
}

// monitorSigningRequests registers for signature requested events emitted by
// specific keep contract.
func monitorSigningRequests(
	ethereumChain chain.Handle,
	clientConfig *Config,
	tssNode *node.Node,
	keep chain.BondedECDSAKeepHandle,
	signer *tss.ThresholdSigner,
	eventDeduplicator *event.Deduplicator,
) (subscription.EventSubscription, error) {
	go checkAwaitingSignature(
		ethereumChain,
		clientConfig,
		tssNode,
		keep,
		signer,
		eventDeduplicator,
	)

	return keep.OnSignatureRequested(
		func(event *chain.SignatureRequestedEvent) {
			logger.Infof(
				"new signature requested from keep [%s] for digest [%+x] at block [%d]",
				keep.ID(),
				event.Digest,
				event.BlockNumber,
			)

			go func(event *chain.SignatureRequestedEvent) {
				err := utils.DoWithDefaultRetry(
					clientConfig.GetSigningTimeout(),
					// TODO: extract the code into a separate function and see if
					// there is a way to deduplicate common parts with
					// checkAwaitingSignature function.
					func(ctx context.Context) error {
						shouldHandle, err := eventDeduplicator.NotifySigningStarted(
							awaitingSignatureEventCheckTimeout,
							keep,
							event.Digest,
						)
						if err != nil {
							logger.Errorf(
								"could not deduplicate signing request event: [%v]",
								err,
							)
							return err
						}

						if !shouldHandle {
							logger.Infof(
								"signing request for keep [%s] and digest [%+x] already handled",
								keep.ID(),
								event.Digest,
							)
							// currently handling or already handled in the past
							// in case this event is a duplicate.
							return nil
						}

						defer eventDeduplicator.NotifySigningCompleted(keep.ID(), event.Digest)

						isAwaitingSignature, err := ethlike.WaitForBlockConfirmations(
							ethereumChain.BlockCounter(),
							event.BlockNumber,
							blockConfirmations,
							func() (bool, error) {
								return keep.IsAwaitingSignature(event.Digest)
							},
						)
						if err != nil {
							logger.Errorf(
								"failed to confirm signing request for keep [%s] and digest [%+x]: [%v]",
								keep.ID(),
								event.Digest,
								err,
							)
							return err
						}

						if !isAwaitingSignature {
							logger.Warningf(
								"keep [%s] is not awaiting a signature for digest [%+x]",
								keep.ID(),
								event.Digest,
							)

							// deeper chain reorg, nothing we should do
							return nil
						}

						if err := tssNode.CalculateSignature(
							ctx,
							keep,
							signer,
							event.Digest,
						); err != nil {
							logger.Errorf(
								"signature calculation failed for keep [%s]: [%v]",
								keep.ID(),
								err,
							)
						}

						return err
					},
				)
				if err != nil {
					logger.Errorf("failed to generate a signature: [%v]", err)
				}
			}(event)
		},
	)
}

func checkAwaitingSignature(
	ethereumChain chain.Handle,
	clientConfig *Config,
	tssNode *node.Node,
	keep chain.BondedECDSAKeepHandle,
	signer *tss.ThresholdSigner,
	eventDeduplicator *event.Deduplicator,
) {
	logger.Debugf("checking awaiting signature for keep [%s]", keep.ID())

	latestDigest, err := keep.LatestDigest()
	if err != nil {
		logger.Errorf("could not get latest digest for keep [%s]", keep.ID())
		return
	}

	isAwaitingDigest, err := keep.IsAwaitingSignature(latestDigest)
	if err != nil {
		logger.Errorf(
			"could not check awaiting signature of "+
				"digest [%+x] for keep [%s]",
			latestDigest,
			keep.ID(),
		)
		return
	}

	if isAwaitingDigest {
		logger.Infof(
			"awaiting a signature from keep [%s] for digest [%+x]",
			keep.ID(),
			latestDigest,
		)

		err := utils.DoWithDefaultRetry(
			clientConfig.GetSigningTimeout(),
			func(ctx context.Context) error {
				shouldHandle, err := eventDeduplicator.NotifySigningStarted(
					awaitingSignatureEventCheckTimeout,
					keep,
					latestDigest,
				)
				if err != nil {
					logger.Errorf(
						"could not deduplicate signing request event: [%v]",
						err,
					)
					return err
				}

				if !shouldHandle {
					logger.Infof(
						"signing request for keep [%s] and digest [%+x] already handled",
						keep.ID(),
						latestDigest,
					)
					// currently handling - it is possible that event
					// subscription also received this event
					return nil
				}

				defer eventDeduplicator.NotifySigningCompleted(keep.ID(), latestDigest)

				startBlock, err := keep.SignatureRequestedBlock(latestDigest)
				if err != nil {
					logger.Errorf(
						"failed to get signature request block height for keep [%s] and digest [%x]: [%v]",
						keep.ID(),
						latestDigest,
						err,
					)
					return err
				}

				isStillAwaitingSignature, err := ethlike.WaitForBlockConfirmations(
					ethereumChain.BlockCounter(),
					startBlock,
					blockConfirmations,
					func() (bool, error) {
						isAwaitingSignature, err := keep.IsAwaitingSignature(latestDigest)
						if err != nil {
							return false, err
						}

						isActive, err := keep.IsActive()
						if err != nil {
							return false, err
						}

						return (isAwaitingSignature && isActive), nil
					},
				)
				if err != nil {
					logger.Errorf(
						"failed to confirm signing request for keep [%s] and digest [%+x]: [%v]",
						keep.ID(),
						latestDigest,
						err,
					)
					return err
				}

				if !isStillAwaitingSignature {
					logger.Warningf(
						"keep [%s] is not awaiting a signature for digest [%+x]",
						keep.ID(),
						latestDigest,
					)

					// deeper chain reorg, nothing we should do
					return nil
				}

				if err := tssNode.CalculateSignature(
					ctx,
					keep,
					signer,
					latestDigest,
				); err != nil {
					logger.Errorf(
						"signature calculation failed for keep [%s]: [%v]",
						keep.ID(),
						err,
					)
				}

				return err
			},
		)
		if err != nil {
			logger.Errorf("failed to generate a signature: [%v]", err)
		}
	}
}

// monitorKeepClosedEvent monitors KeepClosed event and if that event happens
// unsubscribes from signing event for the given keep and unregisters it from
// the keep registry.
func monitorKeepClosedEvents(
	ethereumChain chain.Handle,
	keep chain.BondedECDSAKeepHandle,
	keepsRegistry *registry.Keeps,
	subscriptionOnSignatureRequested subscription.EventSubscription,
	eventDeduplicator *event.Deduplicator,
) {
	keepClosed := make(chan *chain.KeepClosedEvent)

	subscriptionOnKeepClosed, err := keep.OnKeepClosed(
		func(event *chain.KeepClosedEvent) {
			logger.Infof(
				"keep [%s] closed event received at block [%d]",
				keep.ID(),
				event.BlockNumber,
			)

			go func(event *chain.KeepClosedEvent) {
				if shouldHandle := eventDeduplicator.NotifyClosingStarted(keep.ID()); !shouldHandle {
					logger.Infof(
						"close event for keep [%s] already handled",
						keep.ID(),
					)

					// currently handling or already handled in the past
					// in case this event is a duplicate.
					return
				}
				defer eventDeduplicator.NotifyClosingCompleted(keep.ID())

				isKeepActive, err := ethlike.WaitForBlockConfirmations(
					ethereumChain.BlockCounter(),
					event.BlockNumber,
					blockConfirmations,
					func() (bool, error) {
						return keep.IsActive()
					},
				)
				if err != nil {
					logger.Errorf(
						"failed to confirm keep [%s] closed: [%v]",
						keep.ID(),
						err,
					)
					return
				}

				if isKeepActive {
					logger.Warningf("keep [%s] has not been closed", keep.ID())
					return
				}

				keepsRegistry.UnregisterKeep(keep.ID())
				keepClosed <- event
			}(event)
		},
	)
	if err != nil {
		logger.Errorf(
			"failed on registering for closed event for keep [%s]: [%v]",
			keep.ID(),
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
	ethereumChain chain.Handle,
	keep chain.BondedECDSAKeepHandle,
	keepsRegistry *registry.Keeps,
	subscriptionOnSignatureRequested subscription.EventSubscription,
	eventDeduplicator *event.Deduplicator,
) {
	keepTerminated := make(chan *chain.KeepTerminatedEvent)

	subscriptionOnKeepTerminated, err := keep.OnKeepTerminated(
		func(event *chain.KeepTerminatedEvent) {
			logger.Warningf(
				"keep [%s] terminated event received at block [%d]",
				keep.ID(),
				event.BlockNumber,
			)

			go func(event *chain.KeepTerminatedEvent) {
				if shouldHandle := eventDeduplicator.NotifyTerminatingStarted(keep.ID()); !shouldHandle {
					logger.Infof(
						"terminate event for keep [%s] already handled",
						keep.ID(),
					)

					// currently handling or already handled in the past
					// in case this event is a duplicate.
					return
				}
				defer eventDeduplicator.NotifyTerminatingCompleted(keep.ID())

				isKeepActive, err := ethlike.WaitForBlockConfirmations(
					ethereumChain.BlockCounter(),
					event.BlockNumber,
					blockConfirmations,
					func() (bool, error) {
						return keep.IsActive()
					},
				)
				if err != nil {
					logger.Errorf(
						"failed to confirm keep [%s] termination: [%v]",
						keep.ID(),
						err,
					)
					return
				}

				if isKeepActive {
					logger.Warningf("keep [%s] has not been terminated", keep.ID())
					return
				}

				keepsRegistry.UnregisterKeep(keep.ID())
				keepTerminated <- event
			}(event)
		},
	)
	if err != nil {
		logger.Errorf(
			"failed on registering for terminated event for keep [%s]: [%v]",
			keep.ID(),
			err,
		)

		return
	}

	defer subscriptionOnKeepTerminated.Unsubscribe()
	defer subscriptionOnSignatureRequested.Unsubscribe()

	<-keepTerminated

	logger.Info("unsubscribing from events on keep terminated")
}
