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

var logger = log.Logger("app-tbtc")

// Handle represents a chain handle extended with TBTC-specific capabilities.
type Handle interface {
	eth.Handle

	Deposit
	TBTCSystem
}

// Deposit is an interface that provides ability to interact
// with Deposit contracts.
type Deposit interface {
	// KeepAddress returns the underlying keep address for the
	// provided deposit.
	KeepAddress(depositAddress string) (string, error)

	// RetrieveSignerPubkey retrieves the signer public key for the
	// provided deposit.
	RetrieveSignerPubkey(depositAddress string) error
}

// TBTCSystem is an interface that provides ability to interact
// with TBTCSystem contract.
type TBTCSystem interface {
	// OnDepositCreated installs a callback that is invoked when an
	// on-chain notification of a new deposit creation is seen.
	OnDepositCreated(
		handler func(depositAddress string),
	) (subscription.EventSubscription, error)

	// OnDepositRegisteredPubkey installs a callback that is invoked when an
	// on-chain notification of a deposit's pubkey registration is seen.
	OnDepositRegisteredPubkey(
		handler func(depositAddress string),
	) (subscription.EventSubscription, error)
}

// InitializeActions initializes actions specific for the TBTC application.
func InitializeActions(ctx context.Context, handle Handle) error {
	logger.Infof("initializing tbtc actions")

	manager := &actionsManager{handle}

	err := manager.initializeRetrievePubkeyAction(ctx)
	if err != nil {
		return fmt.Errorf(
			"could not initialize retrieve pubkey action: [%v]",
			err,
		)
	}

	logger.Infof("tbtc actions have been initialized")

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

type actionsManager struct {
	handle Handle
}

func (am *actionsManager) initializeRetrievePubkeyAction(
	ctx context.Context,
) error {
	logger.Infof("initializing retrieve pubkey action")

	subscription, err := am.monitorAndAct(
		"retrieve pubkey",
		func(handler depositEventHandler) (subscription.EventSubscription, error) {
			return am.handle.OnDepositCreated(handler)
		},
		func(handler depositEventHandler) (subscription.EventSubscription, error) {
			return am.handle.OnDepositRegisteredPubkey(handler)
		},
		am.setupKeepClosedHandler,
		func(deposit string) error {
			return am.handle.RetrieveSignerPubkey(deposit)
		},
		150*time.Minute,
	)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		subscription.Unsubscribe()
		logger.Infof("retrieve pubkey action has been disabled")
	}()

	logger.Infof("retrieve pubkey action has been initialized")

	return nil
}

// TODO:
//  1. Filter incoming events by operator interest.
//  2. Incoming events deduplication.
//  3. Resume actions after client restart.
func (am *actionsManager) monitorAndAct(
	actionName string,
	triggerEventHandlerSetup depositEventHandlerSetup,
	stopEventHandlerSetup depositEventHandlerSetup,
	keepClosedHandlerSetup keepClosedHandlerSetup,
	transactionSubmitter depositTransactionSubmitter,
	timeout time.Duration,
) (subscription.EventSubscription, error) {
	handleTriggerEvent := func(deposit string) {
		logger.Infof(
			"triggering [%v] action for deposit [%v]",
			actionName,
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
				"could not setup stop event handler for [%v] action "+
					"and deposit [%v]: [%v]",
				actionName,
				deposit,
				err,
			)
			return
		}
		defer stopSubscription.Unsubscribe()

		keepClosedChan, keepClosedCancel, err := keepClosedHandlerSetup(deposit)
		if err != nil {
			logger.Errorf(
				"could not setup keep closed handler for [%v] action "+
					"and deposit [%v]: [%v]",
				actionName,
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
					"stop event occurred for [%v] action and deposit [%v]",
					actionName,
					deposit,
				)
				break monitoring
			case <-keepClosedChan:
				logger.Infof(
					"keep closed event occurred for [%v] action and deposit [%v]",
					actionName,
					deposit,
				)
				break monitoring
			case <-timeoutChan:
				err := transactionSubmitter(deposit)
				if err != nil {
					if transactionAttempt == 3 {
						logger.Errorf(
							"could not submit transaction "+
								"for [%v] action and deposit [%v]: [%v]; "+
								"last attempt failed",
							actionName,
							deposit,
							err,
						)
						break monitoring
					}

					backoff := am.submitterBackoff(transactionAttempt)

					logger.Errorf(
						"could not submit transaction "+
							"for [%v] action and deposit [%v]: [%v]; "+
							"retrying after: [%v]",
						actionName,
						deposit,
						err,
						backoff,
					)

					timeoutChan = time.After(backoff)
					transactionAttempt++
				} else {
					logger.Infof(
						"transaction for [%v] action and deposit [%v] "+
							"has been submitted successfully",
						actionName,
						deposit,
					)
					break monitoring
				}
			}
		}

		logger.Infof(
			"[%v] action for deposit [%v] has been completed",
			actionName,
			deposit,
		)
	}

	return triggerEventHandlerSetup(
		func(deposit string) {
			go handleTriggerEvent(deposit)
		},
	)
}

func (am *actionsManager) setupKeepClosedHandler(
	deposit string,
) (chan struct{}, func(), error) {
	signalChan := make(chan struct{})

	keepAddress, err := am.handle.KeepAddress(deposit)
	if err != nil {
		return nil, nil, err
	}

	keepClosedSubscription, err := am.handle.OnKeepClosed(
		common.HexToAddress(keepAddress),
		func(_ *eth.KeepClosedEvent) {
			signalChan <- struct{}{}
		},
	)
	if err != nil {
		return nil, nil, err
	}

	keepTerminatedSubscription, err := am.handle.OnKeepTerminated(
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

func (am *actionsManager) submitterBackoff(attempt int) time.Duration {
	backoffMillis := math.Pow(2, float64(attempt)) * 1000
	// #nosec G404 (insecure random number source (rand))
	// No need to use secure randomness for jitter value.
	jitterMillis := rand.Intn(100)
	return time.Duration(int(backoffMillis) + jitterMillis)
}
