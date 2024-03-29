package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/chain/celo"
	"github.com/keep-network/keep-common/pkg/chain/ethereum"
	"github.com/keep-network/keep-core/pkg/net/libp2p"
	"github.com/keep-network/keep-ecdsa/pkg/client"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss"
	"github.com/keep-network/keep-ecdsa/pkg/extensions/tbtc"
)

// PasswordEnvVariable environment variable name for ethereum key password.
const PasswordEnvVariable = "KEEP_ETHEREUM_PASSWORD" // #nosec G101 -- it's just env variable name

// Config is the top level config structure.
type Config struct {
	Ethereum               ethereum.Config
	Celo                   celo.Config
	SanctionedApplications SanctionedApplications
	Storage                Storage
	LibP2P                 libp2p.Config
	Client                 client.Config
	TSS                    tss.Config
	Metrics                Metrics
	Diagnostics            Diagnostics
	Extensions             Extensions
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

// Metrics stores meta-info about metrics.
type Metrics struct {
	Port                int
	NetworkMetricsTick  int
	EthereumMetricsTick int
	ClientMetricsTick   int
}

// Diagnostics stores diagnostics-related configuration.
type Diagnostics struct {
	Port int
}

// Extensions stores app-specific extensions configuration.
type Extensions struct {
	TBTC tbtc.Config
}

// ReadConfig reads in the configuration file in .toml format. Chain key file
// password is expected to be provided as environment variable.
func ReadConfig(filePath string) (*Config, error) {
	config := &Config{}
	if _, err := toml.DecodeFile(filePath, config); err != nil {
		return nil, fmt.Errorf("failed to decode file [%s]: [%v]", filePath, err)
	}

	password := os.Getenv(PasswordEnvVariable)

	config.Ethereum.Account.KeyFilePassword = password
	config.Celo.Account.KeyFilePassword = password

	return config, nil
}

// ReadEthereumConfig reads in the configuration file at `filePath` and returns
// its contained Ethereum config, or an error if something fails while reading
// the file.
//
// This is the same as invoking ReadConfig and reading the Ethereum property
// from the returned config, but is available for external functions that expect
// to interact solely with Ethereum and are therefore independent of the rest of
// the config structure.
func ReadEthereumConfig(filePath string) (ethereum.Config, error) {
	config, err := ReadConfig(filePath)
	if err != nil {
		return ethereum.Config{}, err
	}

	return config.Ethereum, nil
}

// ReadCeloConfig reads in the configuration file at `filePath` and returns
// its contained Celo config, or an error if something fails while reading
// the file.
//
// This is the same as invoking ReadConfig and reading the Celo property
// from the returned config, but is available for external functions that expect
// to interact solely with Celo and are therefore independent of the rest of
// the config structure.
func ReadCeloConfig(filePath string) (celo.Config, error) {
	config, err := ReadConfig(filePath)
	if err != nil {
		return celo.Config{}, err
	}

	return config.Celo, nil
}
