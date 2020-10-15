package ethereum

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/tbtc/go/contract"
)

// TBTCEthereumChain represents an Ethereum chain handle with
// TBTC-specific capabilities.
type TBTCEthereumChain struct {
	*EthereumChain

	tbtcSystemContract *contract.TBTCSystem
}

// WithTBTCExtensions extends the Ethereum chain handle with
// TBTC-specific capabilities.
func WithTBTCExtensions(
	ethereumChain *EthereumChain,
	tbtcSystemContractAddress string,
) (*TBTCEthereumChain, error) {
	if !common.IsHexAddress(tbtcSystemContractAddress) {
		return nil, fmt.Errorf("incorrect tbtc system contract address")
	}

	tbtcSystemContract, err := contract.NewTBTCSystem(
		common.HexToAddress(tbtcSystemContractAddress),
		ethereumChain.accountKey,
		ethereumChain.client,
		ethereumChain.nonceManager,
		ethereumChain.miningWaiter,
		ethereumChain.transactionMutex,
	)
	if err != nil {
		return nil, err
	}

	return &TBTCEthereumChain{
		EthereumChain:      ethereumChain,
		tbtcSystemContract: tbtcSystemContract,
	}, nil
}

// OnDepositCreated installs a callback that is invoked when an
// on-chain notification of a new deposit creation is seen.
func (tec *TBTCEthereumChain) OnDepositCreated(
	handler func(depositAddress, keepAddress string, timestamp *big.Int),
) subscription.EventSubscription {
	subscription, err := tec.tbtcSystemContract.WatchCreated(
		func(
			DepositContractAddress common.Address,
			KeepAddress common.Address,
			Timestamp *big.Int,
			blockNumber uint64,
		) {
			handler(DepositContractAddress.Hex(), KeepAddress.Hex(), Timestamp)
		},
		func(err error) error {
			return fmt.Errorf("watch deposit created failed: [%v]", err)
		},
		nil,
		nil,
	)
	if err != nil {
		logger.Errorf("could not watch Created event: [%v]", err)
	}

	return subscription
}
