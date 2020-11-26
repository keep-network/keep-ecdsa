package config

import (
	"math/big"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestReadConfig(t *testing.T) {
	err := os.Setenv("KEEP_ETHEREUM_PASSWORD", "not-my-password")
	if err != nil {
		t.Fatal(err)
	}

	filepath := "../internal/testdata/config.toml"
	cfg, err := ReadConfig(filepath)
	if err != nil {
		t.Fatalf(
			"failed to read test config: [%v]",
			err,
		)
	}

	var configReadTests = map[string]struct {
		readValueFunc func(*Config) interface{}
		expectedValue interface{}
	}{
		"Ethereum.URL": {
			readValueFunc: func(c *Config) interface{} { return c.Ethereum.URL },
			expectedValue: "ws://192.168.0.158:8546",
		},
		"Ethereum.URLRPC": {
			readValueFunc: func(c *Config) interface{} { return c.Ethereum.URLRPC },
			expectedValue: "http://192.168.0.158:8545",
		},
		"Ethereum.MaxGasPrice": {
			readValueFunc: func(c *Config) interface{} { return c.Ethereum.MaxGasPrice.Int },
			expectedValue: big.NewInt(140000000000),
		},
		"Ethereum.BalanceAlertThreshold": {
			readValueFunc: func(c *Config) interface{} { return c.Ethereum.BalanceAlertThreshold.Int },
			expectedValue: big.NewInt(2500000000000000000),
		},
		"Ethereum.ContractAddresses": {
			readValueFunc: func(c *Config) interface{} { return c.Ethereum.ContractAddresses },
			expectedValue: map[string]string{
				"BondedECDSAKeepFactory": "0x2BBE98119100D664eb6dEe5b8DB978aEEeAf42D6",
			},
		},
		"SanctionedApplications": {
			readValueFunc: func(c *Config) interface{} { return c.SanctionedApplications.AddressesStrings },
			expectedValue: []string{
				"0x15095EA15759f4C7d09cA2fcEd179527487ae81b",
				"0xda4c869B9073deac021344fd592c1BB0DC6Fc9a5",
			},
		},
		"Storage.DataDir": {
			readValueFunc: func(c *Config) interface{} { return c.Storage.DataDir },
			expectedValue: "/my/secure/location",
		},
		"LibP2P.Port": {
			readValueFunc: func(c *Config) interface{} { return c.LibP2P.Port },
			expectedValue: 27001,
		},
		"LibP2P.Peers": {
			readValueFunc: func(c *Config) interface{} { return c.LibP2P.Peers },
			expectedValue: []string{"/ip4/127.0.0.1/tcp/27001/ipfs/12D3KooWKRyzVWW6ChFjQjK4miCty85Niy49tpPV95XdKu1BcvMA"},
		},
		"Client.AwaitingKeyGenerationLookback": {
			readValueFunc: func(c *Config) interface{} { return c.Client.GetAwaitingKeyGenerationLookback() },
			expectedValue: time.Duration(172800000000000),
		},
		"Client.KeyGenerationTimeout": {
			readValueFunc: func(c *Config) interface{} { return c.Client.GetKeyGenerationTimeout() },
			expectedValue: time.Duration(6300000000000),
		},
		"Client.SigningTimeout": {
			readValueFunc: func(c *Config) interface{} { return c.Client.GetSigningTimeout() },
			expectedValue: time.Duration(12600000000000),
		},
		"TSS.PreParamsGenerationTimeout": {
			readValueFunc: func(c *Config) interface{} { return c.TSS.GetPreParamsGenerationTimeout() },
			expectedValue: time.Duration(397000000000),
		},
		"TSS.PreParamsTargetPoolSize": {
			readValueFunc: func(c *Config) interface{} { return c.TSS.GetPreParamsTargetPoolSize() },
			expectedValue: 36,
		},
		"Extensions.TBTC.TBTCSystem": {
			readValueFunc: func(c *Config) interface{} { return c.Extensions.TBTC.TBTCSystem },
			expectedValue: "0xa4888eDD97A5a3A739B4E0807C71817c8a418273",
		},
	}

	for testName, test := range configReadTests {
		t.Run(testName, func(t *testing.T) {
			expected := test.expectedValue
			actual := test.readValueFunc(cfg)
			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("\nexpected: %s\nactual:   %s", expected, actual)
			}
		})
	}

}
