package client

import (
	"context"
	"fmt"
	"time"

	"github.com/keep-network/keep-common/pkg/chain/ethlike"

	corechain "github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

const statusCheckIntervalBlocks = 100

// retryDelay defines the delay between retries related to the registration logic
// that do not have their own specific values (like for example `eligibilityRetryDelay`
// for sortition pool join eligibility checks).
const retryDelay = 1 * time.Second

// eligibilityRetryDelay defines the delay between checks whether the operator
// is eligible to join the sortition pool.
const eligibilityRetryDelay = 20 * time.Minute

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
	blockCounter corechain.BlockCounter,
	application chain.BondedECDSAKeepApplicationHandle,
) {
RegistrationLoop:
	for {
		select {
		case <-ctx.Done():
			return
		default:
			isRegistered, err := application.IsRegisteredForApplication()
			if err != nil {
				logger.Errorf(
					"failed to check if member is registered for application [%s]: [%v]",
					application.ID(),
					err,
				)
				time.Sleep(retryDelay) // TODO: #413 Replace with backoff.
				continue RegistrationLoop
			}

			if !isRegistered {
				// if the operator is not registered, we need to register it and
				// wait until registration is confirmed
				registerAsMemberCandidate(ctx, blockCounter, application)
				waitUntilRegistered(ctx, blockCounter, application)
			}

			// once the registration is confirmed or if the client is already
			// registered, we can start to monitor the status
			if err := monitorSignerPoolStatus(ctx, blockCounter, application); err != nil {
				logger.Errorf(
					"failed on signer pool status monitoring; please inspect "+
						"signer's unbonded value and stake: [%v]",
					err,
				)
				time.Sleep(retryDelay) // TODO: #413 Replace with backoff.
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
	blockCounter corechain.BlockCounter,
	application chain.BondedECDSAKeepApplicationHandle,
) {
	// If the operator is eligible right now for registering as a member
	// candidate for the application, we register the operator.
	isEligible, err := application.IsEligibleForApplication()
	if err != nil {
		logger.Errorf(
			"failed to check operator eligibility for application [%s]: [%v]",
			application.ID(),
			err,
		)
	}
	if isEligible {
		logger.Infof(
			"registering member candidate for application [%s]",
			application.ID(),
		)
		err := application.RegisterAsMemberCandidate()
		if err != nil {
			logger.Errorf(
				"failed to register member candidate for application [%s]: [%v]",
				application.ID(),
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
	registerAsMemberCandidateWhenEligible(parentCtx, blockCounter, application)
}

// registerAsMemberCandidateWhenEligible for each new block checks the operator's
// eligibility to be registered as a keep member candidate for the application.
// As soon as the operator becomes eligible, function triggers the registration.
//
// TODO Move this to application.RegisterAsMemberCandidateWhenEligible() <-chan error?
func registerAsMemberCandidateWhenEligible(
	parentCtx context.Context,
	blockCounter corechain.BlockCounter,
	application chain.BondedECDSAKeepApplicationHandle,
) {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	newBlockChan := blockCounter.WatchBlocks(ctx)
	for {
		select {
		case <-newBlockChan:
			isEligible, err := application.IsEligibleForApplication()
			if err != nil {
				logger.Errorf(
					"failed to check operator eligibility for application [%s]: [%v]",
					application.ID(),
					err,
				)
				time.Sleep(retryDelay) // TODO: #413 Replace with backoff.
				continue
			}

			if !isEligible {
				// if the operator is not yet eligible wait for the next
				// block and execute the check again
				logger.Warningf(
					"operator is not eligible for application [%s]",
					application.ID(),
				)
				time.Sleep(eligibilityRetryDelay) // TODO: #413 Replace with backoff.
				continue
			}

			// if the operator is eligible, register it as a keep member
			// candidate for this application
			logger.Infof(
				"registering member candidate for application [%s]",
				application.ID(),
			)
			if err := application.RegisterAsMemberCandidate(); err != nil {
				logger.Errorf(
					"failed to register member candidate for application [%s]: [%v]",
					application.ID(),
					err,
				)
				time.Sleep(retryDelay) // TODO: #413 Replace with backoff.
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
	blockCounter corechain.BlockCounter,
	application chain.BondedECDSAKeepApplicationHandle,
) {
	newBlockChan := blockCounter.WatchBlocks(ctx)

	for {
		select {
		case <-newBlockChan:
			isRegistered, err := application.IsRegisteredForApplication()
			if err != nil {
				logger.Errorf(
					"failed to check if member is registered for application [%s]: [%v]",
					application.ID(),
					err,
				)
				time.Sleep(retryDelay) // TODO: #413 Replace with backoff.
				continue
			}

			if isRegistered {
				logger.Infof(
					"operator is registered for application [%s]",
					application.ID(),
				)
				return
			}

			logger.Infof(
				"operator is not yet registered for application [%s], waiting...",
				application.ID(),
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
	blockCounter corechain.BlockCounter,
	application chain.BondedECDSAKeepApplicationHandle,
) error {
	logger.Debugf(
		"starting monitoring operatator status for application [%s]",
		application.ID(),
	)

	startingBlock, err := blockCounter.CurrentBlock()
	if err != nil {
		return err
	}

	statusCheckTrigger, err := blockCounter.BlockHeightWaiter(
		startingBlock + statusCheckIntervalBlocks,
	)
	if err != nil {
		return err
	}

	for {
		select {
		case statusCheckBlock := <-statusCheckTrigger:
			logger.Debugf(
				"operator status check for application [%s] "+
					"triggered at block [%v]",
				application.ID(),
				statusCheckBlock,
			)

			isUpToDate, err := application.IsStatusUpToDateForApplication()
			if err != nil {
				return fmt.Errorf(
					"failed to check operator status for application [%s]: [%v]",
					application.ID(),
					err,
				)
			}

			if isUpToDate {
				logger.Debugf(
					"operator status is up to date for application [%s]",
					application.ID(),
				)
			} else {
				logger.Infof(
					"updating operator status for application [%s]",
					application.ID(),
				)

				err := application.UpdateStatusForApplication()
				if err != nil {
					return fmt.Errorf(
						"failed to update operator status for application [%s]: [%v]",
						application.ID(),
						err,
					)
				}

				isRegistered, err := ethlike.WaitForBlockConfirmations(
					blockCounter,
					statusCheckBlock,
					blockConfirmations,
					func() (bool, error) {
						return application.IsRegisteredForApplication()
					},
				)
				if err != nil {
					return fmt.Errorf(
						"failed to confirm that operator is registered "+
							"for application [%s]: [%v]",
						application.ID(),
						err,
					)
				}

				if !isRegistered {
					return fmt.Errorf(
						"operator is no longer registered for application [%s]",
						application.ID(),
					)
				}
			}

			statusCheckTrigger, err = blockCounter.BlockHeightWaiter(
				statusCheckBlock + statusCheckIntervalBlocks,
			)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
