package ethereum

import (
	"github.com/keep-network/keep-core/pkg/chain"
)

type ethereumStakeMonitor struct{}

func NewEthereumStakeMonitor() *ethereumStakeMonitor {
	return &ethereumStakeMonitor{}
}

func (esm *ethereumStakeMonitor) HasMinimumStake(address string) (bool, error) {
	// TODO Implementation
	return true, nil
}

func (esm *ethereumStakeMonitor) StakerFor(address string) (chain.Staker, error) {
	// TODO Implementation
	return nil, nil
}
