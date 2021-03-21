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
	celo *celoChain
}

func (esm *celoStakeMonitor) HasMinimumStake(address string) (bool, error) {
	if !common.IsHexAddress(address) {
		return false, fmt.Errorf("not a valid celo address: %v", address)
	}

	return esm.celo.hasMinimumStake(common.HexToAddress(address))
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
func (cc *celoChain) StakeMonitor() (chain.StakeMonitor, error) {
	return &celoStakeMonitor{cc}, nil
}

type celoStaker struct {
	address string
	celo    *celoChain
}

func (es *celoStaker) Address() relaychain.StakerAddress {
	return common.HexToAddress(es.address).Bytes()
}

func (es *celoStaker) Stake() (*big.Int, error) {
	return es.celo.balanceOf(common.HexToAddress(es.address))
}
