package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/keep-network/keep-tecdsa/pkg/chain/btc/blockcypher"
	"github.com/keep-network/keep-tecdsa/pkg/chain/btc/electrum"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/ethereum"
)

const passwordEnvVariable = "KEEP_ETHEREUM_PASSWORD"

// Config is the top level config structure.
type Config struct {
	Electrum    electrum.Config
	BlockCypher blockcypher.Config
	Ethereum    ethereum.Config
}

// ReadConfig reads in the configuration file in .toml format. Ethereum key file
// password is expected to be provided as environment variable.
func ReadConfig(filePath string) (*Config, error) {
	config := &Config{}
	if _, err := toml.DecodeFile(filePath, config); err != nil {
		return nil, fmt.Errorf("failed to decode file [%s]: [%v]", filePath, err)
	}

	config.Ethereum.Account.KeyFilePassword = os.Getenv(passwordEnvVariable)

	return config, nil
}
