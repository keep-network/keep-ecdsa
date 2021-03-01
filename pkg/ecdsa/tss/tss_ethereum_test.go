//+build !celo

package tss

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

var S256 = secp256k1.S256
var SigToPub = crypto.SigToPub
