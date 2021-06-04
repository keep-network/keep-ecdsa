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
	"github.com/keep-network/keep-ecdsa/pkg/chain/bitcoin"
	"github.com/keep-network/keep-ecdsa/pkg/client/event"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss"
	"github.com/keep-network/keep-ecdsa/pkg/extensions/tbtc"
	"github.com/keep-network/keep-ecdsa/pkg/extensions/tbtc/recovery"
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
	derivationIndexStorage *recovery.DerivationIndexStorage,
	clientConfig *Config,
	tbtcConfig *tbtc.Config,
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
				ctx,
				hostChain,
				tbtcApplicationHandle,
				networkProvider,
				clientConfig,
				tbtcConfig,
				tssNode,
				operatorPublicKey,
				keep,
				keepsRegistry,
				derivationIndexStorage,
				eventDeduplicator,
				subscriptionOnSignatureRequested,
			)

		}(keepID)
	}

	go checkAwaitingKeyGeneration(
		ctx,
		hostChain,
		tbtcApplicationHandle,
		networkProvider,
		clientConfig,
		tbtcConfig,
		tssNode,
		operatorPublicKey,
		keepsRegistry,
		derivationIndexStorage,
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
					tbtcApplicationHandle,
					networkProvider,
					clientConfig,
					tbtcConfig,
					tssNode,
					operatorPublicKey,
					keepsRegistry,
					derivationIndexStorage,
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
	hostChain chain.Handle,
	tbtcHandle chain.TBTCHandle,
	networkProvider net.Provider,
	clientConfig *Config,
	tbtcConfig *tbtc.Config,
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	keepsRegistry *registry.Keeps,
	derivationIndexStorage *recovery.DerivationIndexStorage,
	eventDeduplicator *event.Deduplicator,
) {
	keepCount, err := hostChain.GetKeepCount()
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

		keep, err := hostChain.GetKeepAtIndex(keepIndex)
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
			hostChain,
			tbtcHandle,
			networkProvider,
			clientConfig,
			tbtcConfig,
			tssNode,
			operatorPublicKey,
			keepsRegistry,
			derivationIndexStorage,
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
	hostChain chain.Handle,
	tbtcHandle chain.TBTCHandle,
	networkProvider net.Provider,
	clientConfig *Config,
	tbtcConfig *tbtc.Config,
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	keepsRegistry *registry.Keeps,
	derivationIndexStorage *recovery.DerivationIndexStorage,
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
			hostChain,
			tbtcHandle,
			networkProvider,
			clientConfig,
			tbtcConfig,
			tssNode,
			operatorPublicKey,
			keepsRegistry,
			derivationIndexStorage,
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
	hostChain chain.Handle,
	tbtcHandle chain.TBTCHandle,
	networkProvider net.Provider,
	clientConfig *Config,
	tbtcConfig *tbtc.Config,
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	keepsRegistry *registry.Keeps,
	derivationIndexStorage *recovery.DerivationIndexStorage,
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
		hostChain.OperatorID(),
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
		hostChain,
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
		hostChain,
		keep,
		keepsRegistry,
		subscriptionOnSignatureRequested,
		eventDeduplicator,
	)

	go monitorKeepTerminatedEvent(
		ctx,
		hostChain,
		tbtcHandle,
		networkProvider,
		clientConfig,
		tbtcConfig,
		tssNode,
		operatorPublicKey,
		keep,
		keepsRegistry,
		derivationIndexStorage,
		eventDeduplicator,
		subscriptionOnSignatureRequested,
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
	hostChain chain.Handle,
	clientConfig *Config,
	tssNode *node.Node,
	keep chain.BondedECDSAKeepHandle,
	signer *tss.ThresholdSigner,
	eventDeduplicator *event.Deduplicator,
) (subscription.EventSubscription, error) {
	go checkAwaitingSignature(
		hostChain,
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
							hostChain.BlockCounter(),
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
	hostChain chain.Handle,
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
					hostChain.BlockCounter(),
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
	hostChain chain.Handle,
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
					hostChain.BlockCounter(),
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

				// TODO: Rework how unregistering works in the context of
				// completing/confirming btc recovery on the bitcoin chain.
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
	ctx context.Context,
	hostChain chain.Handle,
	tbtcHandle chain.TBTCHandle,
	networkProvider net.Provider,
	clientConfig *Config,
	tbtcConfig *tbtc.Config,
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	keep chain.BondedECDSAKeepHandle,
	keepsRegistry *registry.Keeps,
	derivationIndexStorage *recovery.DerivationIndexStorage,
	eventDeduplicator *event.Deduplicator,
	subscriptionOnSignatureRequested subscription.EventSubscription,
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
				err := utils.DoWithDefaultRetry(
					tbtcConfig.GetLiquidationRecoveryTimeout(),
					func(ctx context.Context) error {
						if shouldHandle := eventDeduplicator.NotifyTerminatingStarted(keep.ID()); !shouldHandle {
							logger.Infof(
								"terminate event for keep [%s] already handled",
								keep.ID(),
							)

							// currently handling or already handled in the past
							// in case this event is a duplicate.
							return nil
						}
						defer eventDeduplicator.NotifyTerminatingCompleted(keep.ID())

						isKeepActive, err := ethlike.WaitForBlockConfirmations(
							hostChain.BlockCounter(),
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
							return err
						}

						if isKeepActive {
							logger.Warningf("keep [%s] has not been terminated", keep.ID())
							return err
						}

						logger.Infof(
							"starting liquidation recovery protocol for keep [%s]",
							keep.ID(),
						)

						members, err := keep.GetMembers()
						if err != nil {
							logger.Errorf(
								"failed to retrieve members from keep [%s]: [%v]",
								keep.ID(),
								err,
							)
							return err
						}
						memberID := tss.MemberIDFromPublicKey(operatorPublicKey)
						memberIDs, err := tssNode.AnnounceSignerPresence(
							ctx,
							operatorPublicKey,
							keep.ID(),
							members,
						)

						if err != nil {
							logger.Errorf(
								"failed to announce signer presence on keep [%s] termination: [%v]",
								keep.ID(),
								err,
							)
							return err
						}

						chainParams, err := tbtcConfig.Bitcoin.ChainParams()
						if err != nil {
							logger.Errorf(
								"failed to parse the the configured net params: [%v]",
								err,
							)
							return err
						}

						bitcoinHandle := bitcoin.Connect(tbtcConfig.Bitcoin.ElectrsURLWithDefault())

						beneficiaryAddress, err := recovery.ResolveAddress(
							tbtcConfig.Bitcoin.BeneficiaryAddress,
							derivationIndexStorage,
							chainParams,
							bitcoinHandle,
						)
						if err != nil {
							logger.Errorf(
								"failed to resolve a btc address for keep: [%s] address: [%s] err: [%v]",
								keep.ID(),
								tbtcConfig.Bitcoin.BeneficiaryAddress,
								err,
							)
							return err
						}

						vbyteFee, vbyteFeeError := bitcoinHandle.VbyteFeeFor25Blocks()
						if vbyteFeeError != nil {
							logger.Errorf(
								"failed to retrieve a vbyte fee estimate from %s, [%v]",
								tbtcConfig.Bitcoin.ElectrsURLWithDefault(),
								vbyteFeeError,
							)
							// Since the electrs connection is optional, we don't return the error
						}
						if vbyteFee == 0 {
							vbyteFee = tbtcConfig.Bitcoin.MaxFeePerVByte
						}
						if vbyteFee == 0 {
							vbyteFee = 75
						}

						btcAddresses, maxFeePerVByte, err := tss.BroadcastRecoveryAddress(
							ctx,
							beneficiaryAddress,
							vbyteFee,
							keep.ID().String(),
							memberID,
							memberIDs,
							uint(len(memberIDs)-1),
							networkProvider,
							hostChain.Signing().PublicKeyToAddress,
							chainParams,
						)
						if err != nil {
							logger.Errorf(
								"failed to communicate recovery details for keep [%s]: [%v]",
								keep.ID(),
								err,
							)
							return err
						}

						signer, err := keepsRegistry.GetSigner(keep.ID())
						if err != nil {
							// If there are no signer for loaded keep that something is clearly
							// wrong. We don't want to continue processing for this keep.
							logger.Errorf(
								"no signer for keep [%s]: [%v]",
								keep.ID(),
								err,
							)
							return err
						}

						logger.Infof(
							"building liquidation recovery transaction for keep [%s] "+
								"with receiving addresses [%v] and maxFeePerVByte [%d]",
							keep.ID(),
							btcAddresses,
							maxFeePerVByte,
						)

						recoveryTransactionHex, err := recovery.BuildBitcoinTransaction(
							ctx,
							networkProvider,
							hostChain,
							tbtcHandle,
							keep,
							signer,
							chainParams,
							btcAddresses,
							maxFeePerVByte,
						)
						if err != nil {
							logger.Errorf(
								"failed to build the transaction for keep [%s]: [%v]",
								keep.ID(),
								err,
							)
							return err
						}

						logger.Debugf(
							"broadcasting liquidation recovery transaction for keep [%s]: [%s]",
							keep.ID(),
							recoveryTransactionHex,
						)

						broadcastError := bitcoinHandle.Broadcast(recoveryTransactionHex)
						if broadcastError != nil {
							logger.Errorf(
								"failed to broadcast the recovery transaction to %s, [%v]",
								*tbtcConfig.Bitcoin.ElectrsURL,
								broadcastError,
							)

							for i := 0; i < 5; i++ {
								logger.Warningf("Please broadcast Bitcoin transaction %s", recoveryTransactionHex)
							}
						}

						logger.Debugf(
							"unregistering keep [%s] after liquidation recovery",
							keep.ID(),
						)

						keepsRegistry.UnregisterKeep(keep.ID())
						keepTerminated <- event

						return nil
					},
				)
				if err != nil {
					logger.Errorf("failed to broadcast the bitcoin recovery transaction: [%v]", err)
				}
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
