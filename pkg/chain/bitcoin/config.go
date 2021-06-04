package bitcoin

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
)

// Config stores configuration related to recovering BTC from a closed keep.
type Config struct {
	BeneficiaryAddress string
	MaxFeePerVByte     int32
	BitcoinChainName   string
	ElectrsURL         *string
}

// Validate returns nil if the configuration is suitable for bitcoin recovery,
// and an error detailing what went wrong if not.
func (c Config) Validate() error {
	if c.BeneficiaryAddress == "" {
		return fmt.Errorf("a bitcoin address or extended public key (*pub) is required; configure one at [Extensions.TBTC.Bitcoin.BeneficiaryAddress]")
	}
	_, err := c.ChainParams()
	if err != nil {
		return fmt.Errorf("a valid chain name is required; choose between [mainnet, regtest, simnet, testnet3] and configure it at [Extensions.TBTC.Bitcoin.BitcoinChainName]: [%v]", err)
	}
	return nil
}

// ChainParams parses the net param name into the associated chaincfg.Params
func (c Config) ChainParams() (*chaincfg.Params, error) {
	switch c.BitcoinChainName {
	case "mainnet", "":
		// If no chain name is provided, use the main net
		return &chaincfg.MainNetParams, nil
	case "regtest":
		return &chaincfg.RegressionNetParams, nil
	case "simnet":
		return &chaincfg.SimNetParams, nil
	case "testnet3":
		return &chaincfg.TestNet3Params, nil
	default:
		return nil, fmt.Errorf("unable to find chaincfg param for name: [%s]", c.BitcoinChainName)
	}
}

// ElectrsURLWithDefault dereferences ElectrsURL in the following way: if there
// is a configured value, use it. Otherwise, default to
// https://blockstream.info/api/. This allows us to add bitcoin connection
// functionality to nodes that haven't made config changes yet while also
// letting a user connect to the node of their choice.
func (c Config) ElectrsURLWithDefault() string {
	if c.ElectrsURL == nil {
		return "https://blockstream.info/api/"
	}
	return *c.ElectrsURL
}
