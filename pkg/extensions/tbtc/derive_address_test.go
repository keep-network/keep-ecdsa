package tbtc

import (
	"testing"
)

// These tests use https://iancoleman.io/bip39/ with the bip39 mnemonic: loyal
// chuckle trade magnet tobacco jungle craft cram reduce climb size flip tongue
// tornado height
var deriveAddressTestData = []struct {
	extendedAddress string
	addressIndex    int
	expectedAddress string
}{
	// BIP44: xpub
	{
		"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1",
		0,
		"1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx",
	},
	{

		"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1",
		4,
		"1EEX8qZnTw1thadyxsueV748v3Y6tTMccc",
	},
	// BIP49: ypub
	{
		"ypub6Xxan668aiJqvh4SVfd7EzqjWvf36gWufTkhWHv3gaxnBh44HpkTi2TTkm1u136qjUxk7F3jGzoyfrGpHvALMgJgbF4WNXpoPu3QYrqogMK",
		0,
		"3Aobe26f7QzKN73mvYQVbt1KLrCU1CgQpD",
	},
	{
		"ypub6Xxan668aiJqvh4SVfd7EzqjWvf36gWufTkhWHv3gaxnBh44HpkTi2TTkm1u136qjUxk7F3jGzoyfrGpHvALMgJgbF4WNXpoPu3QYrqogMK",
		4,
		"3Ap2E4ap2ZqzUHkTT8ZZv2DJm6TqKukBAL",
	},
	// BIP84: zpub
	{
		"zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9",
		0,
		"bc1q46uejlhm9vkswfcqs9plvujzzmqjvtfda3mra6",
	},

	{
		"zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9",
		8,
		"bc1quq0vrufxy05ypk45xmu3hpk6qsmlhr5vr3n8kz",
	},
	// BIP141: ypub again
	{
		"ypub6ZrwUDiiKM2b6CmJy6TSntc3b7FTE7dcUcu36jQLsXDu1uKt8jt19ZZc43vzfGs2f2r1hxWpbz8tCzJibmzwp2piYjzkhUCRzdrWU5qUoVZ",
		0,
		"3Aobe26f7QzKN73mvYQVbt1KLrCU1CgQpD",
	},
	{
		"ypub6ZrwUDiiKM2b6CmJy6TSntc3b7FTE7dcUcu36jQLsXDu1uKt8jt19ZZc43vzfGs2f2r1hxWpbz8tCzJibmzwp2piYjzkhUCRzdrWU5qUoVZ",
		16,
		"3JuNnoMh8eWhtY5YLk3SMXfw7vm8y6zPLg",
	},
	// BIP141 at m/0
	{
		"ypub6TMciWL8Pv4Rk41sLR1Z8ay9beZPMDyrV3T7tbb4Vtw3Vaf3uxWmug1hp5uEry9CbR6448YJEzUopCT8PSgKMPZVFVZKDc2kvQC8xHqdtZa",
		7,
		"3JbDbN7rYRFcBsU1BHwqrfua8TMAeMerqb",
	},
	// BIP141 at m/0'/0'
	{
		"ypub6Uy1fJC1qHtHWY6Gkavofq465y6Wi23cLGfoaQBLbTA5DjKLv7R2Qzz3ZJi6ZU8EopNGZEmHuCrmY7Dey5Fu9Jxa3XAAxZQPHhD3WLaGsin",
		2,
		"38Nfe3mepBtWicrDghCAhZitgeQotggyfF",
	},
}

func Test_DeriveAddress(t *testing.T) {
	for _, testData := range deriveAddressTestData {
		address, err := DeriveAddress(testData.extendedAddress, testData.addressIndex)

		if err != nil {
			t.Errorf(
				"Got %s while trying to derive %s at index %v",
				err,
				testData.extendedAddress,
				testData.addressIndex,
			)
		}

		if address != testData.expectedAddress {
			t.Errorf(
				"Got %s instead of %s while trying to derive %s at index %v",
				address,
				testData.expectedAddress,
				testData.extendedAddress,
				testData.addressIndex,
			)
		}
	}
}
