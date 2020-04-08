package local

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
)

// OnUnbondedValueWithdrawn installs a callback that will be called on unbonded
// value withdraw for the given operator.
func (lc *localChain) OnUnbondedValueWithdrawn(
	operatorAddress common.Address,
	handler func(event *eth.UnbondedValueWithdrawnEvent),
) (subscription.EventSubscription, error) {
	panic("implement")
}

// OnBondCreated installs a callback that will be called on bond creation
// for the given operator.
func (lc *localChain) OnBondCreated(
	operatorAddress common.Address,
	handler func(event *eth.BondCreatedEvent),
) (subscription.EventSubscription, error) {
	panic("implement")
}
