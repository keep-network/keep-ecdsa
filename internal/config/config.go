package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-tecdsa/pkg/chain/btc/blockcypher"
	"github.com/keep-network/keep-tecdsa/pkg/chain/btc/electrum"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/ethereum"
)

const passwordEnvVariable = "KEEP_ETHEREUM_PASSWORD"

// Config is the top level config structure.
type Config struct {
	Electrum               electrum.Config
	BlockCypher            blockcypher.Config
	Ethereum               ethereum.Config
	SanctionedApplications SanctionedApplications
	Storage                Storage
}

// SanctionedApplications contains addresses of applications approved by the
// operator.
type SanctionedApplications struct {
	AddressesStrings []string `toml:"Addresses"`
}

// Addresses returns list of sanctioned applications as a slice of ethereum addresses.
func (sa *SanctionedApplications) Addresses() ([]common.Address, error) {
	applicationsAddresses := make([]common.Address, len(sa.AddressesStrings))

	for i, application := range sa.AddressesStrings {
		if !common.IsHexAddress(application) {
			return applicationsAddresses, fmt.Errorf(
				"application address [%v] is not valid hex address",
				application,
			)
		}

		applicationsAddresses[i] = common.HexToAddress(application)
	}

	return applicationsAddresses, nil
}

// Storage stores meta-info about keeping data on disk
type Storage struct {
	DataDir string
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
