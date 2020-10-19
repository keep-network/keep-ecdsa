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

// InitializeExtensions initializes extensions specific to the TBTC application.
func InitializeExtensions(ctx context.Context, handle Handle) error {
	logger.Infof("initializing tbtc extensions")

	manager := &extensionsManager{handle}

	err := manager.initializeRetrievePubkeyExtension(ctx)
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

type depositEventHandlerSetup func(
	handler depositEventHandler,
) (subscription.EventSubscription, error)

type keepClosedHandlerSetup func(deposit string) (
	keepClosedChan chan struct{},
	keepClosedCancel func(),
	err error,
)

type depositTransactionSubmitter func(deposit string) error

type extensionsManager struct {
	handle Handle
}

func (em *extensionsManager) initializeRetrievePubkeyExtension(
	ctx context.Context,
) error {
	logger.Infof("initializing retrieve pubkey extension")

	subscription, err := em.initializeExtension(
		"retrieve pubkey",
		func(handler depositEventHandler) (subscription.EventSubscription, error) {
			return em.handle.OnDepositCreated(handler)
		},
		func(handler depositEventHandler) (subscription.EventSubscription, error) {
			return em.handle.OnDepositRegisteredPubkey(handler)
		},
		em.setupKeepClosedHandler,
		func(deposit string) error {
			return em.handle.RetrieveSignerPubkey(deposit)
		},
		150*time.Minute,
	)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		subscription.Unsubscribe()
		logger.Infof("retrieve pubkey extension has been disabled")
	}()

	logger.Infof("retrieve pubkey extension has been initialized")

	return nil
}

// TODO:
//  1. Filter incoming events by operator interest.
//  2. Incoming events deduplication.
//  3. Resume extensions executions after client restart.
func (em *extensionsManager) initializeExtension(
	extensionName string,
	triggerEventHandlerSetup depositEventHandlerSetup,
	stopEventHandlerSetup depositEventHandlerSetup,
	keepClosedHandlerSetup keepClosedHandlerSetup,
	transactionSubmitter depositTransactionSubmitter,
	timeout time.Duration,
) (subscription.EventSubscription, error) {
	handleTriggerEvent := func(deposit string) {
		logger.Infof(
			"triggering [%v] extension execution for deposit [%v]",
			extensionName,
			deposit,
		)

		stopEventChan := make(chan struct{})
		keepClosedChan := make(chan struct{})

		stopSubscription, err := stopEventHandlerSetup(
			func(stopEventDeposit string) {
				if deposit == stopEventDeposit {
					stopEventChan <- struct{}{}
				}
			},
		)
		if err != nil {
			logger.Errorf(
				"could not setup stop event handler for [%v] extension "+
					"and deposit [%v]: [%v]",
				extensionName,
				deposit,
				err,
			)
			return
		}
		defer stopSubscription.Unsubscribe()

		keepClosedChan, keepClosedCancel, err := keepClosedHandlerSetup(deposit)
		if err != nil {
			logger.Errorf(
				"could not setup keep closed handler for [%v] extension "+
					"and deposit [%v]: [%v]",
				extensionName,
				deposit,
				err,
			)
			return
		}
		defer keepClosedCancel()

		timeoutChan := time.After(timeout)

		transactionAttempt := 1

	monitoring:
		for {
			select {
			case <-stopEventChan:
				logger.Infof(
					"stop event occurred for [%v] extension "+
						"and deposit [%v]",
					extensionName,
					deposit,
				)
				break monitoring
			case <-keepClosedChan:
				logger.Infof(
					"keep closed event occurred for [%v] extension "+
						"and deposit [%v]",
					extensionName,
					deposit,
				)
				break monitoring
			case <-timeoutChan:
				err := transactionSubmitter(deposit)
				if err != nil {
					if transactionAttempt == 3 {
						logger.Errorf(
							"could not submit transaction "+
								"for [%v] extension and deposit [%v]: [%v]; "+
								"last attempt failed",
							extensionName,
							deposit,
							err,
						)
						break monitoring
					}

					backoff := em.submitterBackoff(transactionAttempt)

					logger.Errorf(
						"could not submit transaction "+
							"for [%v] extension and deposit [%v]: [%v]; "+
							"retrying after: [%v]",
						extensionName,
						deposit,
						err,
						backoff,
					)

					timeoutChan = time.After(backoff)
					transactionAttempt++
				} else {
					logger.Infof(
						"transaction for [%v] extension and deposit [%v] "+
							"has been submitted successfully",
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

	return triggerEventHandlerSetup(
		func(deposit string) {
			go handleTriggerEvent(deposit)
		},
	)
}

func (em *extensionsManager) setupKeepClosedHandler(
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

func (em *extensionsManager) submitterBackoff(attempt int) time.Duration {
	backoffMillis := math.Pow(2, float64(attempt)) * 1000
	// #nosec G404 (insecure random number source (rand))
	// No need to use secure randomness for jitter value.
	jitterMillis := rand.Intn(100)
	return time.Duration(int(backoffMillis) + jitterMillis)
}
