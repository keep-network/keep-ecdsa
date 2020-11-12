// Package client defines ECDSA keep client.
package client

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/keep-network/keep-common/pkg/chain/chainutil"

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

// The number of block confirmations the client waits until it starts the
// requested signing process. This value prevents from reporting unauthorized
// signings by adversaries in case of a chain fork.
const blockConfirmations = 12

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
	clientConfig *Config,
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

		isKeepActive, err := chainutil.WaitForBlockConfirmations(
			ethereumChain.BlockCounter(),
			currentBlock,
			blockConfirmations,
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
					"failed to verify if keep [%s] is still active: [%v]; "+
						"subscriptions for keep signing and closing events are skipped",
					keepAddress.String(),
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

			signer, err := keepsRegistry.GetSigner(keepAddress)
			if err != nil {
				// If there are no signer for loaded keep that something is clearly
				// wrong. We don't want to continue processing for this keep.
				logger.Errorf(
					"no signer for keep [%s]: [%v]",
					keepAddress.String(),
					err,
				)
				return
			}

			subscriptionOnSignatureRequested, err := monitorSigningRequests(
				ethereumChain,
				clientConfig,
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

		}(keepAddress)
	}

	go checkAwaitingKeyGeneration(
		ctx,
		ethereumChain,
		clientConfig,
		tssNode,
		operatorPublicKey,
		keepsRegistry,
		requestedSignatures,
	)

	// Watch for new keeps creation.
	_ = ethereumChain.OnBondedECDSAKeepCreated(func(event *eth.BondedECDSAKeepCreatedEvent) {
		logger.Infof(
			"new keep [%s] created with members: [%x]\n",
			event.KeepAddress.String(),
			event.Members,
		)

		if event.IsMember(ethereumChain.Address()) {
			go func(event *eth.BondedECDSAKeepCreatedEvent) {
				if ok := requestedSigners.add(event.KeepAddress); !ok {
					logger.Warningf(
						"keep creation event for keep [%s] already registered",
						event.KeepAddress.String(),
					)
					return
				}
				defer requestedSigners.remove(event.KeepAddress)

				if keepsRegistry.HasSigner(event.KeepAddress) {
					logger.Warning(
						"signer for keep [%s] already registered;",
						event.KeepAddress.String(),
					)
					return
				}

				generateKeyForKeep(
					ctx,
					ethereumChain,
					clientConfig,
					tssNode,
					operatorPublicKey,
					keepsRegistry,
					requestedSignatures,
					event.KeepAddress,
					event.Members,
					event.HonestThreshold,
				)
			}(event)
		}
	})

	for _, application := range sanctionedApplications {
		go checkStatusAndRegisterForApplication(ctx, ethereumChain, application)
	}
}

func checkAwaitingKeyGeneration(
	ctx context.Context,
	ethereumChain eth.Handle,
	clientConfig *Config,
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	keepsRegistry *registry.Keeps,
	requestedSignatures *requestedSignaturesTrack,
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

		keepOpenedTimestamp, err := ethereumChain.GetOpenedTimestamp(keep)
		if err != nil {
			logger.Warningf(
				"could not check opening timestamp for keep [%s]: [%v]",
				keep.String(),
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
			requestedSignatures,
			keep,
		)
		if err != nil {
			logger.Warningf(
				"could not check awaiting key generation for keep [%s]: [%v]",
				keep.String(),
				err,
			)
		}
	}
}

func checkAwaitingKeyGenerationForKeep(
	ctx context.Context,
	ethereumChain eth.Handle,
	clientConfig *Config,
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	keepsRegistry *registry.Keeps,
	requestedSignatures *requestedSignaturesTrack,
	keep common.Address,
) error {
	publicKey, err := ethereumChain.GetPublicKey(keep)
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
	if keepsRegistry.HasSigner(keep) {
		logger.Warningf(
			"keep public key is not registered on-chain but key material "+
				"is stored on disk; skipping key generation; PLEASE INSPECT "+
				"PUBLIC KEY SUBMISSION TRANSACTION FOR KEEP [%v]",
			keep.String(),
		)
		return nil
	}

	members, err := ethereumChain.GetMembers(keep)
	if err != nil {
		return err
	}

	honestThreshold, err := ethereumChain.GetHonestThreshold(keep)
	if err != nil {
		return err
	}

	for _, member := range members {
		if ethereumChain.Address() == member {
			go generateKeyForKeep(
				ctx,
				ethereumChain,
				clientConfig,
				tssNode,
				operatorPublicKey,
				keepsRegistry,
				requestedSignatures,
				keep,
				members,
				honestThreshold,
			)

			break
		}
	}

	return nil
}

func generateKeyForKeep(
	ctx context.Context,
	ethereumChain eth.Handle,
	clientConfig *Config,
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	keepsRegistry *registry.Keeps,
	requestedSignatures *requestedSignaturesTrack,
	keepAddress common.Address,
	members []common.Address,
	honestThreshold uint64,
) {
	if len(members) < 2 {
		// TODO: #408 Implement single signer support.
		logger.Errorf(
			"keep [%s] has [%d] members; only keeps with at least 2 members are supported",
			keepAddress.String(),
			len(members),
		)
		return
	}

	if honestThreshold != uint64(len(members)) {
		// TODO: #325 Implement threshold support.
		logger.Errorf(
			"keep [%s] has honest threshold [%s] and [%d] members; "+
				"only keeps with honest threshold same as group size are supported",
			keepAddress.String(),
			honestThreshold,
			len(members),
		)
		return
	}

	logger.Infof(
		"member [%s] is starting signer generation for keep [%s]...",
		ethereumChain.Address().String(),
		keepAddress.String(),
	)

	signer, err := generateSignerForKeep(
		ctx,
		clientConfig,
		tssNode,
		operatorPublicKey,
		keepAddress,
		members,
		keepsRegistry,
	)
	if err != nil {
		logger.Errorf(
			"failed to generate signer for keep [%s]: [%v]",
			keepAddress.String(),
			err,
		)
		return
	}

	logger.Infof("initialized signer for keep [%s]", keepAddress.String())

	err = keepsRegistry.RegisterSigner(keepAddress, signer)
	if err != nil {
		logger.Errorf(
			"failed to register threshold signer for keep [%s]: [%v]",
			keepAddress.String(),
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
		keepAddress,
		signer,
		requestedSignatures,
	)
	if err != nil {
		logger.Errorf(
			"failed on registering for requested signature event "+
				"for keep [%s]: [%v]",
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

func generateSignerForKeep(
	ctx context.Context,
	clientConfig *Config,
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	keepAddress common.Address,
	members []common.Address,
	keepsRegistry *registry.Keeps,
) (*tss.ThresholdSigner, error) {
	keygenCtx, cancel := context.WithTimeout(ctx, clientConfig.GetKeyGenerationTimeout())
	defer cancel()

	return tssNode.GenerateSignerForKeep(
		keygenCtx,
		operatorPublicKey,
		keepAddress,
		members,
		keepsRegistry,
	)
}

// monitorSigningRequests registers for signature requested events emitted by
// specific keep contract.
func monitorSigningRequests(
	ethereumChain eth.Handle,
	clientConfig *Config,
	tssNode *node.Node,
	keepAddress common.Address,
	signer *tss.ThresholdSigner,
	requestedSignatures *requestedSignaturesTrack,
) (subscription.EventSubscription, error) {
	go checkAwaitingSignature(
		ethereumChain,
		clientConfig,
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
					logger.Warningf(
						"signature requested event for keep [%s] and digest [%x] already registered",
						keepAddress.String(),
						event.Digest,
					)
					return
				}
				defer requestedSignatures.remove(keepAddress, event.Digest)

				isAwaitingSignature, err := chainutil.WaitForBlockConfirmations(
					ethereumChain.BlockCounter(),
					event.BlockNumber,
					blockConfirmations,
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

				generateSignatureForKeep(clientConfig, tssNode, keepAddress, signer, event.Digest)
			}(event)
		},
	)
}

func checkAwaitingSignature(
	ethereumChain eth.Handle,
	clientConfig *Config,
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
		logger.Infof(
			"awaiting a signature from keep [%s] for digest [%+x]",
			keepAddress.String(),
			latestDigest,
		)

		if ok := requestedSignatures.add(keepAddress, latestDigest); !ok {
			logger.Warningf(
				"signature requested event for keep [%s] and digest [%x] already registered",
				keepAddress.String(),
				latestDigest,
			)
			return
		}
		defer requestedSignatures.remove(keepAddress, latestDigest)

		startBlock, err := ethereumChain.SignatureRequestedBlock(keepAddress, latestDigest)
		if err != nil {
			logger.Errorf(
				"failed to get signature request block height for keep [%s] and digest [%x]: [%v]",
				keepAddress.String(),
				latestDigest,
				err,
			)
			return
		}

		isStillAwaitingSignature, err := chainutil.WaitForBlockConfirmations(
			ethereumChain.BlockCounter(),
			startBlock,
			blockConfirmations,
			func() (bool, error) {
				isAwaitingSignature, err := ethereumChain.IsAwaitingSignature(keepAddress, latestDigest)
				if err != nil {
					return false, err
				}

				isActive, err := ethereumChain.IsActive(keepAddress)
				if err != nil {
					return false, err
				}

				return (isAwaitingSignature && isActive), nil
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

		generateSignatureForKeep(clientConfig, tssNode, keepAddress, signer, latestDigest)
	}
}

func generateSignatureForKeep(
	clientConfig *Config,
	tssNode *node.Node,
	keepAddress common.Address,
	signer *tss.ThresholdSigner,
	digest [32]byte,
) {
	signingCtx, cancel := context.WithTimeout(
		context.Background(),
		clientConfig.GetSigningTimeout(),
	)
	defer cancel()

	if err := tssNode.CalculateSignature(
		signingCtx,
		signer,
		digest,
	); err != nil {
		logger.Errorf(
			"signature calculation failed for keep [%s]: [%v]",
			keepAddress.String(),
			err,
		)
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

			isKeepActive, err := chainutil.WaitForBlockConfirmations(
				ethereumChain.BlockCounter(),
				event.BlockNumber,
				blockConfirmations,
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
			"failed on registering for closed event for keep [%s]: [%v]",
			keepAddress.String(),
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

			isKeepActive, err := chainutil.WaitForBlockConfirmations(
				ethereumChain.BlockCounter(),
				event.BlockNumber,
				blockConfirmations,
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
			"failed on registering for terminated event for keep [%s]: [%v]",
			keepAddress.String(),
			err,
		)

		return
	}

	defer subscriptionOnKeepTerminated.Unsubscribe()
	defer subscriptionOnSignatureRequested.Unsubscribe()

	<-keepTerminated

	logger.Info("unsubscribing from events on keep terminated")
}
