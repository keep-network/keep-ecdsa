package local

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
)

// OnTokensSlashed installs a callback that will be called on token stake
// slash for the given operator.
func (lc *localChain) OnTokensSlashed(
	operatorAddress common.Address,
	handler func(event *eth.TokensSlashedEvent),
) (subscription.EventSubscription, error) {
	panic("implement")
}

// OnTokensSeized installs a callback that will be called on token stake
// seizure for the given operator.
func (lc *localChain) OnTokensSeized(
	operatorAddress common.Address,
	handler func(event *eth.TokensSeizedEvent),
) (subscription.EventSubscription, error) {
	panic("implement")
}
