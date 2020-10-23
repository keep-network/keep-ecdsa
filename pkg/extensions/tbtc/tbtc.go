package tbtc

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"math/big"
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
func Initialize(ctx context.Context, chain Handle) error {
	logger.Infof("initializing tbtc extension")

	tbtc := newTBTC(chain)

	err := tbtc.monitorRetrievePubKey(
		ctx,
		exponentialBackoff,
		150*time.Minute, // 30 minutes before the 3 hours on-chain timeout
	)
	if err != nil {
		return fmt.Errorf(
			"could not initialize retrieve pubkey monitoring: [%v]",
			err,
		)
	}

	err = tbtc.monitorProvideRedemptionSignature(
		ctx,
		exponentialBackoff,
		105*time.Minute, // 15 minutes before the 2 hours on-chain timeout
	)
	if err != nil {
		return fmt.Errorf(
			"could not initialize provide redemption "+
				"signature monitoring: [%v]",
			err,
		)
	}

	logger.Infof("tbtc extension has been initialized")

	return nil
}

type depositEvent interface {
	getDepositAddress() string
}

type baseDepositEvent struct {
	depositAddress string
}

func (bde *baseDepositEvent) getDepositAddress() string {
	return bde.depositAddress
}

type depositRedemptionRequestedEvent struct {
	*baseDepositEvent

	digest      [32]uint8
	blockNumber uint64
}

type depositEventHandler func(depositEvent)

type watchDepositEventFn func(
	handler depositEventHandler,
) (subscription.EventSubscription, error)

type watchKeepClosedFn func(depositAddress string) (
	keepClosedChan chan struct{},
	unsubscribe func(),
	err error,
)

type depositMonitoringContext struct {
	startEvent depositEvent
}

type submitDepositTxFn func(*depositMonitoringContext) error

type backoffFn func(iteration int) time.Duration

type tbtc struct {
	chain     Handle
	eventsLog *tbtcEventsLog
}

func newTBTC(chain Handle) *tbtc {
	return &tbtc{
		chain:     chain,
		eventsLog: newTBTCEventsLog(),
	}
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
			return t.chain.OnDepositCreated(func(depositAddress string) {
				handler(&baseDepositEvent{depositAddress})
			})
		},
		func(handler depositEventHandler) (subscription.EventSubscription, error) {
			return t.chain.OnDepositRegisteredPubkey(func(depositAddress string) {
				handler(&baseDepositEvent{depositAddress})
			})
		},
		t.watchKeepClosed,
		func(monitoringContext *depositMonitoringContext) error {
			return t.chain.RetrieveSignerPubkey(
				monitoringContext.startEvent.getDepositAddress(),
			)
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

func (t *tbtc) monitorProvideRedemptionSignature(
	ctx context.Context,
	actBackoffFn backoffFn,
	timeout time.Duration,
) error {
	monitoringStartFn := func(
		handler depositEventHandler,
	) (subscription.EventSubscription, error) {
		return t.chain.OnDepositRedemptionRequested(
			func(
				depositAddress string,
				requesterAddress string,
				digest [32]uint8,
				utxoValue *big.Int,
				redeemerOutputScript []uint8,
				requestedFee *big.Int,
				outpoint []uint8,
				blockNumber uint64,
			) {
				event := &depositRedemptionRequestedEvent{
					&baseDepositEvent{depositAddress},
					digest,
					blockNumber,
				}

				// Log the event to make it available for future provide
				// redemption proof monitoring process.
				t.eventsLog.logDepositRedemptionRequestedEvent(
					depositAddress,
					event,
				)

				handler(event)
			},
		)
	}

	monitoringStopFn := func(
		handler depositEventHandler,
	) (subscription.EventSubscription, error) {
		signatureSubscription, err := t.chain.OnDepositGotRedemptionSignature(
			func(depositAddress string) {
				handler(&baseDepositEvent{depositAddress})
			},
		)
		if err != nil {
			return nil, err
		}

		redeemedSubscription, err := t.chain.OnDepositRedeemed(
			func(depositAddress string) {
				handler(&baseDepositEvent{depositAddress})
			},
		)
		if err != nil {
			return nil, err
		}

		return subscription.NewEventSubscription(
			func() {
				signatureSubscription.Unsubscribe()
				redeemedSubscription.Unsubscribe()
			},
		), nil
	}

	actFn := func(monitoringContext *depositMonitoringContext) error {
		depositRedemptionRequestedEvent, ok := monitoringContext.startEvent.(*depositRedemptionRequestedEvent)
		if !ok {
			return fmt.Errorf(
				"monitoring context contains unexpected type of start event",
			)
		}

		depositAddress := depositRedemptionRequestedEvent.depositAddress

		keepAddress, err := t.chain.KeepAddress(depositAddress)
		if err != nil {
			return err
		}

		signatureSubmittedEvents, err := t.chain.PastSignatureSubmittedEvents(
			keepAddress,
			depositRedemptionRequestedEvent.blockNumber,
		)
		if err != nil {
			return err
		}

		depositDigest := depositRedemptionRequestedEvent.digest

		for _, signatureSubmittedEvent := range signatureSubmittedEvents {
			if bytes.Equal(signatureSubmittedEvent.Digest[:], depositDigest[:]) {
				// We add 27 to the recovery ID to align it with ethereum and
				// bitcoin protocols where 27 is added to recovery ID to
				// indicate usage of uncompressed public keys.
				v := 27 + signatureSubmittedEvent.RecoveryID

				return t.chain.ProvideRedemptionSignature(
					depositAddress,
					v,
					signatureSubmittedEvent.R,
					signatureSubmittedEvent.S,
				)
			}
		}

		return fmt.Errorf(
			"could not find signature for digest: [%v]",
			depositDigest,
		)
	}

	monitoringSubscription, err := t.monitorAndAct(
		ctx,
		"provide redemption signature",
		monitoringStartFn,
		monitoringStopFn,
		t.watchKeepClosed,
		actFn,
		actBackoffFn,
		timeout,
	)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		monitoringSubscription.Unsubscribe()
		logger.Infof("provide redemption signature monitoring disabled")
	}()

	logger.Infof("provide redemption signature monitoring initialized")

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
	handleStartEvent := func(startEvent depositEvent) {
		logger.Infof(
			"starting [%v] monitoring for deposit [%v]",
			monitoringName,
			startEvent.getDepositAddress(),
		)

		stopEventChan := make(chan struct{})

		stopEventSubscription, err := monitoringStopFn(
			func(stopEvent depositEvent) {
				if startEvent.getDepositAddress() == stopEvent.getDepositAddress() {
					stopEventChan <- struct{}{}
				}
			},
		)
		if err != nil {
			logger.Errorf(
				"could not setup stop event handler for [%v] "+
					"monitoring for deposit [%v]: [%v]",
				monitoringName,
				startEvent.getDepositAddress(),
				err,
			)
			return
		}
		defer stopEventSubscription.Unsubscribe()

		keepClosedChan, keepClosedUnsubscribe, err := keepClosedFn(
			startEvent.getDepositAddress(),
		)
		if err != nil {
			logger.Errorf(
				"could not setup keep closed handler for [%v] "+
					"monitoring for deposit [%v]: [%v]",
				monitoringName,
				startEvent.getDepositAddress(),
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
					startEvent.getDepositAddress(),
				)
				break monitoring
			case <-stopEventChan:
				logger.Infof(
					"stop event occurred for [%v] "+
						"monitoring for deposit [%v]",
					monitoringName,
					startEvent.getDepositAddress(),
				)
				break monitoring
			case <-keepClosedChan:
				logger.Infof(
					"keep closed event occurred for [%v] "+
						"monitoring for deposit [%v]",
					monitoringName,
					startEvent.getDepositAddress(),
				)
				break monitoring
			case <-timeoutChan:
				logger.Infof(
					"[%v] not performed in the expected time frame "+
						"for deposit [%v]; performing the action",
					monitoringName,
					startEvent.getDepositAddress(),
				)

				err := actFn(&depositMonitoringContext{startEvent})
				if err != nil {
					if actionAttempt == maxActAttempts {
						logger.Errorf(
							"could not perform action "+
								"for [%v] monitoring for deposit [%v]: [%v]; "+
								"the maximum number of attempts reached",
							monitoringName,
							startEvent.getDepositAddress(),
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
						startEvent.getDepositAddress(),
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
			startEvent.getDepositAddress(),
		)
	}

	return monitoringStartFn(
		func(startEvent depositEvent) {
			go handleStartEvent(startEvent)
		},
	)
}

func (t *tbtc) watchKeepClosed(
	depositAddress string,
) (chan struct{}, func(), error) {
	signalChan := make(chan struct{})

	keepAddress, err := t.chain.KeepAddress(depositAddress)
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
