package btc

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

// PublicKeyToWitnessPubKeyHashAddress convert public key to bitcoin Witness
// Public Key Hash Address. It calculates the address according to [BIP-173].
// First a public key is serialized to compressed format. Next witness program
// is calculates as RIPEMD-160 hash over SHA-256 hash of compressed public key.
// Finally bitcoin address is created for a specific network.
//
// [BIP-173]: https://github.com/bitcoin/bips/blob/master/bip-0173.mediawiki
func PublicKeyToWitnessPubKeyHashAddress(
	publicKey *ecdsa.PublicKey,
	netParams *chaincfg.Params,
) (string, error) {
	compressedPublicKey := (*btcec.PublicKey)(publicKey).SerializeCompressed()

	// Hash ripemd160(sha256(compressedPublicKey)).
	witnessProgram := btcutil.Hash160(compressedPublicKey)

	address, err := btcutil.NewAddressWitnessPubKeyHash(witnessProgram, netParams)
	if err != nil {
		return "", err
	}

	return address.EncodeAddress(), nil
}
