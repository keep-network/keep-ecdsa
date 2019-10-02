package eth

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestContainsMember(t *testing.T) {
	address1 := common.HexToAddress("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	address2 := common.HexToAddress("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA1")
	address3 := common.HexToAddress("1AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")

	var tests = map[string]struct {
		members        []common.Address
		address        common.Address
		expectedResult bool
	}{
		"matches item in single item slice": {
			members:        []common.Address{address1},
			address:        address1,
			expectedResult: true,
		},
		"matches last item in multiple item slice": {
			members:        []common.Address{address1, address2, address3},
			address:        address3,
			expectedResult: true,
		},
		"returns false when item is not in the slice": {
			members:        []common.Address{address1, address2},
			address:        address3,
			expectedResult: false,
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			event := &ECDSAKeepCreatedEvent{Members: test.members}

			result := event.ContainsMember(test.address)

			if result != test.expectedResult {
				t.Errorf(
					"unexpected result\nexpected: [%v]\nactual:   [%v]",
					test.expectedResult,
					result,
				)
			}
		})
	}
}
