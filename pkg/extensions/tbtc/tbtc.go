package tbtc

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-common/pkg/subscription"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
)

var logger = log.Logger("extensions-tbtc")

const maxTransactionAttempts = 3

// InitializeExtensions initializes extensions specific to the TBTC application.
func InitializeExtensions(ctx context.Context, handle Handle) error {
	logger.Infof("initializing tbtc extensions")

	manager := &extensionsManager{handle}

	err := manager.initializeRetrievePubkeyExtension(
		ctx,
		exponentialBackoff,
		150*time.Minute,
	)
	if err != nil {
		return fmt.Errorf(
			"could not initialize retrieve pubkey extension: [%v]",
			err,
		)
	}

	logger.Infof("tbtc extensions have been initialized")

	return nil
}

type depositEventHandler func(deposit string)

type watchDepositEventFn func(
	handler depositEventHandler,
) (subscription.EventSubscription, error)

type watchKeepClosedFn func(deposit string) (
	keepClosedChan chan struct{},
	keepClosedCancel func(),
	err error,
)

type submitDepositTxFn func(deposit string) error

type backoffFn func(iteration int) time.Duration

type extensionsManager struct {
	handle Handle
}

func (em *extensionsManager) initializeRetrievePubkeyExtension(
	ctx context.Context,
	actBackoffFn backoffFn,
	timeout time.Duration,
) error {
	logger.Infof("initializing retrieve pubkey extension")

	startEventSubscription, err := em.monitorAndAct(
		ctx,
		"retrieve pubkey",
		func(handler depositEventHandler) (subscription.EventSubscription, error) {
			return em.handle.OnDepositCreated(handler)
		},
		func(handler depositEventHandler) (subscription.EventSubscription, error) {
			return em.handle.OnDepositRegisteredPubkey(handler)
		},
		em.watchKeepClosed,
		func(deposit string) error {
			return em.handle.RetrieveSignerPubkey(deposit)
		},
		actBackoffFn,
		timeout,
	)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		startEventSubscription.Unsubscribe()
		logger.Infof("retrieve pubkey extension has been disabled")
	}()

	logger.Infof("retrieve pubkey extension has been initialized")

	return nil
}

// TODO:
//  1. Filter incoming events by operator interest.
//  2. Incoming events deduplication.
//  3. Resume extensions executions after client restart.
func (em *extensionsManager) monitorAndAct(
	ctx context.Context,
	extensionName string,
	monitoringStartFn watchDepositEventFn,
	monitoringStopFn watchDepositEventFn,
	keepClosedFn watchKeepClosedFn,
	actFn submitDepositTxFn,
	actBackoffFn backoffFn,
	timeout time.Duration,
) (subscription.EventSubscription, error) {
	handleStartEvent := func(deposit string) {
		logger.Infof(
			"triggering [%v] extension execution for deposit [%v]",
			extensionName,
			deposit,
		)

		stopEventChan := make(chan struct{})

		stopEventSubscription, err := monitoringStopFn(
			func(stopEventDeposit string) {
				if deposit == stopEventDeposit {
					stopEventChan <- struct{}{}
				}
			},
		)
		if err != nil {
			logger.Errorf(
				"could not setup stop event handler for [%v] "+
					"extension execution and deposit [%v]: [%v]",
				extensionName,
				deposit,
				err,
			)
			return
		}
		defer stopEventSubscription.Unsubscribe()

		keepClosedChan, keepClosedSubscriptionCancel, err := keepClosedFn(deposit)
		if err != nil {
			logger.Errorf(
				"could not setup keep closed handler for [%v] "+
					"extension execution and deposit [%v]: [%v]",
				extensionName,
				deposit,
				err,
			)
			return
		}
		defer keepClosedSubscriptionCancel()

		timeoutChan := time.After(timeout)

		transactionAttempt := 1

	monitoring:
		for {
			select {
			case <-ctx.Done():
				logger.Infof(
					"context is done for [%v] "+
						"extension execution and deposit [%v]",
					extensionName,
					deposit,
				)
				break monitoring
			case <-stopEventChan:
				logger.Infof(
					"stop event occurred for [%v] "+
						"extension execution and deposit [%v]",
					extensionName,
					deposit,
				)
				break monitoring
			case <-keepClosedChan:
				logger.Infof(
					"keep closed event occurred for [%v] "+
						"extension execution and deposit [%v]",
					extensionName,
					deposit,
				)
				break monitoring
			case <-timeoutChan:
				err := actFn(deposit)
				if err != nil {
					if transactionAttempt == maxTransactionAttempts {
						logger.Errorf(
							"could not submit transaction "+
								"for [%v] extension execution and "+
								"deposit [%v]: [%v]; last attempt failed",
							extensionName,
							deposit,
							err,
						)
						break monitoring
					}

					backoff := actBackoffFn(transactionAttempt)

					logger.Errorf(
						"could not submit transaction "+
							"for [%v] extension execution and "+
							"deposit [%v]: [%v]; retrying after: [%v]",
						extensionName,
						deposit,
						err,
						backoff,
					)

					timeoutChan = time.After(backoff)
					transactionAttempt++
				} else {
					logger.Infof(
						"transaction for [%v] extension execution and "+
							"deposit [%v] has been submitted successfully",
						extensionName,
						deposit,
					)
					break monitoring
				}
			}
		}

		logger.Infof(
			"[%v] extension execution for deposit [%v] "+
				"has been completed",
			extensionName,
			deposit,
		)
	}

	return monitoringStartFn(
		func(deposit string) {
			go handleStartEvent(deposit)
		},
	)
}

func (em *extensionsManager) watchKeepClosed(
	deposit string,
) (chan struct{}, func(), error) {
	signalChan := make(chan struct{})

	keepAddress, err := em.handle.KeepAddress(deposit)
	if err != nil {
		return nil, nil, err
	}

	keepClosedSubscription, err := em.handle.OnKeepClosed(
		common.HexToAddress(keepAddress),
		func(_ *eth.KeepClosedEvent) {
			signalChan <- struct{}{}
		},
	)
	if err != nil {
		return nil, nil, err
	}

	keepTerminatedSubscription, err := em.handle.OnKeepTerminated(
		common.HexToAddress(keepAddress),
		func(_ *eth.KeepTerminatedEvent) {
			signalChan <- struct{}{}
		},
	)
	if err != nil {
		return nil, nil, err
	}

	unsubscribe := func() {
		keepClosedSubscription.Unsubscribe()
		keepTerminatedSubscription.Unsubscribe()
	}

	return signalChan, unsubscribe, nil
}

func exponentialBackoff(iteration int) time.Duration {
	backoffMillis := math.Pow(2, float64(iteration)) * 1000
	// #nosec G404 (insecure random number source (rand))
	// No need to use secure randomness for jitter value.
	jitterMillis := rand.Intn(100)
	return time.Duration(int(backoffMillis)+jitterMillis) * time.Millisecond
}
