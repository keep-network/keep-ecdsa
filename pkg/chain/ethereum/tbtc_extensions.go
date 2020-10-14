package ethereum

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/tbtc/bindings/go/contract"
)

// TBTCEthereumChain represents an Ethereum chain handle with
// TBTC-specific capabilities.
type TBTCEthereumChain struct {
	*EthereumChain

	depositLogContract *contract.DepositLog
}

// WithTBTCExtensions extends the Ethereum chain handle with
// TBTC-specific capabilities.
func WithTBTCExtensions(
	ethereumChain *EthereumChain,
	depositLogContractAddress string,
) (*TBTCEthereumChain, error) {
	if !common.IsHexAddress(depositLogContractAddress) {
		return nil, fmt.Errorf("incorrect deposit log contract address")
	}

	depositLogContract, err := contract.NewDepositLog(
		common.HexToAddress(depositLogContractAddress),
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
		depositLogContract: depositLogContract,
	}, nil
}

// OnDepositCreated installs a callback that is invoked when an
// on-chain notification of a new deposit creation is seen.
func (tec *TBTCEthereumChain) OnDepositCreated(
	handler func(depositAddress, keepAddress string, timestamp *big.Int),
) subscription.EventSubscription {
	subscription, err := tec.depositLogContract.WatchCreated(
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
