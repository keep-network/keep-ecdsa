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
	chain "github.com/keep-network/keep-ecdsa/pkg/chain"
)

var logger = log.Logger("tbtc-extension")

const maxActAttempts = 3

// Initialize initializes extension specific to the TBTC application.
func Initialize(ctx context.Context, handle Handle) error {
	logger.Infof("initializing tbtc extension")

	tbtc := &tbtc{handle}

	err := tbtc.monitorRetrievePubKey(
		ctx,
		exponentialBackoff,
		150*time.Minute,
	)
	if err != nil {
		return fmt.Errorf(
			"could not initialize retrieve pubkey monitoring: [%v]",
			err,
		)
	}

	logger.Infof("tbtc extension has been initialized")

	return nil
}

type depositEventHandler func(deposit string)

type watchDepositEventFn func(
	handler depositEventHandler,
) (subscription.EventSubscription, error)

type watchKeepClosedFn func(deposit string) (
	keepClosedChan chan struct{},
	unsubscribe func(),
	err error,
)

type submitDepositTxFn func(deposit string) error

type backoffFn func(iteration int) time.Duration

type tbtc struct {
	chain Handle
}

func (t *tbtc) monitorRetrievePubKey(
	ctx context.Context,
	actBackoffFn backoffFn,
	timeout time.Duration,
) error {
	monitoringSubscription, err := t.monitorAndAct(
		ctx,
		"retrieve pubkey",
		func(handler depositEventHandler) (subscription.EventSubscription, error) {
			return t.chain.OnDepositCreated(handler)
		},
		func(handler depositEventHandler) (subscription.EventSubscription, error) {
			return t.chain.OnDepositRegisteredPubkey(handler)
		},
		t.watchKeepClosed,
		func(deposit string) error {
			return t.chain.RetrieveSignerPubkey(deposit)
		},
		actBackoffFn,
		timeout,
	)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		monitoringSubscription.Unsubscribe()
		logger.Infof("retrieve pubkey monitoring disabled")
	}()

	logger.Infof("retrieve pubkey monitoring initialized")

	return nil
}

// TODO:
//  1. Filter incoming events by operator interest.
//  2. Incoming events deduplication.
//  3. Resume monitoring after client restart.
func (t *tbtc) monitorAndAct(
	ctx context.Context,
	monitoringName string,
	monitoringStartFn watchDepositEventFn,
	monitoringStopFn watchDepositEventFn,
	keepClosedFn watchKeepClosedFn,
	actFn submitDepositTxFn,
	actBackoffFn backoffFn,
	timeout time.Duration,
) (subscription.EventSubscription, error) {
	handleStartEvent := func(deposit string) {
		logger.Infof(
			"starting [%v] monitoring for deposit [%v]",
			monitoringName,
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
					"monitoring for deposit [%v]: [%v]",
				monitoringName,
				deposit,
				err,
			)
			return
		}
		defer stopEventSubscription.Unsubscribe()

		keepClosedChan, keepClosedUnsubscribe, err := keepClosedFn(deposit)
		if err != nil {
			logger.Errorf(
				"could not setup keep closed handler for [%v] "+
					"monitoring for deposit [%v]: [%v]",
				monitoringName,
				deposit,
				err,
			)
			return
		}
		defer keepClosedUnsubscribe()

		timeoutChan := time.After(timeout)

		actionAttempt := 1

	monitoring:
		for {
			select {
			case <-ctx.Done():
				logger.Infof(
					"context is done for [%v] "+
						"monitoring for deposit [%v]",
					monitoringName,
					deposit,
				)
				break monitoring
			case <-stopEventChan:
				logger.Infof(
					"stop event occurred for [%v] "+
						"monitoring for deposit [%v]",
					monitoringName,
					deposit,
				)
				break monitoring
			case <-keepClosedChan:
				logger.Infof(
					"keep closed event occurred for [%v] "+
						"monitoring for deposit [%v]",
					monitoringName,
					deposit,
				)
				break monitoring
			case <-timeoutChan:
				logger.Infof(
					"[%v] not performed in the expected time frame "+
						"for deposit [%v]; performing the action",
					monitoringName,
					deposit,
				)

				err := actFn(deposit)
				if err != nil {
					if actionAttempt == maxActAttempts {
						logger.Errorf(
							"could not perform action "+
								"for [%v] monitoring for deposit [%v]: [%v]; "+
								"the maximum number of attempts reached",
							monitoringName,
							deposit,
							err,
						)
						break monitoring
					}

					backoff := actBackoffFn(actionAttempt)

					logger.Errorf(
						"could not perform action "+
							"for [%v] monitoring for deposit [%v]: [%v]; "+
							"retrying after: [%v]",
						monitoringName,
						deposit,
						err,
						backoff,
					)

					timeoutChan = time.After(backoff)
					actionAttempt++
				} else {
					break monitoring
				}
			}
		}

		logger.Infof(
			"stopped [%v] monitoring for deposit [%v]",
			monitoringName,
			deposit,
		)
	}

	return monitoringStartFn(
		func(deposit string) {
			go handleStartEvent(deposit)
		},
	)
}

func (t *tbtc) watchKeepClosed(
	deposit string,
) (chan struct{}, func(), error) {
	signalChan := make(chan struct{})

	keepAddress, err := t.chain.KeepAddress(deposit)
	if err != nil {
		return nil, nil, err
	}

	keepClosedSubscription, err := t.chain.OnKeepClosed(
		common.HexToAddress(keepAddress),
		func(_ *chain.KeepClosedEvent) {
			signalChan <- struct{}{}
		},
	)
	if err != nil {
		return nil, nil, err
	}

	keepTerminatedSubscription, err := t.chain.OnKeepTerminated(
		common.HexToAddress(keepAddress),
		func(_ *chain.KeepTerminatedEvent) {
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

// Computes the exponential backoff value for given iteration.
// For each iteration the result value will be in range:
// - iteration 1: [2000ms, 2100ms)
// - iteration 2: [4000ms, 4100ms)
// - iteration 3: [8000ms, 8100ms)
// - iteration n: [2^n * 1000ms, (2^n * 1000ms) + 100ms)
func exponentialBackoff(iteration int) time.Duration {
	backoffMillis := math.Pow(2, float64(iteration)) * 1000
	// #nosec G404 (insecure random number source (rand))
	// No need to use secure randomness for jitter value.
	jitterMillis := rand.Intn(100)
	return time.Duration(int(backoffMillis)+jitterMillis) * time.Millisecond
}
