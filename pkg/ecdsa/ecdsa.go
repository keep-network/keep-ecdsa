// Package ecdsa defines interfaces for ECDSA signing based on [SEC 1].
//
//   [SEC 1]: Standards for Efficient Cryptography, SEC 1: Elliptic Curve
//     Cryptography, Certicom Research, https://www.secg.org/sec1-v2.pdf
package ecdsa

import (
	"fmt"
	"math/big"
)

// Signature holds a signature in a form of two big.Int `r` and `s` values and a
// recovery ID value in {0, 1, 2, 3}.
//
// The signature is chain-agnostic. Some chains (e.g. Ethereum and BTC) requires
// `v` to start from 27. Please consult the documentation about what the
// particular chain expects.
type Signature struct {
	R          *big.Int
	S          *big.Int
	RecoveryID int
}

// String formats Signature to a string that contains R and S values as hexadecimals.
func (s *Signature) String() string {
	return fmt.Sprintf("R: %#x, S: %#x, RecoveryID: %d", s.R, s.S, s.RecoveryID)
}
