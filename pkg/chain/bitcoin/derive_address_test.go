package bitcoin

import (
	"strings"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

// These tests use https://iancoleman.io/bip39/ with the bip39 mnemonic: loyal
// chuckle trade magnet tobacco jungle craft cram reduce climb size flip tongue
// tornado height
var deriveAddressTestData = map[string]struct {
	extendedAddress string
	addressIndex    int
	chainParams     *chaincfg.Params
	expectedAddress string
}{
	// mainnet
	"BIP44: xpub at m/44'/0'/0'/0/0": {
		"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1",
		0,
		&chaincfg.MainNetParams,
		"1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx",
	},
	"BIP44: xpub at m/44'/0'/0'/0/4": {
		"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1",
		4,
		&chaincfg.MainNetParams,
		"1EEX8qZnTw1thadyxsueV748v3Y6tTMccc",
	},
	"BIP44: xpub at m/44'/0'/6'/0/2": {
		"xpub6DsnSRbofRYrTZLW3cY5AZJYJ9P3np2ydf8WtTYaSBUzCjbRv3xb8j5v97pysDqNZWoEwWjadpVRBYDwZdApxieyPYDosYLP8VtTVZjmLRR",
		2,
		&chaincfg.MainNetParams,
		"1EZuKz6RrJ6XmBPvFwJiEcREpaEVhUVAt5",
	},
	"BIP49: ypub at m/49'/0'/0'/0/0": {
		"ypub6Xxan668aiJqvh4SVfd7EzqjWvf36gWufTkhWHv3gaxnBh44HpkTi2TTkm1u136qjUxk7F3jGzoyfrGpHvALMgJgbF4WNXpoPu3QYrqogMK",
		0,
		&chaincfg.MainNetParams,
		"3Aobe26f7QzKN73mvYQVbt1KLrCU1CgQpD",
	},
	"BIP49: ypub at m/49'/0'/9'/0/11": {
		"ypub6Xxan668aiJrJ2sq1TtRJmdJSZ8DyA9569rPCay5R14zcRdXUd7RcjBhc2jzhuT2GK2aqGTNQkd4N5riF5DVnbWav3qpQXT57cA7PaL1o1J",
		11,
		&chaincfg.MainNetParams,
		"3M9z86KGrNER18mKKhwxeeNxzmeoD9iUpt",
	},
	"BIP84: zpub at m/84'/0'/0'/0/0": {
		"zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9",
		0,
		&chaincfg.MainNetParams,
		"bc1q46uejlhm9vkswfcqs9plvujzzmqjvtfda3mra6",
	},
	"BIP84: zpub at m/84'/0'/0'/0/8": {
		"zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9",
		8,
		&chaincfg.MainNetParams,
		"bc1quq0vrufxy05ypk45xmu3hpk6qsmlhr5vr3n8kz",
	},
	"BIP84: ypub at m/84'/0'/72'/0/12": {
		"ypub6ZpieGfpesfH3KqGr4zZPETidCze6RzeNMz7FLnSPgABwyQNZZmpA4tpUYFn53xtHkHXaoGviseJJcFhSn3Kw9sgzsiSnP5xEqp6Z2Yy4ZH",
		12,
		&chaincfg.MainNetParams,
		"32n4JF1ytaPfw4951nvN8gvNmAnRuxMvMb",
	},
	"BIP141: P2WPKH nested in P2SH ypub at m/0/0/0/0/0 extrapolated from m/0": {
		"ypub6TMciWL8Pv4Rk41sLR1Z8ay9beZPMDyrV3T7tbb4Vtw3Vaf3uxWmug1hp5uEry9CbR6448YJEzUopCT8PSgKMPZVFVZKDc2kvQC8xHqdtZa",
		0,
		&chaincfg.MainNetParams,
		"398r9poPaoKJ7vHkaVzNVsXBGRB3mFMXEK",
	},
	"BIP141: P2WPKH nested in P2SH ypub at m/6'/0/0/0/7 extrapolated from m/6'": {
		"ypub6TMciWLGjabQBSQtDkD3vG3KZdXDZ89ySQrsjuviw1M6wr7uF5fnbimqae2zhJPVminGn29Q6jHCZS9RFBCnAikDsWQgY57J9hLmptE2oK8",
		7,
		&chaincfg.MainNetParams,
		"33F67PgGyFD73YmDg7JYbwEqufsB89vvpc",
	},
	"BIP141: P2WPKH nested in P2SH ypub at m/6'/0/0/0/0 extrapolated from m/6'/0": {
		"ypub6VGWAW57V8o1eaVnrvNPuKb7xfvP6X4hxb3vxGGQp9oyCKoZCFvbvLDYjy36RE1immwe5RSGGFiQULB65v9Zw1Ej2TbPR6CLTGkzhxMSQ4q",
		0,
		&chaincfg.MainNetParams,
		"3Qt1E64dYpUA1ovvYYfZYYeoCazjDTLww3",
	},
	"BIP141: P2WPKH nested in P2SH ypub at m/6'/4'/0/0/1 extrapolated from m/6'/4'": {
		"ypub6VGWAW5FpoKyziaLoi5RJnL1ezT4Xb5BwZH1HYmArsESpvs4rgP1fM8hPxQA6qEnxXVM5zFbzzSWKuuvB3QUFSbmFppU5hyQHyojgvf8S3f",
		1,
		&chaincfg.MainNetParams,
		"3PMemKpygPQDyzqsjdrnqMswRhNjcka1Wu",
	},
	"BIP141: P2WPKH nested in P2SH ypub at m/6'/4'/9'/0/11 extrapolated from m/6'/4'/9'": {
		"ypub6Xgwmha4MdU4UAfH9oXWtTmCvZxnqZXkp1ReiWsKeHvKid6QEhjA2CUDQcNCYpiExbPkZhr6DHtTUda3ZQp4KM7NfqVJWs3YFqHrJLECa1k",
		11,
		&chaincfg.MainNetParams,
		"382CXgbMHT4gzUiKQ2EHsJgESrBjv6bik8",
	},
	"BIP141: P2WPKH nested in P2SH ypub at m/6'/4'/9'/0/4": {
		"ypub6Z7s8wJuKsxjd16oe85WH1uSbcbbCXuMFEhPMgcf7jQqNhQbT9jE52XVu1eBe18q2J3LwnDd54ufL2jNvidjfCkbd34aVwLtYdztLUqucwR",
		4,
		&chaincfg.MainNetParams,
		"3FQZbk6HBX72j2yyT5L8hWTymaJgwvb9u9",
	},
	"BIP141: P2WPKH zpub at m/0/0/0/0/0 extrapolated from m/0": {
		"zpub6nBt2B13YbbubMCzAmoBLg4emchqHqyMQ9yLfzUwsuJvYgUHAcgLXjfqqHrprso814Croc8rheqMhV4h796L9dF67qFjoWrFC8FnLsVHknB",
		0,
		&chaincfg.MainNetParams,
		"bc1q8dnmvgj4jsvafe0wuwdm89aua2405jp7jp2zhn",
	},
	"BIP141: P2WPKH zpub at m/6'/0/0/0/7 extrapolated from m/6'": {
		"zpub6nBt2B1BtG8t2jc146zg8M8pjbffVk9UMXP6XJpcK1iyzww8VjqMDnRybqzahD3RBMu5XVjxZPdkSikyxscnxxRpjr777yvnRRQRDVAG58G",
		7,
		&chaincfg.MainNetParams,
		"bc1q4nrgadrcxz4tcqke9eh7f6zn8lxg8lz49d8v6k",
	},
	"BIP141: P2WPKH zpub at m/6'/0/0/0/0 extrapolated from m/6'/0": {
		"zpub6p6mUAk2dpLVVsguhHA27Qgd8e4q394Csha9jfAJCABrFRcnSv6AYPsgmAzgR8feBR4Spu2piv4xMcneocZajEvKtoHp111pizpe6aAEqfp",
		0,
		&chaincfg.MainNetParams,
		"bc1q9wwwcgcl2lw74quetxan4j6vhluyvyhy3dwt5l",
	},
	"BIP141: P2WPKH zpub at m/6'/4'/0/0/1 extrapolated from m/6'/4'": {
		"zpub6p6mUAkAyUsTr1mTe4s3WsRWpxbWUD4grfoE4wf4EscKt2gJ7LYaHQnqRAMk6jtiNAc9qTrATeo4DCXUtjpV3gHN8AWtfcntZhsP5UeStdx",
		1,
		&chaincfg.MainNetParams,
		"bc1qwsuszzk93puxlgs6l6f54r66dev882pm2760dv",
	},
	"BIP141: P2WPKH zpub at m/6'/4'/9'/0/11 extrapolated from m/6'/4'/9'": {
		"zpub6rXD5NEyWK1YKTrPzAK96Yri6Y7EnBXFj7wsVumD2JJCmiudVMtieG8MRpKnYjNANEWZKBSefxF1MvBcH7E57anyYBBj6ms2XZMVgsuPzzs",
		11,
		&chaincfg.MainNetParams,
		"bc1qsszrcep8whzqh93ksmmckn77eh9fl55s5dzjx6",
	},
	"BIP141: P2WPKH zpub at m/6'/4'/9'/0/4": {
		"zpub6sx8SbypUZWDUJHvUUs8V6zwmak399trAMDc95WYVjniRoDphotnh6BdvDbmdunkRwA9hFpBXjGDDKLweR3kTSSCVNm15rANpN4XixewDwG",
		4,
		&chaincfg.MainNetParams,
		"bc1q5l3j7e2s3dzg4vxquxldvzw4dwdlvvhdc9c4zh",
	},
	// testnet
	"BIP44: tpub at m/44'/1'/3'/0/4 for testnet": {
		"tpubDEXzoXkNdhoFeYrtS2BJKfok6LwH5PKkr5jPSMR6A2erw2yS3VgY5EoYdcKH24VPqeAgBTF6i82Ft9NG1iVjSQVAvFBfd2wkRQXF1W2Q8W1",
		4,
		&chaincfg.TestNet3Params,
		"ms5BTH2MNLgtp1RNJh4Fnsjdyu3ZuoV7nw",
	},
	"BIP49: ypub at m/49'/1'/9'/0/11 for testnet": {
		"upub5DnYQWgCDSGe6DAr6NZaD4jsh1Qz55jh7DeScVVyJJnHCp5BwyvM7Xm7S5r5n5ZYMJ1WrrM31i4kcsWwW2vxcJS1kfsuKgK9vME2z1cx6aX",
		11,
		&chaincfg.TestNet3Params,
		"2Mx8Mnno8NW8ugL9iYX2vvbiDa94kR8xnii",
	},
	"BIP141: vpub at m/44'/0'/0'/0/4 for testnet": {
		"vpub5Zx5difzitDBNPjrr9pTno6C44dJFd89naYzhyk9QWHFTpF7pJqnyAnADhbVrFYX7eCK8V2WBBVprxzJrSk15NsYHiB8CvV8h4JnXkU66as",
		4,
		&chaincfg.TestNet3Params,
		"tb1qjy5r90er70t2cexwpmmkf9hr4glxdx83jhpwfv",
	},
	// regtest
	"BIP44: tpub at m/44'/1'/3'/0/4 for regtest": {
		"tpubDEXzoXkNdhoFeYrtS2BJKfok6LwH5PKkr5jPSMR6A2erw2yS3VgY5EoYdcKH24VPqeAgBTF6i82Ft9NG1iVjSQVAvFBfd2wkRQXF1W2Q8W1",
		4,
		&chaincfg.RegressionNetParams,
		"ms5BTH2MNLgtp1RNJh4Fnsjdyu3ZuoV7nw",
	},
	"BIP49: ypub at m/49'/1'/9'/0/11 for regtest": {
		"upub5DnYQWgCDSGe6DAr6NZaD4jsh1Qz55jh7DeScVVyJJnHCp5BwyvM7Xm7S5r5n5ZYMJ1WrrM31i4kcsWwW2vxcJS1kfsuKgK9vME2z1cx6aX",
		11,
		&chaincfg.RegressionNetParams,
		"2Mx8Mnno8NW8ugL9iYX2vvbiDa94kR8xnii",
	},
	"BIP141: vpub at m/44'/0'/0'/0/4 for regtest": {
		"vpub5Zx5difzitDBNPjrr9pTno6C44dJFd89naYzhyk9QWHFTpF7pJqnyAnADhbVrFYX7eCK8V2WBBVprxzJrSk15NsYHiB8CvV8h4JnXkU66as",
		4,
		&chaincfg.RegressionNetParams,
		"bcrt1qjy5r90er70t2cexwpmmkf9hr4glxdx83s7cr79",
	},
}

func TestDeriveAddress(t *testing.T) {
	for testName, testData := range deriveAddressTestData {
		t.Run(testName, func(t *testing.T) {
			address, err := DeriveAddress(testData.extendedAddress, uint32(testData.addressIndex), testData.chainParams)
			if err != nil {
				t.Fatalf("failed to derive address: %s", err)
			}

			if address != testData.expectedAddress {
				t.Errorf(
					"unexpected derived address\nexpected: %s\nactual:   %s",
					testData.expectedAddress,
					address,
				)
			}

			// Validate if derived address is valid for bitcoin network.
			decodedAddress, err := btcutil.DecodeAddress(address, testData.chainParams)
			if err != nil {
				t.Fatalf("failed to decode address: %s", err)
			}
			if !decodedAddress.IsForNet(testData.chainParams) {
				t.Errorf(
					"address %s is not valid for network %s",
					address,
					testData.chainParams.Name,
				)
			}
		})
	}
}

func TestDeriveAddress_ExpectedFailures(t *testing.T) {
	deriveAddressTestFailureData := map[string]struct {
		extendedAddress string
		addressIndex    int
		chainParams     *chaincfg.Params
		failure         string
	}{
		"BIP141 P2WPKH nested in P2SH ypub at m/6'/4'/9'/0/11'": {
			"ypub6Z7s8wJuKsxjd16oe85WH1uSbcbbCXuMFEhPMgcf7jQqNhQbT9jE52XVu1eBe18q2J3LwnDd54ufL2jNvidjfCkbd34aVwLtYdztLUqucwR",
			11 + 2147483648,
			&chaincfg.MainNetParams,
			"error deriving requested address index /0/2147483659 from extended key: [cannot derive a hardened key from a public key]",
		},
		"BIP141 P2WPKH nested in P2SH ypub at m/6'/4'/9'/0/11' with a private key": {
			"yprvAL8WjRn1VWQSQX2LY6YVusxi3am6o5BVt1mnZJD3ZPsrVu5SucQyXED23ikCvDeeFHTMeX9q5n5MHNTLWQvCSm3KWnA3KdyZuDXncTn2VW5",
			11 + 2147483648,
			&chaincfg.MainNetParams,
			"unsupported public key format [yprv]",
		},
		"BIP141 ypub is too deep at m/0/0/0/0/0/0": {
			"ypub6bp11ZqNVMqm3C3eXAFGpEvKqNfEZ6Vhznd4Uo3S73RYTSFgmF7q9sWPoCFhLGVMSLqKZZpcpHoKgHNwStDuqQPnDfF13goQwS8qSFA6vnz",
			0,
			&chaincfg.MainNetParams,
			"extended public key is deeper than 4, depth: 5",
		},
		// public key descriptor and network type missmatch
		"ypub and testnet": {
			"ypub6Xxan668aiJqvh4SVfd7EzqjWvf36gWufTkhWHv3gaxnBh44HpkTi2TTkm1u136qjUxk7F3jGzoyfrGpHvALMgJgbF4WNXpoPu3QYrqogMK",
			0,
			&chaincfg.TestNet3Params,
			"public key descriptor [ypub] is invalid for network [testnet3]",
		},
		"zpub and regtest": {
			"zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9",
			0,
			&chaincfg.RegressionNetParams,
			"public key descriptor [zpub] is invalid for network [regtest]",
		},
		"upub and mainnet": {
			"upub5DnYQWgCDSGe6DAr6NZaD4jsh1Qz55jh7DeScVVyJJnHCp5BwyvM7Xm7S5r5n5ZYMJ1WrrM31i4kcsWwW2vxcJS1kfsuKgK9vME2z1cx6aX",
			0,
			&chaincfg.MainNetParams,
			"public key descriptor [upub] is invalid for network [mainnet]",
		},
		// Unsupported SLIP132 key generated with https://jlopp.github.io/xpub-converter.
		"SLIP132: Ypub unsupported prefix": {
			"Ypub6irfuKpa9fsDMGDpSL6655BYEihJK3CVyjQNRZBb4MoBoscy4E8jo9KPZTyNZUKjxx2iyq3rADCUo1tab9KHWARMRiAumwHoHdKH8qBNhaf",
			0,
			&chaincfg.MainNetParams,
			"unsupported public key format [Ypub]",
		},
	}

	for testName, testData := range deriveAddressTestFailureData {
		t.Run(testName, func(t *testing.T) {
			_, err := DeriveAddress(testData.extendedAddress, uint32(testData.addressIndex), testData.chainParams)
			if err == nil || err.Error() != testData.failure {
				t.Errorf(
					"unexpected error message\nexpected: %v\nactual:   %v",
					testData.failure,
					err,
				)
			}
		})
	}
}

func TestValidateAddressOrKey_ExtendedPublicKeys(t *testing.T) {
	for testName, testData := range deriveAddressTestData {
		t.Run(testName, func(t *testing.T) {
			err := ValidateAddressOrKey(testData.extendedAddress, testData.chainParams)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestValidateAddressOrKey_Addresses(t *testing.T) {
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
			err := ValidateAddressOrKey(testData.beneficiaryAddress, testData.chainParams)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestValidateAddressOrKey_ExpectedFailures(t *testing.T) {
	var testData = map[string]struct {
		beneficiaryAddress string
		chainParams        *chaincfg.Params
		err                string
	}{
		"nonsense address": {
			"banana123",
			&chaincfg.MainNetParams,
			"[banana123] is not a valid btc address or extended key using chain [mainnet]: address validation failed with [failed to decode address from [banana123] for chain [mainnet]] and derivation from extended key failed with [error parsing extended public key: [the provided serialized extended key length is invalid]]",
		},
		"empty string": {
			"",
			&chaincfg.RegressionNetParams,
			"[] is not a valid btc address or extended key using chain [regtest]: address validation failed with [failed to decode address from [] for chain [regtest]] and derivation from extended key failed with [error parsing extended public key: [the provided serialized extended key length is invalid]]",
		},
		"Mainnet private key": {
			"5Hwgr3u458GLafKBgxtssHSPqJnYoGrSzgQsPwLFhLNYskDPyyA",
			&chaincfg.MainNetParams,
			"[5Hwgr3u458GLafKBgxtssHSPqJnYoGrSzgQsPwLFhLNYskDPyyA] is not a valid btc address or extended key using chain [mainnet]: address validation failed with [failed to decode address from [5Hwgr3u458GLafKBgxtssHSPqJnYoGrSzgQsPwLFhLNYskDPyyA] for chain [mainnet]] and derivation from extended key failed with [error parsing extended public key: [the provided serialized extended key length is invalid]]",
		},
		"testnet private key": {
			"92Pg46rUhgTT7romnV7iGW6W1gbGdeezqdbJCzShkCsYNzyyNcc",
			&chaincfg.TestNet3Params,
			"[92Pg46rUhgTT7romnV7iGW6W1gbGdeezqdbJCzShkCsYNzyyNcc] is not a valid btc address or extended key using chain [testnet3]: address validation failed with [failed to decode address from [92Pg46rUhgTT7romnV7iGW6W1gbGdeezqdbJCzShkCsYNzyyNcc] for chain [testnet3]] and derivation from extended key failed with [error parsing extended public key: [the provided serialized extended key length is invalid]]",
		},
		"testnet public key against mainnet": {
			"mkHS9ne12qx9pS9VojpwU5xtRd4T7X7ZUt",
			&chaincfg.MainNetParams,
			"[mkHS9ne12qx9pS9VojpwU5xtRd4T7X7ZUt] is not a valid btc address or extended key using chain [mainnet]: address validation failed with [failed to decode address from [mkHS9ne12qx9pS9VojpwU5xtRd4T7X7ZUt] for chain [mainnet]] and derivation from extended key failed with [error parsing extended public key: [the provided serialized extended key length is invalid]]",
		},
		"mainnet public key against testnet": {
			"1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx",
			&chaincfg.TestNet3Params,
			"[1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx] is not a valid btc address or extended key using chain [testnet3]: address validation failed with [failed to decode address from [1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx] for chain [testnet3]] and derivation from extended key failed with [error parsing extended public key: [the provided serialized extended key length is invalid]]",
		},
		"bech32 address against testnet": {
			"bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
			&chaincfg.TestNet3Params,
			"[bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq] is not a valid btc address or extended key using chain [testnet3]: address validation failed with [address [bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq] is not a valid btc address for chain [testnet3]] and derivation from extended key failed with [error parsing extended public key: [the provided serialized extended key length is invalid]]"},
	}
	for testName, testData := range testData {
		t.Run(testName, func(t *testing.T) {
			err := ValidateAddressOrKey(testData.beneficiaryAddress, testData.chainParams)
			if err == nil {
				t.Errorf(
					"unexpected error message\nexpected: %s\nactual:   nil",
					testData.err,
				)
			} else if !ErrorContains(err, testData.err) {
				t.Errorf(
					"unexpected error message\nexpected: %s\nactual:   %s",
					testData.err,
					err.Error(),
				)
			}
		})
	}
}

func TestValidateAddress(t *testing.T) {
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
			err := ValidateAddress(testData.beneficiaryAddress, testData.chainParams)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestValidateAddress_ExpectedFailures(t *testing.T) {
	var testData = map[string]struct {
		beneficiaryAddress string
		chainParams        *chaincfg.Params
		err                string
	}{
		"nonsense address": {
			"banana123",
			&chaincfg.MainNetParams,
			"failed to decode address from [banana123] for chain [mainnet]",
		},
		"empty string": {
			"",
			&chaincfg.RegressionNetParams,
			"failed to decode address from [] for chain [regtest]",
		},
		"mainnet private key": {
			"5Hwgr3u458GLafKBgxtssHSPqJnYoGrSzgQsPwLFhLNYskDPyyA",
			&chaincfg.MainNetParams,
			"failed to decode address from [5Hwgr3u458GLafKBgxtssHSPqJnYoGrSzgQsPwLFhLNYskDPyyA] for chain [mainnet]",
		},
		"testnet public key against mainnet": {
			"mkHS9ne12qx9pS9VojpwU5xtRd4T7X7ZUt",
			&chaincfg.MainNetParams,
			"failed to decode address from [mkHS9ne12qx9pS9VojpwU5xtRd4T7X7ZUt] for chain [mainnet]",
		},
		"mainnet public key against testnet": {
			"1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx",
			&chaincfg.TestNet3Params,
			"failed to decode address from [1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx] for chain [testnet3]",
		},
		"mainnet bech32 address against testnet": {
			"bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
			&chaincfg.TestNet3Params,
			"address [bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq] is not a valid btc address for chain [testnet3]",
		},
		"BIP44: xpub": {
			"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1",
			&chaincfg.MainNetParams,
			"failed to decode address from [xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1] for chain [mainnet]",
		},
	}
	for testName, testData := range testData {
		t.Run(testName, func(t *testing.T) {
			err := ValidateAddress(testData.beneficiaryAddress, testData.chainParams)
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

func ErrorContains(err error, expected string) bool {
	return strings.Contains(err.Error(), expected)
}
