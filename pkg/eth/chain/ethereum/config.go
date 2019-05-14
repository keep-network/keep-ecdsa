package ethereum

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

const (
	// KeepTECDSAGroupContractName name of the group contract.
	KeepTECDSAGroupContractName = "KeepTECDSAGroup"
)

// Config contains configuration of Ethereum chain.
type Config struct {
	URL               string
	ContractAddresses map[string]string
}

// ContractAddress finds a given contract's address configuration and returns it
// as ethereum Address.
func (c *Config) ContractAddress(contractName string) (common.Address, error) {
	contractAddress, ok := c.ContractAddresses[contractName]
	if !ok {
		return common.Address{}, fmt.Errorf(
			"configuration for contract [%s] not found",
			contractName,
		)
	}

	return common.HexToAddress(contractAddress), nil
}
