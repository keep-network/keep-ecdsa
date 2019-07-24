package ethereum

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// Definitions of contract names.
const (
	ECDSAKeepFactoryContractName = "ECDSAKeepFactory"
)

// Config contains configuration of Ethereum chain.
type Config struct {
	URL string
	// ContractAddresses map holds contract name as a key and contract address
	// as a value.
	ContractAddresses map[string]string
	// Private key for client's address to interact with Ethereum chain.
	// TODO: Provide more secure solution, e.g. reading protected files from disk.
	PrivateKey string
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

	if !common.IsHexAddress(contractAddress) {
		return common.Address{}, fmt.Errorf(
			"configured address [%v] for contract [%v] is not valid hex address",
			contractAddress,
			contractName,
		)
	}

	return common.HexToAddress(contractAddress), nil
}
