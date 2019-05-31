package difficulty

import (
	"math/big"
	"testing"

	"github.com/keep-network/keep-tecdsa/pkg/chain/btc/local"
)

func TestGetCurrentDifficulty(t *testing.T) {
	chain := local.Connect()

	// local chain is configured with current difficulty target bits equal 388348790
	expectedResult := big.NewInt(7459680720542)

	actualResult, err := GetCurrentDifficulty(chain)
	if err != nil {
		t.Fatalf("unexpected error: [%s]", err)
	}

	if expectedResult.Cmp(actualResult) != 0 {
		t.Errorf(
			"invalid difficulty\nexpected: %v\nactual:   %v\n",
			expectedResult,
			actualResult,
		)
	}
}
