// Package ecdsa defines interfaces for ECDSA signing based on [SEC 1].
//
//   [SEC 1]: Standards for Efficient Cryptography, SEC 1: Elliptic Curve
//     Cryptography, Certicom Research, https://www.secg.org/sec1-v2.pdf
package ecdsa

import (
	cecdsa "crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

// PublicKey holds a public key of the Signer.
type PublicKey cecdsa.PublicKey

// Marshal serializes Public Key to bytes in uncompressed form as described in
// [SEC 1] section 2.3.3: `04 + <x coordinate> + <y coordinate>`
func (pk *PublicKey) Marshal() []byte {
	return crypto.FromECDSAPub((*cecdsa.PublicKey)(pk))
}

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
