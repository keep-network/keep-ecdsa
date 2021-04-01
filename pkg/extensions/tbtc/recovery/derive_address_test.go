package recovery

import (
	"strings"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
)

// These tests use https://iancoleman.io/bip39/ with the bip39 mnemonic: loyal
// chuckle trade magnet tobacco jungle craft cram reduce climb size flip tongue
// tornado height
var deriveAddressTestData = map[string]struct {
	extendedAddress string
	addressIndex    int
	expectedAddress string
}{
	"BIP44: xpub at m/44'/0'/0'/0/0": {
		"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1",
		0,
		"1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx",
	},
	"BIP44: xpub at m/44'/0'/0'/0/4": {
		"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1",
		4,
		"1EEX8qZnTw1thadyxsueV748v3Y6tTMccc",
	},
	"BIP44: xpub at m/44'/0'/6'/0/2": {
		"xpub6DsnSRbofRYrTZLW3cY5AZJYJ9P3np2ydf8WtTYaSBUzCjbRv3xb8j5v97pysDqNZWoEwWjadpVRBYDwZdApxieyPYDosYLP8VtTVZjmLRR",
		2,
		"1EZuKz6RrJ6XmBPvFwJiEcREpaEVhUVAt5",
	},

	"BIP49: ypub at m/49'/0'/0'/0/0": {
		"ypub6Xxan668aiJqvh4SVfd7EzqjWvf36gWufTkhWHv3gaxnBh44HpkTi2TTkm1u136qjUxk7F3jGzoyfrGpHvALMgJgbF4WNXpoPu3QYrqogMK",
		0,
		"3Aobe26f7QzKN73mvYQVbt1KLrCU1CgQpD",
	},
	"BIP49: ypub at m/49'/0'/9'/0/11": {
		"ypub6Xxan668aiJrJ2sq1TtRJmdJSZ8DyA9569rPCay5R14zcRdXUd7RcjBhc2jzhuT2GK2aqGTNQkd4N5riF5DVnbWav3qpQXT57cA7PaL1o1J",
		11,
		"3M9z86KGrNER18mKKhwxeeNxzmeoD9iUpt",
	},

	"BIP84: zpub at m/84'/0'/0'/0/0": {
		"zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9",
		0,
		"bc1q46uejlhm9vkswfcqs9plvujzzmqjvtfda3mra6",
	},
	"BIP84: zpub at m/84'/0'/0'/0/8": {
		"zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9",
		8,
		"bc1quq0vrufxy05ypk45xmu3hpk6qsmlhr5vr3n8kz",
	},
	"BIP84: ypub at m/84'/0'/72'/0/12": {
		"ypub6ZpieGfpesfH3KqGr4zZPETidCze6RzeNMz7FLnSPgABwyQNZZmpA4tpUYFn53xtHkHXaoGviseJJcFhSn3Kw9sgzsiSnP5xEqp6Z2Yy4ZH",
		12,
		"32n4JF1ytaPfw4951nvN8gvNmAnRuxMvMb",
	},

	"BIP141: P2WPKH nested in P2SH ypub at m/0/0/0/0/0 extrapolated from m/0": {
		"ypub6TMciWL8Pv4Rk41sLR1Z8ay9beZPMDyrV3T7tbb4Vtw3Vaf3uxWmug1hp5uEry9CbR6448YJEzUopCT8PSgKMPZVFVZKDc2kvQC8xHqdtZa",
		0,
		"398r9poPaoKJ7vHkaVzNVsXBGRB3mFMXEK",
	},
	"BIP141: P2WPKH nested in P2SH ypub at m/6'/0/0/0/7 extrapolated from m/6'": {
		"ypub6TMciWLGjabQBSQtDkD3vG3KZdXDZ89ySQrsjuviw1M6wr7uF5fnbimqae2zhJPVminGn29Q6jHCZS9RFBCnAikDsWQgY57J9hLmptE2oK8",
		7,
		"33F67PgGyFD73YmDg7JYbwEqufsB89vvpc",
	},
	"BIP141: P2WPKH nested in P2SH ypub at m/6'/0/0/0/0 extrapolated from m/6'/0": {
		"ypub6VGWAW57V8o1eaVnrvNPuKb7xfvP6X4hxb3vxGGQp9oyCKoZCFvbvLDYjy36RE1immwe5RSGGFiQULB65v9Zw1Ej2TbPR6CLTGkzhxMSQ4q",
		0,
		"3Qt1E64dYpUA1ovvYYfZYYeoCazjDTLww3",
	},
	"BIP141: P2WPKH nested in P2SH ypub at m/6'/4'/0/0/1 extrapolated from m/6'/4'": {
		"ypub6VGWAW5FpoKyziaLoi5RJnL1ezT4Xb5BwZH1HYmArsESpvs4rgP1fM8hPxQA6qEnxXVM5zFbzzSWKuuvB3QUFSbmFppU5hyQHyojgvf8S3f",
		1,
		"3PMemKpygPQDyzqsjdrnqMswRhNjcka1Wu",
	},
	"BIP141: P2WPKH nested in P2SH ypub at m/6'/4'/9'/0/11 extrapolated from m/6'/4'/9'": {
		"ypub6Xgwmha4MdU4UAfH9oXWtTmCvZxnqZXkp1ReiWsKeHvKid6QEhjA2CUDQcNCYpiExbPkZhr6DHtTUda3ZQp4KM7NfqVJWs3YFqHrJLECa1k",
		11,
		"382CXgbMHT4gzUiKQ2EHsJgESrBjv6bik8",
	},
	"BIP141: P2WPKH nested in P2SH ypub at m/6'/4'/9'/0/4": {
		"ypub6Z7s8wJuKsxjd16oe85WH1uSbcbbCXuMFEhPMgcf7jQqNhQbT9jE52XVu1eBe18q2J3LwnDd54ufL2jNvidjfCkbd34aVwLtYdztLUqucwR",
		4,
		"3FQZbk6HBX72j2yyT5L8hWTymaJgwvb9u9",
	},

	"BIP141: P2WPKH zpub at m/0/0/0/0/0 extrapolated from m/0": {
		"zpub6nBt2B13YbbubMCzAmoBLg4emchqHqyMQ9yLfzUwsuJvYgUHAcgLXjfqqHrprso814Croc8rheqMhV4h796L9dF67qFjoWrFC8FnLsVHknB",
		0,
		"bc1q8dnmvgj4jsvafe0wuwdm89aua2405jp7jp2zhn",
	},
	"BIP141: P2WPKH zpub at m/6'/0/0/0/7 extrapolated from m/6'": {
		"zpub6nBt2B1BtG8t2jc146zg8M8pjbffVk9UMXP6XJpcK1iyzww8VjqMDnRybqzahD3RBMu5XVjxZPdkSikyxscnxxRpjr777yvnRRQRDVAG58G",
		7,
		"bc1q4nrgadrcxz4tcqke9eh7f6zn8lxg8lz49d8v6k",
	},
	"BIP141: P2WPKH zpub at m/6'/0/0/0/0 extrapolated from m/6'/0": {
		"zpub6p6mUAk2dpLVVsguhHA27Qgd8e4q394Csha9jfAJCABrFRcnSv6AYPsgmAzgR8feBR4Spu2piv4xMcneocZajEvKtoHp111pizpe6aAEqfp",
		0,
		"bc1q9wwwcgcl2lw74quetxan4j6vhluyvyhy3dwt5l",
	},
	"BIP141: P2WPKH zpub at m/6'/4'/0/0/1 extrapolated from m/6'/4'": {
		"zpub6p6mUAkAyUsTr1mTe4s3WsRWpxbWUD4grfoE4wf4EscKt2gJ7LYaHQnqRAMk6jtiNAc9qTrATeo4DCXUtjpV3gHN8AWtfcntZhsP5UeStdx",
		1,
		"bc1qwsuszzk93puxlgs6l6f54r66dev882pm2760dv",
	},
	"BIP141: P2WPKH zpub at m/6'/4'/9'/0/11 extrapolated from m/6'/4'/9'": {
		"zpub6rXD5NEyWK1YKTrPzAK96Yri6Y7EnBXFj7wsVumD2JJCmiudVMtieG8MRpKnYjNANEWZKBSefxF1MvBcH7E57anyYBBj6ms2XZMVgsuPzzs",
		11,
		"bc1qsszrcep8whzqh93ksmmckn77eh9fl55s5dzjx6",
	},
	"BIP141: P2WPKH zpub at m/6'/4'/9'/0/4": {
		"zpub6sx8SbypUZWDUJHvUUs8V6zwmak399trAMDc95WYVjniRoDphotnh6BdvDbmdunkRwA9hFpBXjGDDKLweR3kTSSCVNm15rANpN4XixewDwG",
		4,
		"bc1q5l3j7e2s3dzg4vxquxldvzw4dwdlvvhdc9c4zh",
	},
}

func TestDeriveAddress(t *testing.T) {
	for testName, testData := range deriveAddressTestData {
		t.Run(testName, func(t *testing.T) {
			address, err := deriveAddress(testData.extendedAddress, uint32(testData.addressIndex))

			if err != nil {
				t.Fatal(err)
			}

			if address != testData.expectedAddress {
				t.Errorf(
					"unexpected derived address\nexpected: %s\nactual:   %s",
					testData.expectedAddress,
					address,
				)
			}
		})
	}
}

var deriveAddressTestFailureData = map[string]struct {
	extendedAddress string
	addressIndex    int
	failure         string
}{
	"BIP141 P2WPKH nested in P2SH ypub at m/6'/4'/9'/0/11'": {
		"ypub6Z7s8wJuKsxjd16oe85WH1uSbcbbCXuMFEhPMgcf7jQqNhQbT9jE52XVu1eBe18q2J3LwnDd54ufL2jNvidjfCkbd34aVwLtYdztLUqucwR",
		11 + 2147483648,
		"cannot derive a hardened key from a public key",
	},

	"BIP141 P2WPKH nested in P2SH ypub at m/6'/4'/9'/0/11' with a private key": {
		"yprvAL8WjRn1VWQSQX2LY6YVusxi3am6o5BVt1mnZJD3ZPsrVu5SucQyXED23ikCvDeeFHTMeX9q5n5MHNTLWQvCSm3KWnA3KdyZuDXncTn2VW5",
		11 + 2147483648,
		"unsupported public key format",
	},

	"BIP141 ypub is too deep at m/0/0/0/0/0/0": {
		"ypub6bp11ZqNVMqm3C3eXAFGpEvKqNfEZ6Vhznd4Uo3S73RYTSFgmF7q9sWPoCFhLGVMSLqKZZpcpHoKgHNwStDuqQPnDfF13goQwS8qSFA6vnz",
		0,
		"extended public key is deeper than 4, depth: 5",
	},
}

func ErrorContains(err error, expected string) bool {
	return strings.Contains(err.Error(), expected)
}

func TestDeriveAddress_ExpectedFailure(t *testing.T) {
	for testName, testData := range deriveAddressTestFailureData {
		t.Run(testName, func(t *testing.T) {
			_, err := deriveAddress(testData.extendedAddress, uint32(testData.addressIndex))
			if !ErrorContains(err, testData.failure) {
				t.Errorf(
					"unexpected error message\nexpected: %s\nactual:   %s",
					testData.failure,
					err.Error(),
				)
			}
		})
	}
}

var resolveAddressData = map[string]struct {
	beneficiaryAddress string
	addressIndex       int
	netParams          *chaincfg.Params
	expectedAddress    string
}{
	"BIP44: xpub at m/44'/0'/0'/0/0": {
		"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1",
		0,
		&chaincfg.MainNetParams,
		"1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx",
	},

	"Standard mainnet P2PKH btc address": {
		"1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx",
		0,
		&chaincfg.MainNetParams,
		"1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx",
	},

	"Standard mainnet P2SH btc address": {
		"3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy",
		0,
		&chaincfg.MainNetParams,
		"3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy",
	},

	"Standard mainnet Bech32 btc address": {
		"bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
		0,
		&chaincfg.MainNetParams,
		"bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
	},

	"Standard testnet btc address": {
		"mkHS9ne12qx9pS9VojpwU5xtRd4T7X7ZUt",
		0,
		&chaincfg.TestNet3Params,
		"mkHS9ne12qx9pS9VojpwU5xtRd4T7X7ZUt",
	},
}

func TestResolveAddress(t *testing.T) {
	for testName, testData := range resolveAddressData {
		t.Run(testName, func(t *testing.T) {
			resolvedAddress, err := ResolveAddress(
				testData.beneficiaryAddress,
				uint32(testData.addressIndex),
				testData.netParams,
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
	netParams       *chaincfg.Params
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
		"the provided serialized extended key length is invalid",
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
			_, err := ResolveAddress(
				testData.extendedAddress,
				0,
				testData.netParams,
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
