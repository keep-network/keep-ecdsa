package config

import (
	"fmt"
	"math/big"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/btcsuite/btcd/chaincfg"
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
				"TBTCSystem":             "0xda4c869B9073deac021344fd592c1BB0DC6Fc9a5",
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
		"Extensions.TBTC.LiquidationRecoveryTimeout": {
			readValueFunc: func(c *Config) interface{} { return c.Extensions.TBTC.GetLiquidationRecoveryTimeout() },
			expectedValue: time.Duration(49 * 60 * 60 * 1000000000), // 49 hours in nanoseconds
		},
		"Extensions.TBTC.Bitcoin.ElectrsURL": {
			readValueFunc: func(c *Config) interface{} { return *c.Extensions.TBTC.Bitcoin.ElectrsURL },
			expectedValue: "example.com",
		},
		"Extensions.TBTC.Bitcoin.ElectrsURLWithDefault()": {
			readValueFunc: func(c *Config) interface{} { return c.Extensions.TBTC.Bitcoin.ElectrsURLWithDefault() },
			expectedValue: "example.com",
		},
		"Extensions.TBTC.Bitcoin.BeneficiaryAddress": {
			readValueFunc: func(c *Config) interface{} { return c.Extensions.TBTC.Bitcoin.BeneficiaryAddress },
			expectedValue: "bcrt1q0umle4fe6penqqyzuwsysqezwwptuyqa82jas4",
		},
		"Extensions.TBTC.Bitcoin.MaxFeePerVByte": {
			readValueFunc: func(c *Config) interface{} { return c.Extensions.TBTC.Bitcoin.MaxFeePerVByte },
			expectedValue: int32(73),
		},
		"Extensions.TBTC.Bitcoin.BitcoinChainName": {
			readValueFunc: func(c *Config) interface{} { return c.Extensions.TBTC.Bitcoin.BitcoinChainName },
			expectedValue: "mainnet",
		},
		"Extensions.TBTC.Bitcoin.ChainParams()": {
			readValueFunc: func(c *Config) interface{} {
				params, _ := c.Extensions.TBTC.Bitcoin.ChainParams()
				return *params
			},
			expectedValue: chaincfg.MainNetParams,
		},
		"Extensions.TBTC.Bitcoin.Validate() error": {
			readValueFunc: func(c *Config) interface{} {
				_, err := c.Extensions.TBTC.Bitcoin.Validate()
				return err
			},
			expectedValue: nil,
		},
		"Extensions.TBTC.Bitcoin.Validate() isUnconfigured": {
			readValueFunc: func(c *Config) interface{} {
				isUnconfigured, _ := c.Extensions.TBTC.Bitcoin.Validate()
				return isUnconfigured
			},
			expectedValue: true,
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

func TestParseChainParams(t *testing.T) {
	var parseChainParamTests = map[string]struct {
		chainName   []string
		chainParams chaincfg.Params
	}{
		"main network": {
			[]string{"mainnet"},
			chaincfg.MainNetParams,
		},
		"regtest": {
			[]string{"regtest"},
			chaincfg.RegressionNetParams,
		},
		"simnet": {
			[]string{"simnet"},
			chaincfg.SimNetParams,
		},
		"testnet3": {
			[]string{"testnet3"},
			chaincfg.TestNet3Params,
		},
		"undefined": {
			[]string{},
			chaincfg.MainNetParams,
		},
		"empty": {
			[]string{""},
			chaincfg.MainNetParams,
		},
	}
	for testName, testData := range parseChainParamTests {
		t.Run(testName, func(t *testing.T) {
			// Use a string builder and a single-value list to represent optionality.
			// If the list has a chain name, we define it in the config, otherwise,
			// we have an empty `[Extensions.TBTC.Bitcoin]` section.
			var b strings.Builder
			fmt.Fprint(&b, "[Extensions.TBTC.Bitcoin]")
			for _, name := range testData.chainName {
				fmt.Fprintf(&b, "\nBitcoinChainName=\"%s\"", name)
			}
			config := &Config{}
			if _, err := toml.Decode(b.String(), config); err != nil {
				t.Fatal(err)
			}
			chainParams, err := config.Extensions.TBTC.Bitcoin.ChainParams()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(*chainParams, testData.chainParams) {
				t.Errorf("unexpected net params\nexpected: %v\nactual:   %v", testData.chainParams, chainParams)
			}
		})
	}
}

func TestParseChainParams_ExpectedFailure(t *testing.T) {
	configString := fmt.Sprintf("[Extensions.TBTC.Bitcoin]\nBitcoinChainName=\"%s\"", "bleeble blabble")
	config := &Config{}
	if _, err := toml.Decode(configString, config); err != nil {
		t.Fatal(err)
	}
	_, err := config.Extensions.TBTC.Bitcoin.ChainParams()
	expectedError := "unable to find chaincfg param for name: [bleeble blabble]"
	if err == nil {
		t.Fatalf("expecting an error but found none")
	}
	if expectedError != err.Error() {
		t.Errorf(
			"unexpected error\nexpected: %s\nactual:   %v",
			expectedError,
			err,
		)
	}
}

func TestDefaultVbyteFee(t *testing.T) {
	config := &Config{}
	if _, err := toml.Decode("[Extensions.TBTC.Bitcoin]", config); err != nil {
		t.Fatal(err)
	}
	fee := config.Extensions.TBTC.Bitcoin.MaxFeePerVByte
	if fee != 0 {
		t.Errorf("unexpected default fee\nexpected: 0\nactual:   %d", fee)
	}
}

func TestElectrsURLWithDefault(t *testing.T) {
	var electrsURLTests = map[string]struct {
		url         []string
		expectedURL string
	}{
		"blockstream": {
			[]string{"https://blockstream.info/api/"},
			"https://blockstream.info/api/",
		},
		"localhost": {
			[]string{"localhost:8080/api/"},
			"localhost:8080/api/",
		},
		"nonsense": {
			[]string{"bleeble blabble"},
			"bleeble blabble",
		},
		"undefined": {
			[]string{},
			"https://blockstream.info/api/",
		},
		"empty": {
			[]string{""},
			"",
		},
	}
	for testName, testData := range electrsURLTests {
		t.Run(testName, func(t *testing.T) {
			// Use a string builder and a single-value list to represent optionality.
			// If the list has a url, we define it in the config, otherwise, we have
			// an empty `[Extensions.TBTC]` section.
			var b strings.Builder
			fmt.Fprint(&b, "[Extensions.TBTC.Bitcoin]")
			for _, url := range testData.url {
				fmt.Fprintf(&b, "\nElectrsURL=\"%s\"", url)
			}
			config := &Config{}
			if _, err := toml.Decode(b.String(), config); err != nil {
				t.Fatal(err)
			}
			url := config.Extensions.TBTC.Bitcoin.ElectrsURLWithDefault()
			if url != testData.expectedURL {
				t.Errorf("unexpected url\nexpected: %s\nactual:   %s", testData.expectedURL, url)
			}
		})
	}
}
