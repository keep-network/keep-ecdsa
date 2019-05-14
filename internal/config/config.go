package config

import (
	"fmt"

	"github.com/keep-network/keep-tecdsa/pkg/btc/chain/blockcypher"
	"github.com/keep-network/keep-tecdsa/pkg/btc/chain/electrum"
	"github.com/BurntSushi/toml"
)

// Config is the top level config structure.
type Config struct {
	Electrum    electrum.Config
	BlockCypher blockcypher.Config
}

// ReadConfig reads in the configuration file in .toml format.
func ReadConfig(filePath string) (*Config, error) {
	config := &Config{}
	if _, err := toml.DecodeFile(filePath, config); err != nil {
		return nil, fmt.Errorf("unable to decode .toml file [%s] error [%s]", filePath, err)
	}

	return config, nil
}
