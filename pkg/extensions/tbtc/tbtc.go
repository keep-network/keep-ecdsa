package tbtc

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/keep-network/keep-common/pkg/chain/ethlike"

	"github.com/keep-network/keep-common/pkg/cache"

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/subscription"
	corechain "github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/utils"
)

var logger = log.Logger("tbtc-extension")

const (
	// Maximum number of action attempts before giving up and returning
	// a monitoring error.
	maxActAttempts = 3

	// Determines how many blocks from the past should be included
	// during the past events lookup.
	pastEventsLookbackBlocks = 10000

	// Number of blocks which should elapse before confirming
	// the given chain state expectations.
	defaultBlockConfirmations = 12

	// Determines how long the monitoring cache will maintain its entries about
	// which deposits should be monitored by this client instance.
	monitoringCachePeriod = 24 * time.Hour

	// Used to calculate the action delay factor for the given signer index
	// to avoid all signers executing the same action for deposit at the
	// same time.
	defaultSignerActionDelayStep = 5 * time.Minute

	// The timeout for confirming initial state of the deposit upon receiving
	// start signal but before setting up monitoring.
	confirmInitialStateTimeout = 30 * time.Second
)

// Initialize initializes extension specific to the TBTC application.
// TODO: Resume monitoring after client restart
func Initialize(
	ctx context.Context,
	tbtcHandle chain.TBTCHandle,
	blockCounter corechain.BlockCounter,
	blockTimestamp func(blockNumber *big.Int) (uint64, error),
) {
	logger.Infof("initializing tbtc extension")

	tbtc := newTBTC(
		tbtcHandle,
		blockCounter,
		blockTimestamp,
	)

	tbtc.monitorRetrievePubKey(
		ctx,
		exponentialBackoff,
		165*time.Minute, // 15 minutes before the 3 hours on-chain timeout
	)

	tbtc.monitorProvideRedemptionSignature(
		ctx,
		exponentialBackoff,
		105*time.Minute, // 15 minutes before the 2 hours on-chain timeout
	)

	tbtc.monitorProvideRedemptionProof(
		ctx,
		exponentialBackoff,
		345*time.Minute, // 15 minutes before the 6 hours on-chain timeout
	)

	logger.Infof("tbtc extension has been initialized")
}

type tbtc struct {
	handle         chain.TBTCHandle
	blockCounter   corechain.BlockCounter
	blockTimestamp func(blockNumber *big.Int) (uint64, error)

	monitoringLocks        sync.Map
	blockConfirmations     uint64
	memberDepositsCache    *cache.TimeCache
	notMemberDepositsCache *cache.TimeCache
	signerActionDelayStep  time.Duration
}

func newTBTC(
	tbtcHandle chain.TBTCHandle,
	blockCounter corechain.BlockCounter,
	blockTimestamp func(blockNumber *big.Int) (uint64, error),
) *tbtc {
	return &tbtc{
		handle:         tbtcHandle,
		blockCounter:   blockCounter,
		blockTimestamp: blockTimestamp,

		blockConfirmations:     defaultBlockConfirmations,
		memberDepositsCache:    cache.NewTimeCache(monitoringCachePeriod),
		notMemberDepositsCache: cache.NewTimeCache(monitoringCachePeriod),
		signerActionDelayStep:  defaultSignerActionDelayStep,
	}
}

func (t *tbtc) monitorRetrievePubKey(
	ctx context.Context,
	actBackoffFn backoffFn,
	timeout time.Duration,
) {
	initialDepositState := chain.AwaitingSignerSetup

	monitoringStartFn := func(
		handler depositEventHandler,
	) subscription.EventSubscription {
		return t.handle.OnDepositCreated(handler)
	}

	shouldMonitorFn := func(depositAddress string) bool {
		return t.shouldMonitorDeposit(
			confirmInitialStateTimeout,
			depositAddress,
			initialDepositState,
		)
	}

	monitoringStopFn := func(
		handler depositEventHandler,
	) subscription.EventSubscription {
		return t.handle.OnDepositRegisteredPubkey(func(depositAddress string) {
			if t.waitDepositStateChangeConfirmation(
				depositAddress,
				initialDepositState,
			) {
				handler(depositAddress)
			} else {
				logger.Warningf(
					"retrieve pubkey monitoring stop event for "+
						"deposit [%v] is not confirmed; "+
						"monitoring will be continued",
					depositAddress,
				)
			}
		})
	}

	actFn := func(depositAddress string) error {
		err := t.handle.RetrieveSignerPubkey(depositAddress)
		if err != nil {
			return err
		}

		if !t.waitDepositStateChangeConfirmation(
			depositAddress,
			initialDepositState,
		) {
			return fmt.Errorf("deposit state change is not confirmed")
		}

		return nil
	}

	timeoutFn := func(depositAddress string) (time.Duration, error) {
		actionDelay, err := t.getSignerActionDelay(depositAddress)
		if err != nil {
			return 0, err
		}

		return timeout + actionDelay, nil
	}

	monitoringSubscription := t.monitorAndAct(
		ctx,
		"retrieve pubkey",
		shouldMonitorFn,
		monitoringStartFn,
		monitoringStopFn,
		t.watchKeepClosed,
		actFn,
		actBackoffFn,
		timeoutFn,
	)

	go func() {
		<-ctx.Done()
		monitoringSubscription.Unsubscribe()
		logger.Infof("retrieve pubkey monitoring disabled")
	}()

	logger.Infof("retrieve pubkey monitoring initialized")
}

func (t *tbtc) monitorProvideRedemptionSignature(
	ctx context.Context,
	actBackoffFn backoffFn,
	timeout time.Duration,
) {
	initialDepositState := chain.AwaitingWithdrawalSignature

	monitoringStartFn := func(
		handler depositEventHandler,
	) subscription.EventSubscription {
		// Start right after a redemption has been requested or the redemption
		// fee has been increased.
		return t.handle.OnDepositRedemptionRequested(handler)
	}

	shouldMonitorFn := func(depositAddress string) bool {
		return t.shouldMonitorDeposit(
			confirmInitialStateTimeout,
			depositAddress,
			initialDepositState,
		)
	}

	monitoringStopFn := func(
		handler depositEventHandler,
	) subscription.EventSubscription {
		// Stop in case the redemption signature has been provided by someone else.
		signatureSubscription := t.handle.OnDepositGotRedemptionSignature(
			func(depositAddress string) {
				if t.waitDepositStateChangeConfirmation(
					depositAddress,
					initialDepositState,
				) {
					handler(depositAddress)
				} else {
					logger.Warningf(
						"provide redemption signature monitoring stop "+
							"event for deposit [%v] is not confirmed; "+
							"monitoring will be continued",
						depositAddress,
					)
				}
			},
		)

		// Stop in case the redemption proof has been provided by someone else.
		redeemedSubscription := t.handle.OnDepositRedeemed(
			func(depositAddress string) {
				if t.waitDepositStateChangeConfirmation(
					depositAddress,
					initialDepositState,
				) {
					handler(depositAddress)
				} else {
					logger.Warningf(
						"provide redemption signature monitoring stop "+
							"event for deposit [%v] is not confirmed; "+
							"monitoring will be continued",
						depositAddress,
					)
				}
			},
		)

		return subscription.NewEventSubscription(
			func() {
				signatureSubscription.Unsubscribe()
				redeemedSubscription.Unsubscribe()
			},
		)
	}

	actFn := func(depositAddress string) error {
		keep, err := t.handle.Keep(depositAddress)
		if err != nil {
			return err
		}

		redemptionRequestedEvents, err := t.handle.PastDepositRedemptionRequestedEvents(
			t.pastEventsLookupStartBlock(),
			depositAddress,
		)
		if err != nil {
			return err
		}

		if len(redemptionRequestedEvents) == 0 {
			return fmt.Errorf(
				"no redemption requested events found for deposit: [%v]",
				depositAddress,
			)
		}

		latestRedemptionRequestedEvent :=
			redemptionRequestedEvents[len(redemptionRequestedEvents)-1]

		signatureSubmittedEvents, err := keep.PastSignatureSubmittedEvents(
			latestRedemptionRequestedEvent.BlockNumber,
		)
		if err != nil {
			return err
		}

		if len(signatureSubmittedEvents) == 0 {
			return fmt.Errorf(
				"no signature submitted events found for deposit: [%v]",
				depositAddress,
			)
		}

		latestSignatureSubmittedEvent :=
			signatureSubmittedEvents[len(signatureSubmittedEvents)-1]

		depositDigest := latestRedemptionRequestedEvent.Digest

		if !bytes.Equal(latestSignatureSubmittedEvent.Digest[:], depositDigest[:]) {
			return fmt.Errorf(
				"could not find signature for digest: [%v]",
				depositDigest,
			)
		}

		// We add 27 to the recovery ID to align it with ethereum and
		// bitcoin protocols where 27 is added to recovery ID to
		// indicate usage of uncompressed public keys.
		err = t.handle.ProvideRedemptionSignature(
			depositAddress,
			27+latestSignatureSubmittedEvent.RecoveryID,
			latestSignatureSubmittedEvent.R,
			latestSignatureSubmittedEvent.S,
		)
		if err != nil {
			return err
		}

		if !t.waitDepositStateChangeConfirmation(
			depositAddress,
			initialDepositState,
		) {
			return fmt.Errorf("deposit state change is not confirmed")
		}

		return nil
	}

	timeoutFn := func(depositAddress string) (time.Duration, error) {
		actionDelay, err := t.getSignerActionDelay(depositAddress)
		if err != nil {
			return 0, err
		}

		return timeout + actionDelay, nil
	}

	monitoringSubscription := t.monitorAndAct(
		ctx,
		"provide redemption signature",
		shouldMonitorFn,
		monitoringStartFn,
		monitoringStopFn,
		t.watchKeepClosed,
		actFn,
		actBackoffFn,
		timeoutFn,
	)

	go func() {
		<-ctx.Done()
		monitoringSubscription.Unsubscribe()
		logger.Infof("provide redemption signature monitoring disabled")
	}()

	logger.Infof("provide redemption signature monitoring initialized")
}

func (t *tbtc) monitorProvideRedemptionProof(
	ctx context.Context,
	actBackoffFn backoffFn,
	timeout time.Duration,
) {
	initialDepositState := chain.AwaitingWithdrawalProof

	monitoringStartFn := func(
		handler depositEventHandler,
	) subscription.EventSubscription {
		// Start right after a redemption signature has been provided.
		return t.handle.OnDepositGotRedemptionSignature(handler)
	}

	shouldMonitorFn := func(depositAddress string) bool {
		return t.shouldMonitorDeposit(
			confirmInitialStateTimeout,
			depositAddress,
			initialDepositState,
		)
	}

	monitoringStopFn := func(
		handler depositEventHandler,
	) subscription.EventSubscription {
		// Stop in case the redemption fee has been increased by someone else.
		redemptionRequestedSubscription := t.handle.OnDepositRedemptionRequested(
			func(depositAddress string) {
				if t.waitDepositStateChangeConfirmation(
					depositAddress,
					initialDepositState,
				) {
					handler(depositAddress)
				} else {
					logger.Warningf(
						"provide redemption proof monitoring stop "+
							"event for deposit [%v] is not confirmed; "+
							"monitoring will be continued",
						depositAddress,
					)
				}
			},
		)

		// Stop in case the redemption proof has been provided by someone else.
		redeemedSubscription := t.handle.OnDepositRedeemed(
			func(depositAddress string) {
				if t.waitDepositStateChangeConfirmation(
					depositAddress,
					initialDepositState,
				) {
					handler(depositAddress)
				} else {
					logger.Warningf(
						"provide redemption proof monitoring stop "+
							"event for deposit [%v] is not confirmed; "+
							"monitoring will be continued",
						depositAddress,
					)
				}
			},
		)

		return subscription.NewEventSubscription(
			func() {
				redemptionRequestedSubscription.Unsubscribe()
				redeemedSubscription.Unsubscribe()
			},
		)
	}

	actFn := func(depositAddress string) error {
		redemptionRequestedEvents, err := t.handle.PastDepositRedemptionRequestedEvents(
			t.pastEventsLookupStartBlock(),
			depositAddress,
		)
		if err != nil {
			return err
		}

		if len(redemptionRequestedEvents) == 0 {
			return fmt.Errorf(
				"no redemption requested events found for deposit: [%v]",
				depositAddress,
			)
		}

		// TODO: Check whether the redemption proof can be submitted by
		//  interacting with the BTC chain. If yes, construct and submit
		//  the proof. If not, try to increase the redemption fee.

		latestRedemptionRequestedEvent :=
			redemptionRequestedEvents[len(redemptionRequestedEvents)-1]

		// Deposit expects that the fee is always increased by a constant value
		// equal to the fee of the initial redemption request.
		feeBumpStep := big.NewInt(0)
		if len(redemptionRequestedEvents) == 1 {
			feeBumpStep = latestRedemptionRequestedEvent.RequestedFee // initial fee
		} else {
			// When there are many events on-chain we don't need to get the very
			// first one, it is enough to calculate a difference between the
			// latest fee and the one before the latest fee.
			feeBumpStep = new(big.Int).Sub(
				latestRedemptionRequestedEvent.RequestedFee,
				redemptionRequestedEvents[len(redemptionRequestedEvents)-2].RequestedFee,
			)
		}

		previousOutputValue := new(big.Int).Sub(
			latestRedemptionRequestedEvent.UtxoValue,
			latestRedemptionRequestedEvent.RequestedFee,
		)

		newOutputValue := new(big.Int).Sub(
			previousOutputValue,
			feeBumpStep,
		)

		err = t.handle.IncreaseRedemptionFee(
			depositAddress,
			toLittleEndianBytes(previousOutputValue),
			toLittleEndianBytes(newOutputValue),
		)
		if err != nil {
			return err
		}

		if !t.waitDepositStateChangeConfirmation(
			depositAddress,
			initialDepositState,
		) {
			return fmt.Errorf("deposit state change is not confirmed")
		}

		return nil
	}

	timeoutFn := func(depositAddress string) (time.Duration, error) {
		// Get the seconds timestamp in the moment when this function is
		// invoked. This is when the monitoring starts in response of
		// the `GotRedemptionSignature` event.
		gotRedemptionSignatureTimestamp := uint64(time.Now().Unix())

		redemptionRequestedEvents, err := t.handle.PastDepositRedemptionRequestedEvents(
			t.pastEventsLookupStartBlock(),
			depositAddress,
		)
		if err != nil {
			return 0, err
		}

		if len(redemptionRequestedEvents) == 0 {
			return 0, fmt.Errorf(
				"no redemption requested events found for deposit: [%v]",
				depositAddress,
			)
		}

		latestRedemptionRequestedEvent :=
			redemptionRequestedEvents[len(redemptionRequestedEvents)-1]

		// Get the seconds timestamp for the latest redemption request.
		redemptionRequestedTimestamp, err := t.blockTimestamp(
			new(big.Int).SetUint64(latestRedemptionRequestedEvent.BlockNumber),
		)
		if err != nil {
			return 0, err
		}

		// We must shift the constant timeout value by subtracting the time
		// elapsed between the redemption request and the redemption signature.
		// This way we obtain a value close to the redemption proof timeout
		// and it doesn't matter when the redemption signature arrives.
		timeoutShift := time.Duration(
			gotRedemptionSignatureTimestamp-redemptionRequestedTimestamp,
		) * time.Second

		actionDelay, err := t.getSignerActionDelay(depositAddress)
		if err != nil {
			return 0, err
		}

		return (timeout - timeoutShift) + actionDelay, nil
	}

	monitoringSubscription := t.monitorAndAct(
		ctx,
		"provide redemption proof",
		shouldMonitorFn,
		monitoringStartFn,
		monitoringStopFn,
		t.watchKeepClosed,
		actFn,
		actBackoffFn,
		timeoutFn,
	)

	go func() {
		<-ctx.Done()
		monitoringSubscription.Unsubscribe()
		logger.Infof("provide redemption proof monitoring disabled")
	}()

	logger.Infof("provide redemption proof monitoring initialized")
}

type shouldMonitorDepositFn func(depositAddress string) bool

type depositEventHandler func(depositAddress string)

type watchDepositEventFn func(
	handler depositEventHandler,
) subscription.EventSubscription

type watchKeepClosedFn func(depositAddress string) (
	keepClosedChan chan struct{},
	unsubscribe func(),
	err error,
)

type submitDepositTxFn func(depositAddress string) error

type backoffFn func(iteration int) time.Duration

type timeoutFn func(depositAddress string) (time.Duration, error)

func (t *tbtc) monitorAndAct(
	ctx context.Context,
	monitoringName string,
	shouldMonitorFn shouldMonitorDepositFn,
	monitoringStartFn watchDepositEventFn,
	monitoringStopFn watchDepositEventFn,
	keepClosedFn watchKeepClosedFn,
	actFn submitDepositTxFn,
	actBackoffFn backoffFn,
	timeoutFn timeoutFn,
) subscription.EventSubscription {
	handleStartEvent := func(depositAddress string) {
		if !shouldMonitorFn(depositAddress) {
			return
		}

		if !t.acquireMonitoringLock(depositAddress, monitoringName) {
			logger.Warningf(
				"[%v] monitoring for deposit [%v] is already running",
				monitoringName,
				depositAddress,
			)
			return
		}
		defer t.releaseMonitoringLock(depositAddress, monitoringName)

		logger.Infof(
			"starting [%v] monitoring for deposit [%v]",
			monitoringName,
			depositAddress,
		)

		stopEventChan := make(chan struct{})

		stopEventSubscription := monitoringStopFn(
			func(stopEventDepositAddress string) {
				if depositAddress == stopEventDepositAddress {
					stopEventChan <- struct{}{}
				}
			},
		)
		defer stopEventSubscription.Unsubscribe()

		keepClosedChan, keepClosedUnsubscribe, err := keepClosedFn(
			depositAddress,
		)
		if err != nil {
			logger.Errorf(
				"could not setup keep closed handler for [%v] "+
					"monitoring for deposit [%v]: [%v]",
				monitoringName,
				depositAddress,
				err,
			)
			return
		}
		defer keepClosedUnsubscribe()

		timeout, err := timeoutFn(depositAddress)
		if err != nil {
			logger.Errorf(
				"could determine timeout value for [%v] "+
					"monitoring for deposit [%v]: [%v]",
				monitoringName,
				depositAddress,
				err,
			)
			return
		}

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
					depositAddress,
				)
				break monitoring
			case <-stopEventChan:
				logger.Infof(
					"stop event occurred for [%v] "+
						"monitoring for deposit [%v]",
					monitoringName,
					depositAddress,
				)
				break monitoring
			case <-keepClosedChan:
				logger.Infof(
					"keep closed event occurred for [%v] "+
						"monitoring for deposit [%v]",
					monitoringName,
					depositAddress,
				)
				break monitoring
			case <-timeoutChan:
				logger.Infof(
					"[%v] not performed in the expected time frame "+
						"for deposit [%v]; performing the action",
					monitoringName,
					depositAddress,
				)

				err := actFn(depositAddress)
				if err != nil {
					if actionAttempt == maxActAttempts {
						logger.Errorf(
							"could not perform action "+
								"for [%v] monitoring for deposit [%v]: [%v]; "+
								"the maximum number of attempts reached",
							monitoringName,
							depositAddress,
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
						depositAddress,
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
			depositAddress,
		)
	}

	return monitoringStartFn(
		func(depositAddress string) {
			go handleStartEvent(depositAddress)
		},
	)
}

func (t *tbtc) watchKeepClosed(
	depositAddress string,
) (chan struct{}, func(), error) {
	signalChan := make(chan struct{})

	keep, err := t.handle.Keep(depositAddress)
	if err != nil {
		return nil, nil, err
	}

	keepClosedSubscription, err := keep.OnKeepClosed(
		func(_ *chain.KeepClosedEvent) {
			if t.waitKeepNotActiveConfirmation(keep) {
				signalChan <- struct{}{}
			}
		},
	)
	if err != nil {
		return nil, nil, err
	}

	keepTerminatedSubscription, err := keep.OnKeepTerminated(
		func(_ *chain.KeepTerminatedEvent) {
			if t.waitKeepNotActiveConfirmation(keep) {
				signalChan <- struct{}{}
			}
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

func (t *tbtc) shouldMonitorDeposit(
	confirmStateTimeout time.Duration,
	depositAddress string,
	expectedInitialState chain.DepositState,
) bool {
	t.memberDepositsCache.Sweep()
	t.notMemberDepositsCache.Sweep()

	if t.notMemberDepositsCache.Has(depositAddress) {
		return false
	}

	hasInitialState, err := utils.ConfirmWithTimeoutDefaultBackoff(
		confirmStateTimeout,
		func(ctx context.Context) (bool, error) {
			currentState, err := t.handle.CurrentState(depositAddress)
			if err != nil {
				return false, err
			}

			return currentState == expectedInitialState, nil
		},
	)
	if err != nil {
		logger.Errorf(
			"could not check if deposit [%v] should be monitored: "+
				"failed to confirm initial state: [%v]",
			depositAddress,
			err,
		)
		// return false but don't cache the result in case of an error
		return false
	}
	if !hasInitialState {
		// false start signal, probably an old event
		return false
	}

	if t.memberDepositsCache.Has(depositAddress) {
		return true
	}

	signerIndex, err := t.getSignerIndex(depositAddress)
	if err != nil {
		logger.Errorf(
			"could not check if deposit [%v] should be monitored: "+
				"failed to get signer index: [%v]",
			depositAddress,
			err,
		)
		// return false but don't cache the result in case of an error
		return false
	}

	if signerIndex < 0 {
		t.notMemberDepositsCache.Add(depositAddress)
		return false
	}

	t.memberDepositsCache.Add(depositAddress)
	return true
}

func (t *tbtc) getSignerIndex(depositAddress string) (int, error) {
	keep, err := t.handle.Keep(depositAddress)
	if err != nil {
		return -1, err
	}

	return keep.OperatorIndex()
}

func (t *tbtc) getSignerActionDelay(
	depositAddress string,
) (time.Duration, error) {
	signerIndex, err := t.getSignerIndex(depositAddress)
	if err != nil {
		return 0, err
	}

	// just in case this function is not invoked in the right context
	if signerIndex < 0 {
		return 0, fmt.Errorf("signer index is less than zero")
	}

	return time.Duration(signerIndex) * t.signerActionDelayStep, nil
}

func (t *tbtc) waitDepositStateChangeConfirmation(
	depositAddress string,
	initialDepositState chain.DepositState,
) bool {
	stateCheck := func() (bool, error) {
		currentState, err := t.handle.CurrentState(depositAddress)
		if err != nil {
			return false, err
		}

		return currentState != initialDepositState, nil
	}

	currentBlock, err := t.blockCounter.CurrentBlock()
	if err != nil {
		logger.Errorf(
			"could not get current block while confirming "+
				"state [%v] change for deposit [%v]: [%v]",
			initialDepositState,
			depositAddress,
			err,
		)
		return false
	}

	confirmed, err := ethlike.WaitForBlockConfirmations(
		t.blockCounter,
		currentBlock,
		t.blockConfirmations,
		stateCheck,
	)
	if err != nil {
		logger.Errorf(
			"could not confirm state [%v] change for deposit [%v]: [%v]",
			initialDepositState,
			depositAddress,
			err,
		)
		return false
	}

	return confirmed
}

func (t *tbtc) waitKeepNotActiveConfirmation(
	keep chain.BondedECDSAKeepHandle,
) bool {
	currentBlock, err := t.blockCounter.CurrentBlock()
	if err != nil {
		logger.Errorf(
			"could not get current block while confirming "+
				"keep [%v] is not active: [%v]",
			keep.ID(),
			err,
		)
		return false
	}

	isKeepActive, err := ethlike.WaitForBlockConfirmations(
		t.blockCounter,
		currentBlock,
		t.blockConfirmations,
		func() (bool, error) {
			return keep.IsActive()
		},
	)
	if err != nil {
		logger.Errorf(
			"could not confirm if keep [%v] is not active: [%v]",
			keep.ID(),
			err,
		)
		return false
	}

	return !isKeepActive
}

func (t *tbtc) pastEventsLookupStartBlock() uint64 {
	currentBlock, err := t.blockCounter.CurrentBlock()
	if err != nil {
		return 0 // if something went wrong, start from block `0`
	}

	if currentBlock <= pastEventsLookbackBlocks {
		return 0
	}

	return currentBlock - pastEventsLookbackBlocks
}

func (t *tbtc) acquireMonitoringLock(depositAddress, monitoringName string) bool {
	_, isExistingKey := t.monitoringLocks.LoadOrStore(
		monitoringLockKey(depositAddress, monitoringName),
		true,
	)

	return !isExistingKey
}

func (t *tbtc) releaseMonitoringLock(depositAddress, monitoringName string) {
	t.monitoringLocks.Delete(monitoringLockKey(depositAddress, monitoringName))
}

func monitoringLockKey(
	depositAddress string,
	monitoringName string,
) string {
	return fmt.Sprintf(
		"%v-%v",
		depositAddress,
		strings.ReplaceAll(monitoringName, " ", ""),
	)
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

func toLittleEndianBytes(value *big.Int) [8]byte {
	var valueBytes [8]byte
	binary.LittleEndian.PutUint64(valueBytes[:], value.Uint64())
	return valueBytes
}
