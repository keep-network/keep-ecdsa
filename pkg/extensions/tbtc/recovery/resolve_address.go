package recovery

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/keep-network/keep-ecdsa/pkg/chain/bitcoin"
)

// ResolveAddress resolves a configured beneficiaryAddress into a
// valid bitcoin address. If the supplied address is already a valid bitcoin
// address, we don't have to do anything. If the supplied address is an
// extended public key of a HD wallet, attempt to derive the bitcoin address at
// the specified index. The function will store an index of the last resolved
// bitcoin address unless it is a dry run.
//
// The function does not validate inputs. It is expected that validations are
// performed before calling this function. Especially the beneficiary address
// should be validated with the given chain network type.
func ResolveAddress(
	beneficiaryAddress string,
	storage *DerivationIndexStorage,
	chainParams *chaincfg.Params,
	handle bitcoin.Handle,
	isDryRun bool,
) (string, error) {
	// If the address decodes without error, then we have a valid bitcoin
	// address. Otherwise, we assume that it's an extended key and we attempt to
	// derive the address.
	decodedAddress, err := btcutil.DecodeAddress(beneficiaryAddress, chainParams)
	if err != nil {
		derivedAddress, err := storage.GetNextAddress(
			beneficiaryAddress,
			handle,
			chainParams,
			isDryRun,
		)
		if err != nil {
			return "", err
		}
		return derivedAddress, nil
	}
	return decodedAddress.EncodeAddress(), nil
}
