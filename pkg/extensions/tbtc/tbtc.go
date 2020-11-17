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

	"github.com/keep-network/keep-common/pkg/cache"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/chain/chainutil"
	"github.com/keep-network/keep-common/pkg/subscription"
	chain "github.com/keep-network/keep-ecdsa/pkg/chain"
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
)

// TODO: Resume monitoring after client restart
// Initialize initializes extension specific to the TBTC application.
func Initialize(ctx context.Context, chain chain.TBTCHandle) error {
	logger.Infof("initializing tbtc extension")

	tbtc := newTBTC(chain)

	err := tbtc.monitorRetrievePubKey(
		ctx,
		exponentialBackoff,
		165*time.Minute, // 15 minutes before the 3 hours on-chain timeout
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

	err = tbtc.monitorProvideRedemptionProof(
		ctx,
		exponentialBackoff,
		345*time.Minute, // 15 minutes before the 6 hours on-chain timeout
	)
	if err != nil {
		return fmt.Errorf(
			"could not initialize provide redemption "+
				"proof monitoring: [%v]",
			err,
		)
	}

	logger.Infof("tbtc extension has been initialized")

	return nil
}

type tbtc struct {
	chain                     chain.TBTCHandle
	monitoringLocks           sync.Map
	blockConfirmations        uint64
	monitoredDepositsCache    *cache.TimeCache
	notMonitoredDepositsCache *cache.TimeCache
	signerActionDelayStep     time.Duration
}

func newTBTC(chain chain.TBTCHandle) *tbtc {
	return &tbtc{
		chain:                     chain,
		blockConfirmations:        defaultBlockConfirmations,
		monitoredDepositsCache:    cache.NewTimeCache(monitoringCachePeriod),
		notMonitoredDepositsCache: cache.NewTimeCache(monitoringCachePeriod),
		signerActionDelayStep:     defaultSignerActionDelayStep,
	}
}

func (t *tbtc) monitorRetrievePubKey(
	ctx context.Context,
	actBackoffFn backoffFn,
	timeout time.Duration,
) error {
	initialDepositState := chain.AwaitingSignerSetup

	monitoringStartFn := func(
		handler depositEventHandler,
	) (subscription.EventSubscription, error) {
		return t.chain.OnDepositCreated(handler)
	}

	monitoringStopFn := func(
		handler depositEventHandler,
	) (subscription.EventSubscription, error) {
		return t.chain.OnDepositRegisteredPubkey(func(depositAddress string) {
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
		err := t.chain.RetrieveSignerPubkey(depositAddress)
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

	monitoringSubscription, err := t.monitorAndAct(
		ctx,
		"retrieve pubkey",
		t.shouldMonitorDeposit,
		monitoringStartFn,
		monitoringStopFn,
		t.watchKeepClosed,
		actFn,
		actBackoffFn,
		timeoutFn,
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
	initialDepositState := chain.AwaitingWithdrawalSignature

	monitoringStartFn := func(
		handler depositEventHandler,
	) (subscription.EventSubscription, error) {
		// Start right after a redemption has been requested or the redemption
		// fee has been increased.
		return t.chain.OnDepositRedemptionRequested(handler)
	}

	monitoringStopFn := func(
		handler depositEventHandler,
	) (subscription.EventSubscription, error) {
		// Stop in case the redemption signature has been provided by someone else.
		signatureSubscription, err := t.chain.OnDepositGotRedemptionSignature(
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
		if err != nil {
			return nil, err
		}

		// Stop in case the redemption proof has been provided by someone else.
		redeemedSubscription, err := t.chain.OnDepositRedeemed(
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

	actFn := func(depositAddress string) error {
		keepAddress, err := t.chain.KeepAddress(depositAddress)
		if err != nil {
			return err
		}

		redemptionRequestedEvents, err := t.chain.PastDepositRedemptionRequestedEvents(
			depositAddress,
			t.pastEventsLookupStartBlock(),
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

		signatureSubmittedEvents, err := t.chain.PastSignatureSubmittedEvents(
			keepAddress,
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
		err = t.chain.ProvideRedemptionSignature(
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

	monitoringSubscription, err := t.monitorAndAct(
		ctx,
		"provide redemption signature",
		t.shouldMonitorDeposit,
		monitoringStartFn,
		monitoringStopFn,
		t.watchKeepClosed,
		actFn,
		actBackoffFn,
		timeoutFn,
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

func (t *tbtc) monitorProvideRedemptionProof(
	ctx context.Context,
	actBackoffFn backoffFn,
	timeout time.Duration,
) error {
	initialDepositState := chain.AwaitingWithdrawalProof

	monitoringStartFn := func(
		handler depositEventHandler,
	) (subscription.EventSubscription, error) {
		// Start right after a redemption signature has been provided.
		return t.chain.OnDepositGotRedemptionSignature(handler)
	}

	monitoringStopFn := func(
		handler depositEventHandler,
	) (subscription.EventSubscription, error) {
		// Stop in case the redemption fee has been increased by someone else.
		redemptionRequestedSubscription, err := t.chain.OnDepositRedemptionRequested(
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
		if err != nil {
			return nil, err
		}

		// Stop in case the redemption proof has been provided by someone else.
		redeemedSubscription, err := t.chain.OnDepositRedeemed(
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
		if err != nil {
			return nil, err
		}

		return subscription.NewEventSubscription(
			func() {
				redemptionRequestedSubscription.Unsubscribe()
				redeemedSubscription.Unsubscribe()
			},
		), nil
	}

	actFn := func(depositAddress string) error {
		redemptionRequestedEvents, err := t.chain.PastDepositRedemptionRequestedEvents(
			depositAddress,
			t.pastEventsLookupStartBlock(),
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

		err = t.chain.IncreaseRedemptionFee(
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

		redemptionRequestedEvents, err := t.chain.PastDepositRedemptionRequestedEvents(
			depositAddress,
			t.pastEventsLookupStartBlock(),
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
		redemptionRequestedTimestamp, err := t.chain.BlockTimestamp(
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

	monitoringSubscription, err := t.monitorAndAct(
		ctx,
		"provide redemption proof",
		t.shouldMonitorDeposit,
		monitoringStartFn,
		monitoringStopFn,
		t.watchKeepClosed,
		actFn,
		actBackoffFn,
		timeoutFn,
	)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		monitoringSubscription.Unsubscribe()
		logger.Infof("provide redemption proof monitoring disabled")
	}()

	logger.Infof("provide redemption proof monitoring initialized")

	return nil
}

type shouldMonitorDepositFn func(depositAddress string) bool

type depositEventHandler func(depositAddress string)

type watchDepositEventFn func(
	handler depositEventHandler,
) (subscription.EventSubscription, error)

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
) (subscription.EventSubscription, error) {
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

		stopEventSubscription, err := monitoringStopFn(
			func(stopEventDepositAddress string) {
				if depositAddress == stopEventDepositAddress {
					stopEventChan <- struct{}{}
				}
			},
		)
		if err != nil {
			logger.Errorf(
				"could not setup stop event handler for [%v] "+
					"monitoring for deposit [%v]: [%v]",
				monitoringName,
				depositAddress,
				err,
			)
			return
		}
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

	keepAddress, err := t.chain.KeepAddress(depositAddress)
	if err != nil {
		return nil, nil, err
	}

	keepClosedSubscription, err := t.chain.OnKeepClosed(
		common.HexToAddress(keepAddress),
		func(_ *chain.KeepClosedEvent) {
			if t.waitKeepNotActiveConfirmation(keepAddress) {
				signalChan <- struct{}{}
			}
		},
	)
	if err != nil {
		return nil, nil, err
	}

	keepTerminatedSubscription, err := t.chain.OnKeepTerminated(
		common.HexToAddress(keepAddress),
		func(_ *chain.KeepTerminatedEvent) {
			if t.waitKeepNotActiveConfirmation(keepAddress) {
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

func (t *tbtc) shouldMonitorDeposit(depositAddress string) bool {
	t.monitoredDepositsCache.Sweep()
	t.notMonitoredDepositsCache.Sweep()

	if t.monitoredDepositsCache.Has(depositAddress) {
		return true
	}

	if t.notMonitoredDepositsCache.Has(depositAddress) {
		return false
	}

	signerIndex, err := t.getSignerIndex(depositAddress)
	if err != nil {
		logger.Errorf(
			"could not check if deposit [%v] should be monitored: "+
				"failed to get signer index: [%v]",
			depositAddress,
			err,
		)
		return false // return false but don't cache the result in case of error
	}

	if signerIndex < 0 {
		t.notMonitoredDepositsCache.Add(depositAddress)
		return false
	}

	t.monitoredDepositsCache.Add(depositAddress)
	return true
}

func (t *tbtc) getSignerIndex(depositAddress string) (int, error) {
	keepAddress, err := t.chain.KeepAddress(depositAddress)
	if err != nil {
		return -1, err
	}

	members, err := t.chain.GetMembers(common.HexToAddress(keepAddress))
	if err != nil {
		return -1, err
	}

	for index, member := range members {
		if member == t.chain.Address() {
			return index, nil
		}
	}

	return -1, nil
}

func (t *tbtc) getSignerActionDelay(
	depositAddress string,
) (time.Duration, error) {
	signerIndex, err := t.getSignerIndex(depositAddress)
	if err != nil {
		return 0, err
	}

	return time.Duration(signerIndex) * t.signerActionDelayStep, nil
}

func (t *tbtc) waitDepositStateChangeConfirmation(
	depositAddress string,
	initialDepositState chain.DepositState,
) bool {
	stateCheck := func() (bool, error) {
		currentState, err := t.chain.CurrentState(depositAddress)
		if err != nil {
			return false, err
		}

		return currentState != initialDepositState, nil
	}

	currentBlock, err := t.chain.BlockCounter().CurrentBlock()
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

	confirmed, err := chainutil.WaitForBlockConfirmations(
		t.chain.BlockCounter(),
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
	keepAddress string,
) bool {
	currentBlock, err := t.chain.BlockCounter().CurrentBlock()
	if err != nil {
		logger.Errorf(
			"could not get current block while confirming "+
				"keep [%v] is not active: [%v]",
			keepAddress,
			err,
		)
		return false
	}

	isKeepActive, err := chainutil.WaitForBlockConfirmations(
		t.chain.BlockCounter(),
		currentBlock,
		t.blockConfirmations,
		func() (bool, error) {
			return t.chain.IsActive(common.HexToAddress(keepAddress))
		},
	)
	if err != nil {
		logger.Errorf(
			"could not confirm if keep [%v] is not active: [%v]",
			keepAddress,
			err,
		)
		return false
	}

	return !isKeepActive
}

func (t *tbtc) pastEventsLookupStartBlock() uint64 {
	currentBlock, err := t.chain.BlockCounter().CurrentBlock()
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
