package tss

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
)

func TestValidateValidateReceivedBtcAddress(t *testing.T) {
	var validateAddressData = map[string]struct {
		beneficiaryAddress string
		chainParams        *chaincfg.Params
	}{
		"Mainnet P2PKH btc address": {
			"1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx",
			&chaincfg.MainNetParams,
		},
		"Mainnet P2SH btc address": {
			"3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy",
			&chaincfg.MainNetParams,
		},
		"Mainnet Bech32 btc address": {
			"bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
			&chaincfg.MainNetParams,
		},
		"Testnet btc address": {
			"mkHS9ne12qx9pS9VojpwU5xtRd4T7X7ZUt",
			&chaincfg.TestNet3Params,
		},
		"Regression Network btc address": {
			"bcrt1qlmyyz6klzk6ckv7lqy65k26763xdp6y4dpn9he",
			&chaincfg.RegressionNetParams,
		},
		"Mainnet public key hash": {
			"17VZNX1SN5NtKa8UQFxwQbFeFc3iqRYhem",
			&chaincfg.MainNetParams,
		},
		"Mainnet script hash": {
			"3EktnHQD7RiAE6uzMj2ZifT9YgRrkSgzQX",
			&chaincfg.MainNetParams,
		},
		"Testnet public key hash": {
			"mipcBbFg9gMiCh81Kj8tqqdgoZub1ZJRfn",
			&chaincfg.TestNet3Params,
		},
		"Testnet script hash": {
			"2MzQwSSnBHWHqSAqtTVQ6v47XtaisrJa1Vc",
			&chaincfg.TestNet3Params,
		},
		"public key": {
			"03b0bd634234abbb1ba1e986e884185c61cf43e001f9137f23c2c409273eb16e65",
			&chaincfg.MainNetParams,
		},
	}
	for testName, testData := range validateAddressData {
		t.Run(testName, func(t *testing.T) {
			err := ValidateReceivedBtcAddress(testData.beneficiaryAddress, testData.chainParams)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestValidateReceivedBtcAddress_ExpectedFailures(t *testing.T) {
	var testData = map[string]struct {
		beneficiaryAddress string
		chainParams        *chaincfg.Params
		err                string
	}{
		"nonsense address": {
			"banana123",
			&chaincfg.MainNetParams,
			"failed to decode address [banana123] for chain [mainnet]",
		},
		"empty string": {
			"",
			&chaincfg.RegressionNetParams,
			"failed to decode address [] for chain [regtest]",
		},
		"mainnet private key": {
			"5Hwgr3u458GLafKBgxtssHSPqJnYoGrSzgQsPwLFhLNYskDPyyA",
			&chaincfg.MainNetParams,
			"failed to decode address [5Hwgr3u458GLafKBgxtssHSPqJnYoGrSzgQsPwLFhLNYskDPyyA] for chain [mainnet]",
		},
		"testnet public key against mainnet": {
			"mkHS9ne12qx9pS9VojpwU5xtRd4T7X7ZUt",
			&chaincfg.MainNetParams,
			"failed to decode address [mkHS9ne12qx9pS9VojpwU5xtRd4T7X7ZUt] for chain [mainnet]",
		},
		"mainnet public key against testnet": {
			"1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx",
			&chaincfg.TestNet3Params,
			"failed to decode address [1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx] for chain [testnet3]",
		},
		"mainnet bech32 address against testnet": {
			"bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
			&chaincfg.TestNet3Params,
			"address [bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq] is not a valid btc address for chain [testnet3]",
		},
		"BIP44: xpub": {
			"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1",
			&chaincfg.MainNetParams,
			"failed to decode address [xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1] for chain [mainnet]",
		},
	}
	for testName, testData := range testData {
		t.Run(testName, func(t *testing.T) {
			err := ValidateReceivedBtcAddress(testData.beneficiaryAddress, testData.chainParams)
			if err == nil || err.Error() != testData.err {
				t.Errorf(
					"unexpected error message\nexpected: %s\nactual:   %s",
					testData.err,
					err,
				)
			}
		})
	}
}
