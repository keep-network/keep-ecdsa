//+build celo

package tss

import (
	"github.com/celo-org/celo-blockchain/crypto"
	"github.com/celo-org/celo-blockchain/crypto/secp256k1"
)

var S256 = secp256k1.S256
var SigToPub = crypto.SigToPub
