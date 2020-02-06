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

	Account Account
}

// Account is a struct that contains the configuration for accessing Ethereum
// network.
type Account struct {
	// Keyfile is a full path to a key file. Normally this file is one of the
	// imported keys in your local Ethereum server. It can normally be found in
	// a directory <some-path>/data/keystore/ and starts with its creation date
	// "UTC--.*".
	KeyFile string

	// KeyFilePassword is the password used to unlock the account specified in
	// KeyFile.
	KeyFilePassword string
}

// ContractAddress finds a given contract's address configuration and returns it
// as ethereum Address.
func (c *Config) ContractAddress(contractName string) (common.Address, error) {
	contractAddress, ok := c.ContractAddresses[contractName]
	if !ok {
		return common.Address{}, fmt.Errorf(
			"failed to find configuration for contract [%s]",
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
