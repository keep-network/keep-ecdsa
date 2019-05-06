package config

import (
	"fmt"

	"github.com/keep-network/toml"
	"github.com/keep-network/keep-tecdsa/pkg/chain/electrum"
)

// Config is the top level config structure.
type Config struct {
	Electrum electrum.Config
}

// ReadConfig reads in the configuration file in .toml format.
func ReadConfig(filePath string) (*Config, error) {
	config := &Config{}
	if _, err := toml.DecodeFile(filePath, config); err != nil {
		return nil, fmt.Errorf("unable to decode .toml file [%s] error [%s]", filePath, err)
	}

	return config, nil
}
