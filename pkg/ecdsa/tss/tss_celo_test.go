//+build celo

package tss

import (
	"github.com/celo-org/celo-blockchain/crypto"
	"github.com/celo-org/celo-blockchain/crypto/secp256k1"
)

// Components defined in this file act as a facade in front of components
// provided by the `celo-blockchain/crypto` package. This way, the client code
// does not have to import `celo-blockchain/crypto` directly. As result, the
// backing package can be swapped to an other `celo-blockchain/crypto`-like
// package at build time, via according build tags. That solution aims to
// address problems with package conflicts as package `celo-blockchain/crypto`
// uses some native C code underneath. If another `celo-blockchain/crypto`-like
// package is used in the same time, the compilation may fail due to C linker
// errors caused by duplicated symbols. Such a problem can be observed while
// trying to use the `celo-blockchain/crypto` and `ethereum-go/crypto` at the
// same time.

var S256 = secp256k1.S256
var SigToPub = crypto.SigToPub
