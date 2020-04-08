// Package client defines ECDSA keep client.
package client

import (
	"context"
	"fmt"
	"math/big"
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
				go monitorKeepTerminatedEvent(
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

	go checkAwaitingKeyGeneration(
		ctx,
		ethereumChain,
		tssNode,
		operatorPublicKey,
		keepsRegistry,
	)

	// Watch for new keeps creation.
	ethereumChain.OnBondedECDSAKeepCreated(func(event *eth.BondedECDSAKeepCreatedEvent) {
		logger.Infof(
			"new keep [%s] created with members: [%x]\n",
			event.KeepAddress.String(),
			event.Members,
		)

		if event.IsMember(ethereumChain.Address()) {
			go generateKeyForKeep(
				ctx,
				ethereumChain,
				tssNode,
				operatorPublicKey,
				keepsRegistry,
				event.KeepAddress,
				event.Members,
			)
		}
	})

	for _, application := range sanctionedApplications {
		go checkStatusAndRegisterForApplication(ctx, ethereumChain, application)
	}
}

func checkAwaitingKeyGeneration(
	ctx context.Context,
	ethereumChain eth.Handle,
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	keepsRegistry *registry.Keeps,
) {
	keepCount, err := ethereumChain.GetKeepCount()
	if err != nil {
		logger.Warningf("could not get keep count: [%v]", err)
		return
	}

	zero := big.NewInt(0)
	one := big.NewInt(1)

	lastIndex := new(big.Int).Sub(keepCount, one)

	// Iterate through keeps starting from the end.
	for keepIndex := new(big.Int).Set(lastIndex); keepIndex.Cmp(zero) != -1; keepIndex.Sub(keepIndex, one) {

		keep, err := ethereumChain.GetKeepAtIndex(keepIndex)
		if err != nil {
			logger.Warningf(
				"could not get keep at index [%v]: [%v]",
				keepIndex,
				err,
			)
			continue
		}

		keygenTimeout, err := ethereumChain.HasKeyGenerationTimedOut(keep)
		if err != nil {
			logger.Warningf(
				"could not check key generation timeout "+
					"for keep [%v]: [%v]",
				keep.String(),
				err,
			)
			continue
		}

		// If key generation timeout of current keep has been exceeded,
		// there is no sense to continue because the next keep (created earlier)
		// will have this timeout exceeded as well.
		if keygenTimeout {
			break
		}

		err = checkAwaitingKeyGenerationForKeep(
			ctx,
			ethereumChain,
			tssNode,
			operatorPublicKey,
			keepsRegistry,
			keep,
		)
		if err != nil {
			logger.Warningf(
				"could not check awaiting key generation "+
					"for keep [%s]: [%v]",
				keep.String(),
				err,
			)
		}
	}
}

func checkAwaitingKeyGenerationForKeep(
	ctx context.Context,
	ethereumChain eth.Handle,
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	keepsRegistry *registry.Keeps,
	keep common.Address,
) error {
	publicKey, err := ethereumChain.GetPublicKey(keep)
	if err != nil {
		return err
	}

	if len(publicKey) != 0 {
		return nil
	}

	members, err := ethereumChain.GetMembers(keep)
	if err != nil {
		return err
	}

	for _, member := range members {
		if ethereumChain.Address() == member {
			go generateKeyForKeep(
				ctx,
				ethereumChain,
				tssNode,
				operatorPublicKey,
				keepsRegistry,
				keep,
				members,
			)

			break
		}
	}

	return nil
}

func generateKeyForKeep(
	ctx context.Context,
	ethereumChain eth.Handle,
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	keepsRegistry *registry.Keeps,
	keepAddress common.Address,
	members []common.Address,
) {
	logger.Infof(
		"member [%s] is starting signer generation for keep [%s]...",
		ethereumChain.Address().String(),
		keepAddress.String(),
	)

	signer, err := generateSignerForKeep(
		ctx,
		tssNode,
		operatorPublicKey,
		keepAddress,
		members,
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
	}

	subscriptionOnSignatureRequested, err := monitorSigningRequests(
		ethereumChain,
		tssNode,
		keepAddress,
		signer,
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
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	keepAddress common.Address,
	members []common.Address,
) (*tss.ThresholdSigner, error) {
	keygenCtx, cancel := context.WithTimeout(ctx, keyGenerationTimeout)
	defer cancel()

	return tssNode.GenerateSignerForKeep(
		keygenCtx,
		operatorPublicKey,
		keepAddress,
		members,
	)
}

// monitorSigningRequests registers for signature requested events emitted by
// specific keep contract.
func monitorSigningRequests(
	ethereumChain eth.Handle,
	tssNode *node.Node,
	keepAddress common.Address,
	signer *tss.ThresholdSigner,
) (subscription.EventSubscription, error) {
	go checkAwaitingSignature(ethereumChain, tssNode, keepAddress, signer)

	return ethereumChain.OnSignatureRequested(
		keepAddress,
		func(event *eth.SignatureRequestedEvent) {
			logger.Infof(
				"new signature requested from keep [%s] for digest [%+x] at block [%d]",
				keepAddress.String(),
				event.Digest,
				event.BlockNumber,
			)

			isAwaitingSignature, err := waitForEventConfirmation(
				ethereumChain,
				event.BlockNumber,
				func() (bool, error) {
					return ethereumChain.IsAwaitingSignature(keepAddress, event.Digest)
				},
			)
			if err != nil {
				logger.Errorf(
					"failed to confirm signing request for digest [%+x] and keep [%s]",
					event.Digest,
					keepAddress.String(),
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

			go generateSignatureForKeep(tssNode, keepAddress, signer, event.Digest)
		},
	)
}

func checkAwaitingSignature(
	ethereumChain eth.Handle,
	tssNode *node.Node,
	keepAddress common.Address,
	signer *tss.ThresholdSigner,
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

			isKeepActive, err := waitForEventConfirmation(
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

			isKeepActive, err := waitForEventConfirmation(
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

// waitForEventConfirmation ensures that after receiving specific number of block
// confirmations the state of the chain is actually as expected. It waits for
// predefined number of blocks since the start block number provided. After the
// required block number is reached it performs a check of the chain state with
// a provided function returning a boolean value.
func waitForEventConfirmation(
	ethereumChain eth.Handle,
	startBlockNumber uint64,
	stateCheck func() (bool, error),
) (bool, error) {
	blockHeight := startBlockNumber + blockConfirmations
	logger.Infof("waiting for [%d] block", blockHeight)

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
