package bitcoin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestBroadcast(t *testing.T) {
	transaction := "01000000000101ba84a592005742406bd1d6683e3a894c7ab13385bd437ff7bd7c74929bf141320000000000000000000309cc320000000000160014a0aedee089b0cfa34e1e29c2dd2e618b19e8b95309cc320000000000160014f8c4e8695f8c2e0f598f8f00c2c4a83b17b0c4fa09cc320000000000160014fada4235022b32a31f97adbc954e6a7bbb7b32ba024830450221008dd10d4f331a61c2afe948dec6f900b29996c29262a79fa4d72acacd0c19497a022063229d6751c47e3e9b67e567bd98500b66630a09ef8515377a18d0479135c84f01210329fb706ee25a944362c4a53a5b4fa6f47201354d567e753c3998f15a36996b8100000000"

	electrs := newTestElectrsConnection(
		mockClient{
			mockPost: mockPost(fmt.Sprintf("%s/tx", testApiUrl), transaction, 200, "<fake-tx-id>", t),
		},
	)

	err := electrs.Broadcast(transaction)
	if err != nil {
		t.Error(err)
	}
}

func TestBroadcast_EmptyApiURL(t *testing.T) {
	expectedError := "attempted to call Broadcast with no apiURL"

	electrs := &electrsConnection{}

	err := electrs.Broadcast("1234567890")
	if err.Error() != expectedError {
		t.Errorf(
			"unexpected error\nexpected: %v\nactual:   %v",
			err,
			expectedError,
		)
	}
}

func TestBroadcast_ExpectFailure(t *testing.T) {
	transaction := "0123456789aBcDeF"
	mockedResponseCode := 400
	mockedResponseBody := `sendrawtransaction RPC error: {"code":-25,"message":"bad-txns-inputs-missingorspent"}`
	expectedError := `failed to broadcast transaction - status: [400 Bad Request], payload: [sendrawtransaction RPC error: {"code":-25,"message":"bad-txns-inputs-missingorspent"}]; raw transaction: [0123456789aBcDeF]`

	electrs := newTestElectrsConnection(
		mockClient{
			mockPost: mockPost(
				fmt.Sprintf("%s/tx", testApiUrl),
				transaction,
				mockedResponseCode,
				mockedResponseBody,
				t,
			),
		},
	)

	err := electrs.Broadcast(transaction)
	checkWrappedError(err, expectedError, t)
}

func TestVbyteFeeFor25Blocks(t *testing.T) {
	mockedResponseCode := 200
	mockedResponseBody := `{ "1": 87.882, "2": 87.882, "3": 87.882, "4": 87.882, "5": 81.129, "6": 68.285, "7": 65.182, "8": 63.876, "9": 61.153, "10": 60.172, "11": 57.721, "12": 54.753, "13": 52.879, "14": 46.872, "15": 42.871, "16": 39.989, "17": 35.919, "18": 30.821, "19": 25.888, "20": 21.876, "21": 16.156, "22": 11.222, "23": 10.982, "24": 9.654, "25": 7.883, "144": 1.027, "504": 1.027, "1008": 1.027 }`
	expectedFee := int32(7)

	electrs := newTestElectrsConnection(
		mockClient{
			mockGet: mockGet(
				fmt.Sprintf("%s/fee-estimates", testApiUrl),
				mockedResponseCode,
				mockedResponseBody,
				t,
			),
		},
	)

	fee, err := electrs.VbyteFeeFor25Blocks()
	if err != nil {
		t.Fatal(err)
	}
	if fee != expectedFee {
		t.Errorf("unexpected fee\nexpected: %d\nactual:   %d", expectedFee, fee)
	}
}

func TestVbyteFeeFor25Blocks_EmptyApiURL(t *testing.T) {
	expectedError := "attempted to call VbyteFeeFor25Blocks with no apiURL"
	expectedFee := int32(0)

	electrs := &electrsConnection{}

	fee, err := electrs.VbyteFeeFor25Blocks()
	if err.Error() != expectedError {
		t.Errorf(
			"unexpected error\nexpected: %v\nactual:   %v",
			err,
			expectedError,
		)
	}
	if fee != expectedFee {
		t.Errorf("unexpected fee\nexpected: %d\nactual:   %d", expectedFee, fee)
	}
}

func TestVbyteFeeFor25Blocks_ExpectFailure(t *testing.T) {
	mockedResponseCode := 500
	mockedResponseBody := `the dumpster is on fire`
	expectedError := `failed to get fee estimates - status: [500 Internal Server Error], payload: [the dumpster is on fire]`
	expectedFee := int32(0)

	electrs := newTestElectrsConnection(
		mockClient{
			mockGet: mockGet(
				fmt.Sprintf("%s/fee-estimates", testApiUrl),
				mockedResponseCode,
				mockedResponseBody,
				t,
			),
		},
	)

	fee, err := electrs.VbyteFeeFor25Blocks()

	checkWrappedError(err, expectedError, t)

	if fee != expectedFee {
		t.Errorf("unexpected fee\nexpected: %d\nactual:   %d", expectedFee, fee)
	}

}

func TestIsAddressUnused(t *testing.T) {
	testData := map[string]struct {
		btcAddress         string
		response           string
		expectedUnusedFlag bool
	}{
		"address with unspent funds": {
			"bcrt1qy6n80gen875en87ka798svvzrneq2erhhwfzzf",
			`[{"txid":"2fd4fd49a9719be53affe55c4761abf00df1cda9b7a02419411bc9c04174c3f7","version":2,"locktime":14154,"vin":[{"txid":"fc4e62537ad932a84b4888826e10810119c854cbed1244091e864502282cee92","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":9765625},"scriptsig":"","scriptsig_asm":"","witness":["304402202f3bf6827a90d566f7543119156d0b2aa19262379429ca58e95728479258c64d02202aee28e0083c900fe57a4a7f3f1a65fc2b012acdde88b99349a5bca17336486901","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294},{"txid":"14e4defea6b2a164701562b0e9d379b23288b41b44ab8201749e5c5c312f0e59","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":4768},"scriptsig":"","scriptsig_asm":"","witness":["304402206feee35d7a3de261c4ddd0e9bbccdc65b24226a9be8aef2bcacc87b032eca93702206e2040856bd7d61f3a8d9f24ef515e7660a99984f4e55fcd2f7ea16ad56ecf2b01","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294},{"txid":"246d552c11741b7196f0f7dfd7c65a22daba9da20264edc26dbb8afcc2906465","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":4768},"scriptsig":"","scriptsig_asm":"","witness":["304402206fc18ffa446f681d8901d7d211f7d972e5e18420ac991d44befd13f9740a26a5022030079817769551fb818149b1d146ff6c02859060e56118796b711ba6e9f5e37801","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294},{"txid":"91517b69b67dee45c3cca4e2b1d5f0c4e04abc19363fc161b6dde9f72a2615ed","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":152587},"scriptsig":"","scriptsig_asm":"","witness":["3044022073658ccf19f921eba6d46fa642dadaed54bf011438c63656a0466ed21ceaf58202202a90e99d3b9e599f1f88e0816f2b8076ca8238ef9b5149c0bfe1dfd5a4967a2801","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294},{"txid":"8cff43db341777dc4a8de065a2357e5d546a922524adabc6c650b82bdb4a9db1","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":10461},"scriptsig":"","scriptsig_asm":"","witness":["3044022056b40dd139f57caef154b327cb815e6976e4cacf226fa15c70f14ab0abe1059802204b864b4842e578c0ff11a0fea59bb91fb44162d1f7e9e9e048a1c560491cf62701","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294},{"txid":"16c414db432c78fb2636e618b6b470265cbb68137d56101249c2d86a796cfc74","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":12901},"scriptsig":"","scriptsig_asm":"","witness":["304402202c3d563122099edf16996dadf0bebcc5105eb6e5e53d5e5a8ab8939731e12ff802205bc8bdf99feef369b55832e5676a9d39f006643af35b7778c9dfa891a4b9792b01","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294},{"txid":"3d93c2fef5cd7c8349170ec48dd96a4477bb6ca6958b6fb80a074ee7e80a421f","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":19073},"scriptsig":"","scriptsig_asm":"","witness":["30440220034c0a3dbab8ccc7c9004ce8d5b1da25d3a527493368c98531acb47e147df37c02207646277753c9694aebc59ac47a230735c0c2b6c0ecf5d2251a9350b9171b330201","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294},{"txid":"eb0d6c1f32f975b121fdbb0d5b010174fd47594fc7a2faa6184862d0688ac64d","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":4768},"scriptsig":"","scriptsig_asm":"","witness":["30440220748e72af5018b7cc5774d7debe225b7ad304a1cfe9e0a469470affe9622fbac102200ea7c7a94bd7813b0ac472be36b1c7cb77d59b4b63363aac39e580003ec55c8501","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294},{"txid":"0beb097b66aafd83f914dad6ebfee3bfbdc17c2111fa4dde964b928bba724330","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":38146},"scriptsig":"","scriptsig_asm":"","witness":["30440220507ca93500dfd6f83f5b45b401b335563287a09906ee5a35e7877f2c1b87eaa202204c68e3eb43fa5f3054d823bacadfe644b4626e1b938d1ab564ee30dfe07dff8d01","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294}],"vout":[{"scriptpubkey":"001426a677a3333fa9999fd6ef8a7831821cf2056477","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 26a677a3333fa9999fd6ef8a7831821cf2056477","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qy6n80gen875en87ka798svvzrneq2erhhwfzzf","value":10000000}],"size":1375,"weight":2605,"fee":13097,"status":{"confirmed":true,"block_height":14208,"block_hash":"3c95707c627031feca93af0473cf5dc81e3f4fd6a660023924a85900d3b294ce","block_time":1620420106}}]`,
			false,
		},
		"address with no transactions": {
			"bcrt1qy6n80gen875en87ka798svvzrneq2erhhwfzzf",
			`[]`,
			true,
		},
		"address used for recovery": {
			"bcrt1q07njh90vzjzdjwfg7mr6ek7swylm99z2l4cg7q",
			`[{"txid":"157617f0573262e466563272b643ce422dd378f86c0cfcac292776a979829b00","version":1,"locktime":0,"vin":[{"txid":"2fd4fd49a9719be53affe55c4761abf00df1cda9b7a02419411bc9c04174c3f7","vout":0,"prevout":{"scriptpubkey":"001426a677a3333fa9999fd6ef8a7831821cf2056477","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 26a677a3333fa9999fd6ef8a7831821cf2056477","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qy6n80gen875en87ka798svvzrneq2erhhwfzzf","value":10000000},"scriptsig":"","scriptsig_asm":"","witness":["3045022100fac6791f32fe18bcdd7373a7d0eec94640c34d392b3c4257220f6775c80dfeba022027bad28595c5ce84d9120b89645f4a074236e76df36ceeceaf50e6efe50a76a001","020a401579ff57588e7596fd192e04478dd7e24cddff34a2433d94eee5ce6462df"],"is_coinbase":false,"sequence":0}],"vout":[{"scriptpubkey":"00147fa72b95ec1484d93928f6c7acdbd0713fb2944a","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 7fa72b95ec1484d93928f6c7acdbd0713fb2944a","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1q07njh90vzjzdjwfg7mr6ek7swylm99z2l4cg7q","value":3329033},{"scriptpubkey":"00142dd3fe9a8675e368702d232f5542a34f231e23b8","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 2dd3fe9a8675e368702d232f5542a34f231e23b8","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1q9hflax5xwh3ksupdyvh42s4rfu33ugac8zjz4v","value":3329033},{"scriptpubkey":"0014c1dd3c3b6ec603aaf86be1cea289730c83492a75","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 c1dd3c3b6ec603aaf86be1cea289730c83492a75","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qc8wncwmwccp647rtu8829ztnpjp5j2n4x4pzxw","value":3329033}],"size":254,"weight":686,"fee":12901,"status":{"confirmed":true,"block_height":14212,"block_hash":"6b40428c4096d37d1edbc00234e8b4e5af52e42d1d1fdbb844fb735039af2f5a","block_time":1620420226}}]`,
			false,
		},
	}
	for testName, testData := range testData {
		t.Run(testName, func(t *testing.T) {
			electrs := newTestElectrsConnection(
				mockClient{
					mockGet: mockGet(
						fmt.Sprintf("%s/address/%s/txs", testApiUrl, testData.btcAddress),
						200,
						testData.response,
						t,
					),
				},
			)

			unusedFlag, err := electrs.IsAddressUnused(testData.btcAddress)
			if err != nil {
				t.Fatal(err)
			}
			if unusedFlag != testData.expectedUnusedFlag {
				t.Errorf(
					"unexpected IsAddresssUnused result\nexpected: %v\nactual:   %v",
					testData.expectedUnusedFlag,
					unusedFlag,
				)
			}
		})
	}
}

func TestIsAddressUnused_EmptyApiURL(t *testing.T) {
	expectedError := "attempted to call IsAddressUnused with no apiURL"
	expectedUnusedFlag := false

	electrs := &electrsConnection{}

	unusedFlag, err := electrs.IsAddressUnused("BtcAddress123")
	if err.Error() != expectedError {
		t.Errorf(
			"unexpected error\nexpected: %v\nactual:   %v",
			err,
			expectedError,
		)
	}
	if unusedFlag != expectedUnusedFlag {
		t.Errorf(
			"unexpected unused flag\nexpected: %v\nactual:   %v",
			expectedUnusedFlag,
			unusedFlag,
		)
	}
}

func TestIsAddressUnused_ExpectedFailures(t *testing.T) {
	expectedUnusedFlag := false

	testData := map[string]struct {
		btcAddress   string
		responseCode int
		response     string
		err          string
	}{
		"invalid btc address": {
			"banana",
			400,
			"Invalid Bitcoin address",
			"something went wrong trying to get information about address [banana] - status: [400 Bad Request], payload: [Invalid Bitcoin address]",
		},
		"garbled response": {
			"bcrt1qy6n80gen875en87ka798svvzrneq2erhhwfzzf",
			200,
			`[{"txi4fd49a9719be53affe55c4761abf00df1cda9b7a02419411bc9c04174c3f7","version":2,"locktime":14154,"vin":[{"txid":"fc4e62537ad932a84b4888826e10810119c854cbed1244091e864502282cee92","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":9765625},"scriptsig":"","scriptsig_asm":"","witness":["304402202f3bf6827a90d566f7543119156d0b2aa19262379429ca58e95728479258c64d02202aee28e0083c900fe57a4a7f3f1a65fc2b012acdde88b99349a5bca17336486901","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294},{"txid":"14e4defea6b2a164701562b0e9d379b23288b41b44ab8201749e5c5c312f0e59","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":4768},"scriptsig":"","scriptsig_asm":"","witness":["304402206feee35d7a3de261c4ddd0e9bbccdc65b24226a9be8aef2bcacc87b032eca93702206e2040856bd7d61f3a8d9f24ef515e7660a99984f4e55fcd2f7ea16ad56ecf2b01","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294},{"txid":"246d552c11741b7196f0f7dfd7c65a22daba9da20264edc26dbb8afcc2906465","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":4768},"scriptsig":"","scriptsig_asm":"","witness":["304402206fc18ffa446f681d8901d7d211f7d972e5e18420ac991d44befd13f9740a26a5022030079817769551fb818149b1d146ff6c02859060e56118796b711ba6e9f5e37801","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294},{"txid":"91517b69b67dee45c3cca4e2b1d5f0c4e04abc19363fc161b6dde9f72a2615ed","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":152587},"scriptsig":"","scriptsig_asm":"","witness":["3044022073658ccf19f921eba6d46fa642dadaed54bf011438c63656a0466ed21ceaf58202202a90e99d3b9e599f1f88e0816f2b8076ca8238ef9b5149c0bfe1dfd5a4967a2801","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294},{"txid":"8cff43db341777dc4a8de065a2357e5d546a922524adabc6c650b82bdb4a9db1","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":10461},"scriptsig":"","scriptsig_asm":"","witness":["3044022056b40dd139f57caef154b327cb815e6976e4cacf226fa15c70f14ab0abe1059802204b864b4842e578c0ff11a0fea59bb91fb44162d1f7e9e9e048a1c560491cf62701","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294},{"txid":"16c414db432c78fb2636e618b6b470265cbb68137d56101249c2d86a796cfc74","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":12901},"scriptsig":"","scriptsig_asm":"","witness":["304402202c3d563122099edf16996dadf0bebcc5105eb6e5e53d5e5a8ab8939731e12ff802205bc8bdf99feef369b55832e5676a9d39f006643af35b7778c9dfa891a4b9792b01","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294},{"txid":"3d93c2fef5cd7c8349170ec48dd96a4477bb6ca6958b6fb80a074ee7e80a421f","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":19073},"scriptsig":"","scriptsig_asm":"","witness":["30440220034c0a3dbab8ccc7c9004ce8d5b1da25d3a527493368c98531acb47e147df37c02207646277753c9694aebc59ac47a230735c0c2b6c0ecf5d2251a9350b9171b330201","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294},{"txid":"eb0d6c1f32f975b121fdbb0d5b010174fd47594fc7a2faa6184862d0688ac64d","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":4768},"scriptsig":"","scriptsig_asm":"","witness":["30440220748e72af5018b7cc5774d7debe225b7ad304a1cfe9e0a469470affe9622fbac102200ea7c7a94bd7813b0ac472be36b1c7cb77d59b4b63363aac39e580003ec55c8501","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294},{"txid":"0beb097b66aafd83f914dad6ebfee3bfbdc17c2111fa4dde964b928bba724330","vout":0,"prevout":{"scriptpubkey":"001401160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 01160c280c429cb349c5c37d786eedb5b7bfd3e2","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qqytqc2qvg2wtxjw9cd7hsmhdkkmml5lzqc04tk","value":38146},"scriptsig":"","scriptsig_asm":"","witness":["30440220507ca93500dfd6f83f5b45b401b335563287a09906ee5a35e7877f2c1b87eaa202204c68e3eb43fa5f3054d823bacadfe644b4626e1b938d1ab564ee30dfe07dff8d01","0360eda8d2a5e92cfa752f756a972442271381c1b2081f8e43e6f3475c7feedf76"],"is_coinbase":false,"sequence":4294967294}],"vout":[{"scriptpubkey":"001426a677a3333fa9999fd6ef8a7831821cf2056477","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 26a677a3333fa9999fd6ef8a7831821cf2056477","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"bcrt1qy6n80gen875en87ka798svvzrneq2erhhwfzzf","value":10000000}],"size":1375,"weight":2605,"fee":13097,"status":{"confirmed":true,"block_height":14208,"block_hash":"3c95707c627031feca93af0473cf5dc81e3f4fd6a660023924a85900d3b294ce","block_time":1620420106}}]`,
			"failed to decode response body: [invalid character ',' after object key]",
		},
		"unexpected integer response": {
			"bcrt1qy6n80gen875en87ka798svvzrneq2erhhwfzzf",
			200,
			"73",
			"failed to decode response body: [json: cannot unmarshal number into Go value of type []interface {}]",
		},
	}

	for testName, testData := range testData {
		t.Run(testName, func(t *testing.T) {
			electrs := newTestElectrsConnection(
				mockClient{
					mockGet: mockGet(
						fmt.Sprintf("%s/address/%s/txs", testApiUrl, testData.btcAddress),
						testData.responseCode,
						testData.response,
						t,
					),
				},
			)

			unusedFlag, err := electrs.IsAddressUnused(testData.btcAddress)
			if err == nil {
				t.Errorf("unexected error\nexpected: %s\nactual:   nil", testData.err)
			}

			checkWrappedError(err, testData.err, t)

			if unusedFlag != expectedUnusedFlag {
				t.Errorf(
					"unexpected unused flag\nexpected: %v\nactual:   %v",
					expectedUnusedFlag,
					unusedFlag,
				)
			}
		})
	}
}

const testApiUrl = "example.org/api"

func newTestElectrsConnection(client mockClient) *electrsConnection {
	electrs := &electrsConnection{
		apiURL:  testApiUrl,
		timeout: 100 * time.Millisecond,
	}

	electrs.setClient(client)

	return electrs
}

type mockClient struct {
	mockGet  func(url string) (*http.Response, error)
	mockPost func(url string, contentType string, reader io.Reader) (*http.Response, error)
}

func (m mockClient) Get(url string) (*http.Response, error) {
	return m.mockGet(url)
}

func (m mockClient) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	return m.mockPost(url, contentType, body)
}

func mockGet(expectedURL string, responseStatusCode int, responseBody string, t *testing.T) func(url string) (*http.Response, error) {
	return func(url string) (*http.Response, error) {
		if url != expectedURL {
			t.Fatalf("unexpected url\nexpected: %s\nactual:   %s", expectedURL, url)
		}

		return mockResponse(responseStatusCode, responseBody), nil
	}
}

func mockPost(expectedURL string, expectedRequestBody string, responseStatusCode int, responseBody string, t *testing.T) func(url string, contentType string, reader io.Reader) (*http.Response, error) {
	return func(url string, contentType string, body io.Reader) (*http.Response, error) {
		if url != expectedURL {
			t.Fatalf("unexpected url\nexpected: %s\nactual:   %s", expectedURL, url)
		}

		if contentType != "text/plain" {
			t.Fatalf(
				"unexpected content type\nexpected: %s\nactual:   %s",
				"text/plain",
				contentType,
			)
		}

		bodyBytes, err := io.ReadAll(body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		if string(bodyBytes) != expectedRequestBody {
			t.Fatalf(
				"unexpected request body\nexpected: %s\nactual:   %s",
				expectedRequestBody,
				bodyBytes,
			)
		}

		return mockResponse(responseStatusCode, responseBody), nil
	}
}

func mockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(body))),
	}
}

func checkWrappedError(errWrapper error, expectedWrappedError string, t *testing.T) {
	if errWrapper == nil {
		t.Errorf(
			"unexpected error\nexpected: %v\nactual:   %v",
			expectedWrappedError,
			errWrapper,
		)
	}

	wrappedError := errors.Unwrap(errWrapper)
	if wrappedError.Error() != expectedWrappedError {
		t.Errorf(
			"unexpected unwrapped error\nexpected: %v\nactual:   %v",
			expectedWrappedError,
			wrappedError,
		)
	}
}
