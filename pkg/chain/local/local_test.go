package local

import (
	"fmt"
	"reflect"
	"testing"
)

func TestBroadcastTransaction(t *testing.T) {
	chain := Connect()

	var tests = map[string]struct {
		rawTx          string
		expectedResult string
		expectedError  error
	}{
		"successful transaction broadcast": {
			rawTx:          "0123456789ABCDEF",
			expectedResult: "2125b2c332b1113aae9bfc5e9f7e3b4c91d828cb942c2df1eeb02502eccae9e9",
		},
		"failed transaction broadcast": {
			rawTx:          "",
			expectedResult: "",
			expectedError:  fmt.Errorf("empty transaction provided"),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			result, err := chain.SubmitTransaction(test.rawTx)

			if !reflect.DeepEqual(test.expectedError, err) {
				t.Errorf(
					"unexpected error\nexpected: %v\nactual:   %v\n",
					test.expectedError,
					err,
				)
			}

			if test.expectedResult != result {
				t.Errorf(
					"unexpected result\nexpected: %v\nactual:   %v\n",
					test.expectedResult,
					result,
				)
			}
		})
	}
}
