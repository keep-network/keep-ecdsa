package ethereum

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
)

// OnUnbondedValueWithdrawn installs a callback that will be called on unbonded
// value withdraw for the given operator.
func (ec *EthereumChain) OnUnbondedValueWithdrawn(
	operatorAddress common.Address,
	handler func(event *eth.UnbondedValueWithdrawnEvent),
) (subscription.EventSubscription, error) {
	return ec.keepBondingContract.WatchUnbondedValueWithdrawn(
		func(
			Operator common.Address,
			Amount *big.Int,
			blockNumber uint64,
		) {
			handler(&eth.UnbondedValueWithdrawnEvent{
				Operator: Operator,
				Amount:   Amount,
			})
		},
		func(err error) error {
			return fmt.Errorf("watch unbonded value withdrawn failed: [%v]", err)
		},
		[]common.Address{operatorAddress},
	)
}

// OnBondCreated installs a callback that will be called on bond creation
// for the given operator.
func (ec *EthereumChain) OnBondCreated(
	operatorAddress common.Address,
	handler func(event *eth.BondCreatedEvent),
) (subscription.EventSubscription, error) {
	return ec.keepBondingContract.WatchBondCreated(
		func(
			Operator common.Address,
			Holder common.Address,
			SortitionPool common.Address,
			ReferenceID *big.Int,
			Amount *big.Int,
			blockNumber uint64,
		) {
			handler(&eth.BondCreatedEvent{
				Operator:    Operator,
				Holder:      Holder,
				SignerPool:  SortitionPool,
				ReferenceID: ReferenceID,
				Amount:      Amount,
			})
		},
		func(err error) error {
			return fmt.Errorf("watch bond creation failed: [%v]", err)
		},
		[]common.Address{operatorAddress},
		[]common.Address{},
		[]common.Address{},
	)
}
