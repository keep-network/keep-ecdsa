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
// the specified index.
func ResolveAddress(
	beneficiaryAddress string,
	storage *DerivationIndexStorage,
	chainParams *chaincfg.Params,
	handle bitcoin.Handle,
) (string, error) {
	// If the address decodes without error, then we have a valid bitcoin
	// address. Otherwise, we assume that it's an extended key and we attempt to
	// derive the address.
	_, err := btcutil.DecodeAddress(beneficiaryAddress, chainParams)
	if err != nil {
		derivedAddress, err := storage.GetNextAddress(beneficiaryAddress, handle)
		if err != nil {
			return "", err
		}
		return derivedAddress, nil
	}
	return beneficiaryAddress, nil
}
