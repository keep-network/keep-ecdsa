package ethereum

import (
	"github.com/keep-network/keep-core/pkg/chain"
)

type ethereumStakeMonitor struct{}

func NewEthereumStakeMonitor() *ethereumStakeMonitor {
	return &ethereumStakeMonitor{}
}

func (esm *ethereumStakeMonitor) HasMinimumStake(address string) (bool, error) {
	// TODO Will be addressed in https://github.com/keep-network/keep-tecdsa/pull/192
	return true, nil
}

func (esm *ethereumStakeMonitor) StakerFor(address string) (chain.Staker, error) {
	// TODO Will be addressed in https://github.com/keep-network/keep-tecdsa/pull/192
	return nil, nil
}
