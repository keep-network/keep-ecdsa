package bitcoin

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
)

// Config stores configuration related to recovering BTC from a closed keep.
type Config struct {
	BeneficiaryAddress string
	MaxFeePerVByte     int32
	ChainName          string
}

// ChainParams parses the net param name into the associated chaincfg.Params
func (c Config) ChainParams() (*chaincfg.Params, error) {
	switch c.ChainName {
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
		return nil, fmt.Errorf("unable to find chaincfg param for name: [%s]", c.ChainName)
	}
}
