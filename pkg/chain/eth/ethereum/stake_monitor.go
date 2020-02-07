package ethereum

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-core/pkg/chain"
)

type ethereumStakeMonitor struct {
	ethereum *EthereumChain
}

func (esm *ethereumStakeMonitor) HasMinimumStake(address string) (bool, error) {
	if !common.IsHexAddress(address) {
		return false, fmt.Errorf("not a valid ethereum address: %v", address)
	}

	return esm.ethereum.HasMinimumStake(common.HexToAddress(address))
}

func (esm *ethereumStakeMonitor) StakerFor(address string) (chain.Staker, error) {
	return nil, fmt.Errorf("not implemented")
}

func (ec *EthereumChain) StakeMonitor() (chain.StakeMonitor, error) {
	return &ethereumStakeMonitor{ec}, nil
}
