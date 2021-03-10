//+build !celo

package tss

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// Components defined in this file act as a facade in front of components
// provided by the `go-ethereum/crypto` package. This way, the client code
// does not have to import `go-ethereum/crypto` directly. As result, the
// backing package can be swapped to an other `go-ethereum/crypto`-like package
// at build time, via according build tags. That solution aims to address
// problems with package conflicts as package `go-ethereum/crypto` uses some
// native C code underneath. If another `go-ethereum/crypto`-like package
// is used in the same time, the compilation may fail due to C linker errors
// caused by duplicated symbols. Such a problem can be observed while trying
// to use the `go-ethereum/crypto` and `celo-blockchain/crypto` at the
// same time.

var S256 = secp256k1.S256
var SigToPub = crypto.SigToPub
