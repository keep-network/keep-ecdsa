package transaction

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/keep-network/keep-tecdsa/pkg/chain/electrum"
	"github.com/keep-network/keep-tecdsa/pkg/chain/local"
)

func TestPublish(t *testing.T) {
	chain := local.Connect()

	var tests = map[string]struct {
		rawTx          string
		expectedResult string
		expectedError  error
	}{
		"successful transaction publication": {
			rawTx:          "0123456789ABCDEF",
			expectedResult: "2125b2c332b1113aae9bfc5e9f7e3b4c91d828cb942c2df1eeb02502eccae9e9",
		},
		"failed transaction publication": {
			rawTx:          "",
			expectedResult: "",
			expectedError:  fmt.Errorf("transaction broadcast failed [empty transaction provided]"),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			result, err := Publish(chain, test.rawTx)

			if !reflect.DeepEqual(test.expectedError, err) {
				t.Errorf(
					"unexpected error\nexpected: %v\nactual:   %v\n",
					test.expectedError,
					err,
				)
			}

			if test.expectedResult != result {
				t.Errorf(
					"\nexpected: %v\nactual:   %v\n",
					test.expectedResult,
					result,
				)
			}
		})
	}
}

// TODO: This is a temporary integration test. It should not be executed as an
// unit test. It is expected to be fail due to transaction already in block chain.
func TestPublishElectrum(t *testing.T) {
	// BTC testnet electrum servers: https://1209k.com/bitcoin-eye/ele.php?chain=tbtc
	config := &electrum.Config{
		ServerHost: "testnet.hsmiths.com",
		ServerPort: "53012",
	}

	chain, err := electrum.Connect(config)
	if err != nil {
		t.Fatal(err)
	}

	rawTx := "02000000000101506fda83a9788dab896b90bfc122c700afccd30459e10a0ff270e951202612481600000017160014997f3e8bcf47183fbbe0c8464175047bf391ea83feffffff0279d513000000000016001432b027edb95eee83b40003762a2ff25ae47d560d40420f000000000016001469abce7925fce369303e247c3a465447f6519b780247304402205be426e3e0c2e243eac4808cca306e3f821f7e431e2f0fc0c534851d5551779402202e98a4c7899d0eb2f473af77af6b434a5a47a0d898ed5b41ca2a7692601e36eb012102cb34ff4b355a7f02104cb912dbf4e35a14733030450ac311f6ef498d150a6ad2eb171700"

	result, err := Publish(chain, rawTx)
	if err != nil {
		t.Errorf("unexpected error [%s]", err)
	}
	t.Logf("Result: %v", result)
}
