package ethereum

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/contract"
)

type bondedEcdsaKeepManager struct {
	handle  *ethereumChain
	address common.Address
}

func (ec *ethereumChain) BondedECDSAKeepManager() (chain.BondedECDSAKeepManager, error) {
	return &bondedEcdsaKeepManager{
		// FIXME This should probably be ec.bondedECDSAKeepFactoryContract
		// FIXME instead of the whole ec kit and kaboodle.
		handle:  ec,
		address: ec.accountKey.Address,
	}, nil
}

// OnBondedECDSAKeepCreated installs a callback that is invoked when an on-chain
// notification of a new ECDSA keep creation is seen.
func (bekm *bondedEcdsaKeepManager) OnBondedECDSAKeepCreated(
	handler func(event *chain.BondedECDSAKeepCreatedEvent),
) subscription.EventSubscription {
	onEvent := func(
		KeepAddress common.Address,
		Members []common.Address,
		Owner common.Address,
		Application common.Address,
		HonestThreshold *big.Int,
		blockNumber uint64,
	) {
		contract, _ := bekm.getKeepContract(KeepAddress)
		memberIDs := []chain.KeepMemberID{}
		for _, memberAddress := range Members {
			memberIDs = append(memberIDs, combinedChainID(memberAddress))
		}

		// FIXME This should pass a BondedECSAKeep instead of an address +
		// FIXME members.
		handler(&chain.BondedECDSAKeepCreatedEvent{
			Keep: &bondedEcdsaKeep{
				bekm.handle,
				contract,
				KeepAddress,
			},
			Members:         memberIDs,
			HonestThreshold: HonestThreshold.Uint64(),
			BlockNumber:     blockNumber,
		})
	}

	return bekm.handle.bondedECDSAKeepFactoryContract.BondedECDSAKeepCreated(
		nil,
		nil,
		nil,
		nil,
	).OnEvent(onEvent)
}

func (bekm *bondedEcdsaKeepManager) getKeepContract(address common.Address) (*contract.BondedECDSAKeep, error) {
	bondedECDSAKeepContract, err := contract.NewBondedECDSAKeep(
		address,
		bekm.handle.accountKey,
		bekm.handle.client,
		bekm.handle.nonceManager,
		bekm.handle.miningWaiter,
		bekm.handle.blockCounter,
		bekm.handle.transactionMutex,
	)
	if err != nil {
		return nil, err
	}

	return bondedECDSAKeepContract, nil
}

// GetKeepCount returns number of keeps.
func (bekm *bondedEcdsaKeepManager) GetKeepCount() (*big.Int, error) {
	return bekm.handle.bondedECDSAKeepFactoryContract.GetKeepCount()
}

// GetKeepAtIndex returns the address of the keep at the given index.
func (bekm *bondedEcdsaKeepManager) GetKeepAtIndex(
	keepIndex *big.Int,
) (chain.BondedECDSAKeepHandle, error) {
	keepAddress, err := bekm.handle.bondedECDSAKeepFactoryContract.GetKeepAtIndex(keepIndex)
	if err != nil {
		return nil, err
	}

	return bekm.GetKeepWithID(keepAddress)
}

// IsOperatorAuthorized checks if the factory has the authorization to
// operate on stake represented by the provided operator.
func (bekm *bondedEcdsaKeepManager) IsOperatorAuthorized(operatorID chain.OperatorID) (bool, error) {
	// Inside the Ethereum chain, operator ids are always addresses.
	operatorAddressString := operatorID.String()
	if !common.IsHexAddress(operatorAddressString) {
		return false, fmt.Errorf("incorrect operator address [%s]", operatorAddressString)
	}

	return bekm.handle.bondedECDSAKeepFactoryContract.IsOperatorAuthorized(
		common.HexToAddress(operatorAddressString),
	)
}
