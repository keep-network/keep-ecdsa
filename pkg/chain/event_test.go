package chain

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

type keepMemberIDAddress common.Address
func (kmia keepMemberIDAddress) String() string {
	return common.Address(kmia).String()
}
func (kmia keepMemberIDAddress) OperatorID() OperatorID {
	return kmia
}
func (kmia keepMemberIDAddress) KeepMemberID(keepID KeepID) KeepMemberID {
	return kmia
}

func TestIsMember(t *testing.T) {
	address1 := keepMemberIDAddress(common.HexToAddress("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"))
	address2 := keepMemberIDAddress(common.HexToAddress("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA1"))
	address3 := keepMemberIDAddress(common.HexToAddress("1AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"))

	var tests = map[string]struct {
		members        []KeepMemberID
		address        keepMemberIDAddress
		expectedResult bool
	}{
		"matches item in single item slice": {
			members:        []KeepMemberID{address1},
			address:        address1,
			expectedResult: true,
		},
		"matches first item in multiple item slice": {
			members:        []KeepMemberID{address1, address2, address3},
			address:        address1,
			expectedResult: true,
		},
		"matches middle item in multiple item slice": {
			members:        []KeepMemberID{address1, address2, address3},
			address:        address2,
			expectedResult: true,
		},
		"matches last item in multiple item slice": {
			members:        []KeepMemberID{address1, address2, address3},
			address:        address3,
			expectedResult: true,
		},
		"returns false when item is not in the slice": {
			members:        []KeepMemberID{address1, address2},
			address:        address3,
			expectedResult: false,
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			event := &BondedECDSAKeepCreatedEvent{Members: test.members}

			result := event.IsMember(test.address)

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
