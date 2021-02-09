package ethereum

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

type tbtcApplicationHandle struct {
	bekm              *bondedEcdsaKeepManager
	tbtcSystemAddress common.Address
}

func (bekm *bondedEcdsaKeepManager) TBTCApplicationHandle() (chain.BondedECDSAKeepApplicationHandle, error) {
	return &tbtcApplicationHandle{
		// FIXME This should probably be ec.bondedECDSAKeepFactoryContract
		// FIXME instead of the whole ec kit and kaboodle.
		bekm:              bekm,
		tbtcSystemAddress: common.Address{},
	}, nil
}

// RegisterAsMemberCandidate registers client as a candidate to be selected
// to a keep.
func (tah *tbtcApplicationHandle) RegisterAsMemberCandidate() error {
	gasEstimate, err := tah.bekm.handle.bondedECDSAKeepFactoryContract.RegisterMemberCandidateGasEstimate(
		tah.tbtcSystemAddress,
	)
	if err != nil {
		return fmt.Errorf("failed to estimate gas [%v]", err)
	}

	// If we have multiple sortition pool join transactions queued - and that
	// happens when multiple operators become eligible to join at the same time,
	// e.g. after lowering the minimum bond requirement, transactions mined at
	// the end may no longer have valid gas limits as they were estimated based
	// on a different state of the pool. We add 20% safety margin to the original
	// gas estimation to account for that.
	gasEstimateWithMargin := float64(gasEstimate) * float64(1.2)
	transaction, err := tah.bekm.handle.bondedECDSAKeepFactoryContract.RegisterMemberCandidate(
		tah.tbtcSystemAddress,
		ethutil.TransactionOptions{
			GasLimit: uint64(gasEstimateWithMargin),
		},
	)
	if err != nil {
		return err
	}

	logger.Debugf("submitted RegisterMemberCandidate transaction with hash: [%x]", transaction.Hash())

	return nil
}

// IsRegisteredForApplication checks if the operator is registered
// as a signer candidate in the factory for the given application.
func (tah *tbtcApplicationHandle) IsRegisteredForApplication() (bool, error) {
	return tah.bekm.handle.bondedECDSAKeepFactoryContract.IsOperatorRegistered(
		tah.bekm.address,
		tah.tbtcSystemAddress,
	)
}

// IsEligibleForApplication checks if the operator is eligible to register
// as a signer candidate for the given application.
func (tah *tbtcApplicationHandle) IsEligibleForApplication() (bool, error) {
	return tah.bekm.handle.bondedECDSAKeepFactoryContract.IsOperatorEligible(
		tah.bekm.address,
		tah.tbtcSystemAddress,
	)
}

// IsStatusUpToDateForApplication checks if the operator's status
// is up to date in the signers' pool of the given application.
func (tah *tbtcApplicationHandle) IsStatusUpToDateForApplication() (bool, error) {
	return tah.bekm.handle.bondedECDSAKeepFactoryContract.IsOperatorUpToDate(
		tah.bekm.address,
		tah.tbtcSystemAddress,
	)
}

// UpdateStatusForApplication updates the operator's status in the signers'
// pool for the given application.
func (tah *tbtcApplicationHandle) UpdateStatusForApplication() error {
	transaction, err := tah.bekm.handle.bondedECDSAKeepFactoryContract.UpdateOperatorStatus(
		tah.bekm.address,
		tah.tbtcSystemAddress,
	)
	if err != nil {
		return err
	}

	logger.Debugf(
		"submitted UpdateOperatorStatus transaction with hash: [%x]",
		transaction.Hash(),
	)

	return nil
}

// IsOperatorAuthorized checks if the factory has the authorization to
// operate on stake represented by the provided operator.
func (tah *tbtcApplicationHandle) IsOperatorAuthorized(operatorID chain.OperatorID) (bool, error) {
	// Inside the Ethereum chain, operator ids are always addresses.
	operatorAddressString := operatorID.String()
	if !common.IsHexAddress(operatorAddressString) {
		return false, fmt.Errorf("incorrect operator address [%s]", operatorAddressString)
	}

	return tah.bekm.handle.bondedECDSAKeepFactoryContract.IsOperatorAuthorized(
		common.HexToAddress(operatorAddressString),
	)
}
