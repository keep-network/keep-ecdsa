package ethereum

import (
	"github.com/keep-network/keep-common/pkg/chain/ethereum"
)

type EthStakeChainHandle struct {
	*EthereumChain

	// keepFactoryContract *contract.FullyBackedBondedECDSAKeepFactory
}

// TODO: Implementation of the chain.Handle interface.
func NewEthStakeChainHandle(
	ethereumChain *EthereumChain,
	config *ethereum.Config,
) (*EthStakeChainHandle, error) {
	// keepFactoryContractAddress, err := config.ContractAddress(
	// 	FullyBackedBondedECDSAKeepFactoryContractName,
	// )
	// if err != nil {
	// 	return nil, err
	// }
	//
	// keepFactoryContract, err := contract.NewFullyBackedBondedECDSAKeepFactory(
	// 	*keepFactoryContractAddress,
	// 	ethereumChain.accountKey,
	// 	ethereumChain.client,
	// 	ethereumChain.nonceManager,
	// 	ethereumChain.miningWaiter,
	// 	ethereumChain.transactionMutex,
	// )
	// if err != nil {
	// 	return nil, err
	// }

	return &EthStakeChainHandle{
		EthereumChain: ethereumChain,
		// keepFactoryContract: keepFactoryContract,
	}, nil
}
