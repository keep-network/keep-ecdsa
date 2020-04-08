package ethereum

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
)

// OnTokensSlashed installs a callback that will be called on token stake
// slash for the given operator.
func (ec *EthereumChain) OnTokensSlashed(
	operatorAddress common.Address,
	handler func(event *eth.TokensSlashedEvent),
) (subscription.EventSubscription, error) {
	return ec.tokenStakingContract.WatchTokensSlashed(
		func(
			Operator common.Address,
			Amount *big.Int,
			blockNumber uint64,
		) {
			handler(&eth.TokensSlashedEvent{
				Operator: Operator,
				Amount:   Amount,
			})
		},
		func(err error) error {
			return fmt.Errorf("watch tokens slashing failed: [%v]", err)
		},
		[]common.Address{operatorAddress},
	)
}

// OnTokensSeized installs a callback that will be called on token stake
// seizure for the given operator.
func (ec *EthereumChain) OnTokensSeized(
	operatorAddress common.Address,
	handler func(event *eth.TokensSeizedEvent),
) (subscription.EventSubscription, error) {
	return ec.tokenStakingContract.WatchTokensSeized(
		func(
			Operator common.Address,
			Amount *big.Int,
			blockNumber uint64,
		) {
			handler(&eth.TokensSeizedEvent{
				Operator: Operator,
				Amount:   Amount,
			})
		},
		func(err error) error {
			return fmt.Errorf("watch tokens seizure failed: [%v]", err)
		},
		[]common.Address{operatorAddress},
	)
}
