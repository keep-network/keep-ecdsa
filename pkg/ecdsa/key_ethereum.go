//+build !celo

package ecdsa

import (
	cecdsa "crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
)

// PublicKey holds a public key of the Signer.
type PublicKey cecdsa.PublicKey

// Marshal serializes Public Key to bytes in uncompressed form as described in
// [SEC 1] section 2.3.3: `04 + <x coordinate> + <y coordinate>`
func (pk *PublicKey) Marshal() []byte {
	return crypto.FromECDSAPub((*cecdsa.PublicKey)(pk))
}
