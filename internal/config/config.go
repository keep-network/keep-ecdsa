package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/keep-network/keep-tecdsa/pkg/chain/blockcypher"
	"github.com/keep-network/keep-tecdsa/pkg/chain/electrum"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/ethereum"
)

// Config is the top level config structure.
type Config struct {
	Electrum    electrum.Config
	BlockCypher blockcypher.Config
	Ethereum    ethereum.Config
}

// ReadConfig reads in the configuration file in .toml format.
func ReadConfig(filePath string) (*Config, error) {
	config := &Config{}
	if _, err := toml.DecodeFile(filePath, config); err != nil {
		return nil, fmt.Errorf("unable to decode .toml file [%s] error [%s]", filePath, err)
	}

	return config, nil
}
