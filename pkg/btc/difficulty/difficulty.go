package difficulty

import (
	"fmt"
	"math/big"

	"github.com/keep-network/keep-tecdsa/pkg/chain/btc"
)

const difficultyTarget1 = "0x00000000FFFF0000000000000000000000000000000000000000000000000000"

// TODO: Publish difficulty to contract

// GetCurrentDifficulty TODO: write doc
func GetCurrentDifficulty(chain btc.Interface) (*big.Int, error) {
	targetBits, err := chain.GetCurrentTargetBits()
	if err != nil {
		return nil, fmt.Errorf("cannot get current target: [%s]", err)
	}
	// TODO: Publish to contract
	difficulty, err := calculateDifficulty(targetBits)
	if err != nil {
		return nil, fmt.Errorf("difficulty calculation failed: [%s]", err)
	}

	return difficulty, nil
}

func calculateDifficulty(targetBits int) (*big.Int, error) {
	target1, ok := new(big.Int).SetString(difficultyTarget1, 0)
	if !ok {
		return nil, fmt.Errorf("difficulty target 1 decoding failed")
	}

	target := compactToBig(uint32(targetBits))

	curentDifficulty := new(big.Int).Div(target1, target)

	return curentDifficulty, nil
}

// Copied from `CompactToBig` function in `github.com/btcsuite/btcd/blockchain`
// package.
func compactToBig(compact uint32) *big.Int {
	// Extract the mantissa, sign bit, and exponent.
	mantissa := compact & 0x007fffff
	isNegative := compact&0x00800000 != 0
	exponent := uint(compact >> 24)

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes to represent the full 256-bit number.  So,
	// treat the exponent as the number of bytes and shift the mantissa
	// right or left accordingly.  This is equivalent to:
	// N = mantissa * 256^(exponent-3)
	var bn *big.Int
	if exponent <= 3 {
		mantissa >>= 8 * (3 - exponent)
		bn = big.NewInt(int64(mantissa))
	} else {
		bn = big.NewInt(int64(mantissa))
		bn.Lsh(bn, 8*(exponent-3))
	}

	// Make it negative if the sign bit is set.
	if isNegative {
		bn = bn.Neg(bn)
	}

	return bn
}
