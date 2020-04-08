package client

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
)

// checkStatusAndRegisterForApplication checks whether the operator is
// registered as a member candidate for keep for the given application.
// If not, checks operators's eligibility and retries until the operator is
// eligible. Eventually, once the operator is eligible, it is registered
// as a keep member candidate.
// Also, once the client is confirmed as registered, it triggers the monitoring
// process to keep the operator's status up to date in the pool.
// If operator status in the pool cannot be monitored, e.g. when operator is
// removed from the pool it triggers the registration process from the begining.
func checkStatusAndRegisterForApplication(
	ctx context.Context,
	ethereumChain eth.Handle,
	application common.Address,
) {
RegistrationLoop:
	for {
		select {
		case <-ctx.Done():
			return
		default:
			isRegistered, err := ethereumChain.IsRegisteredForApplication(application)
			if err != nil {
				logger.Errorf(
					"failed to check if member is registered for application [%s]: [%v]",
					application.String(),
					err,
				)
				continue RegistrationLoop
			}

			if !isRegistered {
				// if the operator is not registered, we need to register it and
				// wait until registration is confirmed
				registerAsMemberCandidate(ctx, ethereumChain, application)
				waitUntilRegistered(ctx, ethereumChain, application)
			}

			// once the registration is confirmed or if the client is already
			// registered, we can start to monitor the status
			if err := monitorSignerPoolStatus(ctx, ethereumChain, application); err != nil {
				logger.Errorf("failed on signer pool status monitoring: [%v]", err)
				continue RegistrationLoop
			}
		}
	}
}

// registerAsMemberCandidate checks current operator's eligibility to become
// keep member candidate for the given application and if it is positive,
// registers the operator as a keep member candidate for the given application.
// If the operator is not eligible, it executes the check for each new mined
// block until the operator is finally eligible and can be registered.
func registerAsMemberCandidate(
	parentCtx context.Context,
	ethereumChain eth.Handle,
	application common.Address,
) {
	// If the operator is eligible right now for registering as a member
	// candidate for the application, we register the operator.
	isEligible, err := ethereumChain.IsEligibleForApplication(application)
	if err != nil {
		logger.Errorf(
			"failed to check operator eligibility for application [%s]: [%v]",
			application.String(),
			err,
		)
	}
	if isEligible {
		logger.Infof(
			"registering member candidate for application [%s]",
			application.String(),
		)
		err := ethereumChain.RegisterAsMemberCandidate(application)
		if err != nil {
			logger.Errorf(
				"failed to register member candidate for application [%s]: [%v]",
				application.String(),
				err,
			)
		} else {
			return
		}
	}

	// If the operator is not yet eligible to be registered as a member candidate
	// for the application, we start monitoring eligibility each now block.
	// We do the same in case the registration of eligible operator failed for
	// some reason. As soon as the operator is eligible, we will proceed with
	// the registration.
	registerAsMemberCandidateWhenEligible(parentCtx, ethereumChain, application)
}

// registerAsMemberCandidateWhenEligible for each new block checks the operator's
// eligibility to be registered as a keep member candidate for the application.
// As soon as the operator becomes eligible, function triggers the registration.
func registerAsMemberCandidateWhenEligible(
	parentCtx context.Context,
	ethereumChain eth.Handle,
	application common.Address,
) {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	newBlockChan := ethereumChain.BlockCounter().WatchBlocks(ctx)
	for {
		select {
		case <-newBlockChan:
			isEligible, err := ethereumChain.IsEligibleForApplication(application)
			if err != nil {
				logger.Errorf(
					"failed to check operator eligibility for application [%s]: [%v]",
					application.String(),
					err,
				)
				continue
			}

			if !isEligible {
				// if the operator is not yet eligible wait for the next
				// block and execute the check again
				logger.Warningf(
					"operator is not eligible for application [%s]",
					application.String(),
				)
				continue
			}

			// if the operator is eligible, register it as a keep member
			// candidate for this application
			logger.Infof(
				"registering member candidate for application [%s]",
				application.String(),
			)
			if err := ethereumChain.RegisterAsMemberCandidate(application); err != nil {
				logger.Errorf(
					"failed to register member candidate for application [%s]: [%v]",
					application.String(),
					err,
				)
				continue
			}

			// we cancel the context in case the registration was successful,
			// we don't want to do it again
			cancel()
		case <-ctx.Done():
			return
		}
	}
}

// waitUntilRegistered blocks until the operator is registered as a keep member
// candidate for the given application.
func waitUntilRegistered(
	ctx context.Context,
	ethereumChain eth.Handle,
	application common.Address,
) {
	newBlockChan := ethereumChain.BlockCounter().WatchBlocks(ctx)

	for {
		select {
		case <-newBlockChan:
			isRegistered, err := ethereumChain.IsRegisteredForApplication(application)
			if err != nil {
				logger.Errorf(
					"failed to check if member is registered for application [%s]: [%v]",
					application.String(),
					err,
				)
				continue
			}

			if isRegistered {
				logger.Infof(
					"operator is registered for application [%s]",
					application.String(),
				)
				return
			}

			logger.Infof(
				"operator is not yet registered for application [%s], waiting...",
				application.String(),
			)
		case <-ctx.Done():
			return
		}
	}
}

// monitorSignerPoolStatus tracks operator's state in the signing pool
// (staking weight, bonding) and updates the status when it gets out of date.
func monitorSignerPoolStatus(
	ctx context.Context,
	ethereumChain eth.Handle,
	application common.Address,
) error {
	logger.Debugf(
		"starting monitoring operatator status for application [%s]",
		application.String(),
	)

	updateTrigger := make(chan interface{})

	unsubscribeAll, err := subscribeStatusUpdateEvents(
		ethereumChain,
		updateTrigger,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to subscribe for status update events for application [%s]: [%v]",
			application.String(),
			err,
		)
	}
	defer unsubscribeAll()

	logger.Infof("subscribed for status update events")

	for {
		select {
		case <-updateTrigger:
			logger.Infof(
				"operator status update triggered for application: [%s]",
				application.String(),
			)

			isUpToDate, err := ethereumChain.IsStatusUpToDateForApplication(application)
			if err != nil {
				return fmt.Errorf(
					"failed to check operator status for application [%s]: [%v]",
					application.String(),
					err,
				)
			}

			if isUpToDate {
				logger.Debugf(
					"operator status is up to date for application [%s]",
					application.String(),
				)
			} else {
				logger.Infof(
					"updating operator status for application [%s]",
					application.String(),
				)

				err := ethereumChain.UpdateStatusForApplication(application)
				if err != nil {
					return fmt.Errorf(
						"failed to update operator status for application [%s]: [%v]",
						application.String(),
						err,
					)
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func subscribeStatusUpdateEvents(
	ethereumChain eth.Handle,
	updateTrigger chan<- interface{},
) (func(), error) {
	subscriptions := []subscription.EventSubscription{}

	unsubscribeAll := func() {
		for _, s := range subscriptions {
			s.Unsubscribe()
		}
	}

	subscriptionUnbondedValueWithdrawn, err := ethereumChain.OnUnbondedValueWithdrawn(
		ethereumChain.Address(),
		func(event *eth.UnbondedValueWithdrawnEvent) {
			updateTrigger <- event
		},
	)
	if err != nil {
		unsubscribeAll()
		return nil, fmt.Errorf("failed on registering for unbonded value withdraw: [%v]", err)
	}
	subscriptions = append(subscriptions, subscriptionUnbondedValueWithdrawn)

	subscriptionBondCreated, err := ethereumChain.OnBondCreated(
		ethereumChain.Address(),
		func(event *eth.BondCreatedEvent) {
			updateTrigger <- event
		},
	)
	if err != nil {
		unsubscribeAll()
		return nil, fmt.Errorf("failed on registering for bond creation: [%v]", err)
	}
	subscriptions = append(subscriptions, subscriptionBondCreated)

	subscriptionTokenSlashed, err := ethereumChain.OnTokensSlashed(
		ethereumChain.Address(),
		func(event *eth.TokensSlashedEvent) {
			updateTrigger <- event
		},
	)
	if err != nil {
		unsubscribeAll()
		return nil, fmt.Errorf("failed on registering for tokens slashing: [%v]", err)
	}
	subscriptions = append(subscriptions, subscriptionTokenSlashed)

	subscriptionTokenSeized, err := ethereumChain.OnTokensSeized(
		ethereumChain.Address(),
		func(event *eth.TokensSeizedEvent) {
			updateTrigger <- event
		},
	)
	if err != nil {
		unsubscribeAll()
		return nil, fmt.Errorf("failed on registering for tokens seizure: [%v]", err)
	}
	subscriptions = append(subscriptions, subscriptionTokenSeized)

	return unsubscribeAll, nil
}
