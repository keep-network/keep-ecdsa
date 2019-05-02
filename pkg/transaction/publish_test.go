package transaction

import (
	"fmt"
	"reflect"
	"testing"

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
