package recovery

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
)

func ErrorContains(err error, expected string) bool {
	return strings.Contains(err.Error(), expected)
}

var resolveAddressData = map[string]struct {
	beneficiaryAddress string
	usedIndexes        []uint32
	chainParams        *chaincfg.Params
	expectedAddress    string
}{
	"BIP44: xpub at m/44'/0'/0'/0/0": {
		"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1",
		[]uint32{},
		&chaincfg.MainNetParams,
		"1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx",
	},
	"BIP44: xpub at m/44'/0'/0'/0/4": {
		"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1",
		[]uint32{3},
		&chaincfg.MainNetParams,
		"1EEX8qZnTw1thadyxsueV748v3Y6tTMccc",
	},
	// P2PKH
	"Standard mainnet P2PKH btc address": {
		"1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx",
		[]uint32{},
		&chaincfg.MainNetParams,
		"1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx",
	},
	"Standard testnet P2PKH btc address": {
		"mkHS9ne12qx9pS9VojpwU5xtRd4T7X7ZUt",
		[]uint32{},
		&chaincfg.TestNet3Params,
		"mkHS9ne12qx9pS9VojpwU5xtRd4T7X7ZUt",
	},
	// P2SH
	"Standard mainnet P2SH btc address": {
		"3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy",
		[]uint32{},
		&chaincfg.MainNetParams,
		"3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy",
	},
	"Standard testnet P2SH btc address": {
		"2NBFNJTktNa7GZusGbDbGKRZTxdK9VVez3n",
		[]uint32{},
		&chaincfg.TestNet3Params,
		"2NBFNJTktNa7GZusGbDbGKRZTxdK9VVez3n",
	},
	// SegWit
	"Standard mainnet Bech32 (segwit) P2WPKH btc address": {
		"bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
		[]uint32{},
		&chaincfg.MainNetParams,
		"bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
	},
	"Standard mainnet Bech32 (segwit) P2WPSH btc address": {
		"bc1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3qccfmv3",
		[]uint32{},
		&chaincfg.MainNetParams,
		"bc1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3qccfmv3",
	},
	"Standard testnet Bech32 (segwit) P2WPKH btc address": {
		"tb1qw508d6qejxtdg4y5r3zarvary0c5xw7kxpjzsx",
		[]uint32{},
		&chaincfg.TestNet3Params,
		"tb1qw508d6qejxtdg4y5r3zarvary0c5xw7kxpjzsx",
	},
	// P2PK - public keys
	"Mainnet P2PK compressed btc public key (0x02)": {
		"02192d74d0cb94344c9569c2e77901573d8d7903c3ebec3a957724895dca52c6b4",
		[]uint32{},
		&chaincfg.MainNetParams,
		"13CG6SJ3yHUXo4Cr2RY4THLLJrNFuG3gUg",
	},
	"Mainnet P2PK compressed btc public key (0x03)": {
		"03b0bd634234abbb1ba1e986e884185c61cf43e001f9137f23c2c409273eb16e65",
		[]uint32{},
		&chaincfg.MainNetParams,
		"15sHANNUBSh6nDp8XkDPmQcW6n3EFwmvE6",
	},
	"Mainnet P2PK uncompressed btc public key (0x04)": {
		"0411db93e1dcdb8a016b49840f8c53bc1eb68a382e97b1482ecad7b148a6909a5cb2e0eaddfb84ccf9744464f82e160bfa9b8b64f9d4c03f999b8643f656b412a3",
		[]uint32{},
		&chaincfg.MainNetParams,
		"12cbQLTFMXRnSzktFkuoG3eHoMeFtpTu3S",
	},
	"Testnet P2PK compressed btc public key (0x02)": {
		"02192d74d0cb94344c9569c2e77901573d8d7903c3ebec3a957724895dca52c6b4",
		[]uint32{},
		&chaincfg.TestNet3Params,
		"mhiDPVP2nJunaAgTjzWSHCYfAqxxrxzjmo",
	},
}

func TestResolveAddress(t *testing.T) {
	for testName, testData := range resolveAddressData {
		t.Run(testName, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "example")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)
			dis, err := NewDerivationIndexStorage(dir)
			if err != nil {
				t.Fatal(err)
			}
			for _, usedIndex := range testData.usedIndexes {
				dis.save(testData.beneficiaryAddress, usedIndex)
			}

			handle := newMockBitcoinHandle()

			resolvedAddress, err := ResolveAddress(
				testData.beneficiaryAddress,
				dis,
				testData.chainParams,
				handle,
			)
			if err != nil {
				t.Fatal(err)
			}
			if resolvedAddress != testData.expectedAddress {
				t.Errorf(
					"the resolved address does not match\nexpected: %s\nactual:   %s",
					testData.expectedAddress,
					resolvedAddress,
				)
			}
		})
	}
}

var resolveAddressExpectedFailureData = map[string]struct {
	extendedAddress string
	chainParams     *chaincfg.Params
	failure         string
}{
	"WIF": {
		"5Hwgr3u458GLafKBgxtssHSPqJnYoGrSzgQsPwLFhLNYskDPyyA",
		&chaincfg.MainNetParams,
		"the provided serialized extended key length is invalid",
	},
	"empty string": {
		"",
		&chaincfg.MainNetParams,
		"insufficient length for public key",
	},
	"BIP32 private key": {
		"xprv9s21ZrQH143K24Mfq5zL5MhWK9hUhhGbd45hLXo2Pq2oqzMMo63oStZzF93Y5wvzdUayhgkkFoicQZcP3y52uPPxFnfoLZB21Teqt1VvEHx",
		&chaincfg.MainNetParams,
		"unusable seed",
	},
	"complete nonsense": {
		"lorem ipsum dolor sit amet, consec",
		&chaincfg.MainNetParams,
		"the provided serialized extended key length is invalid",
	},
}

func TestResolveBeneficiaryAddress_ExpectedFailure(t *testing.T) {
	for testName, testData := range resolveAddressExpectedFailureData {
		t.Run(testName, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "example")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)
			dis, err := NewDerivationIndexStorage(dir)
			if err != nil {
				t.Fatal(err)
			}
			_, err = ResolveAddress(
				testData.extendedAddress,
				dis,
				testData.chainParams,
				nil,
			)
			if err == nil {
				t.Errorf("no error found\nexpected: %s", testData.failure)
			} else if !ErrorContains(err, testData.failure) {
				t.Errorf(
					"unexpected error message\nexpected: %s\nactual:   %s",
					testData.failure,
					err.Error(),
				)
			}
		})
	}
}
