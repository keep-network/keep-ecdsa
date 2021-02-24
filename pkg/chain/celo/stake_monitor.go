//+build celo

package celo

import (
	"fmt"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	relaychain "github.com/keep-network/keep-core/pkg/beacon/relay/chain"
	"github.com/keep-network/keep-core/pkg/chain"
)

type celoStakeMonitor struct {
	celo *Chain
}

func (esm *celoStakeMonitor) HasMinimumStake(address string) (bool, error) {
	if !common.IsHexAddress(address) {
		return false, fmt.Errorf("not a valid celo address: %v", address)
	}

	return esm.celo.HasMinimumStake(toExternalAddress(common.HexToAddress(address)))
}

func (esm *celoStakeMonitor) StakerFor(address string) (chain.Staker, error) {
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("not a valid celo address: %v", address)
	}

	return &celoStaker{
		address: address,
		celo:    esm.celo,
	}, nil
}

// StakeMonitor generates a new `chain.StakeMonitor` from the chain
func (c *Chain) StakeMonitor() (chain.StakeMonitor, error) {
	return &celoStakeMonitor{c}, nil
}

type celoStaker struct {
	address string
	celo    *Chain
}

func (es *celoStaker) Address() relaychain.StakerAddress {
	return common.HexToAddress(es.address).Bytes()
}

func (es *celoStaker) Stake() (*big.Int, error) {
	return es.celo.BalanceOf(toExternalAddress(common.HexToAddress(es.address)))
}
