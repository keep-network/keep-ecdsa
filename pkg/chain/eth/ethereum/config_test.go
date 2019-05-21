package ethereum

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestContractAddress(t *testing.T) {
	contractName1 := "KeepECDSAContract"
	validContractAddress := "0xbb2Ea17985f13D43e3AEC3963506A1B25ADDd57F"

	contractName2 := "InvalidContract"
	invalidHex := "0xZZZ"

	config := &Config{
		ContractAddresses: map[string]string{
			contractName1: validContractAddress,
			contractName2: invalidHex,
		},
	}

	var tests = map[string]struct {
		contractName    string
		expectedAddress common.Address
		expectedError   error
	}{
		"contract name matching valid configuration": {
			contractName:    contractName1,
			expectedAddress: common.HexToAddress(validContractAddress),
		},
		"invalid contract hex address": {
			contractName:    contractName2,
			expectedAddress: common.Address{},
			expectedError: fmt.Errorf(
				"configured address [%v] for contract [%v] is not valid hex address",
				invalidHex,
				contractName2,
			),
		},
		"missing contract configuration": {
			contractName:    "Peekaboo",
			expectedAddress: common.Address{},
			expectedError:   fmt.Errorf("configuration for contract [Peekaboo] not found"),
		},
	}
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {

			actualAddress, err := config.ContractAddress(test.contractName)
			if !reflect.DeepEqual(test.expectedError, err) {
				t.Errorf(
					"unexpected error\nexpected: %v\nactual:   %v\n",
					test.expectedError,
					err,
				)
			}

			if !bytes.Equal(test.expectedAddress.Bytes(), actualAddress.Bytes()) {
				t.Errorf(
					"unexpected address\nexpected: %v\nactual:   %v\n",
					test.expectedAddress,
					actualAddress,
				)
			}

		})
	}
}
