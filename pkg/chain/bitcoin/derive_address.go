package bitcoin

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
)

// DeriveAddress uses the specified extended public key and address index to
// derive an address string in the appropriate format at the specified address
// index. The extended public key can be at any level. deriveAddress will take
// the first child `/0` until a depth of 4 is reached, and then produce the
// address at the supplied index. Thus, calling deriveAddress with an xpub
// generated at m/44'/0' and passing the address index 5 will produce the
// address at path m/44'/0'/0/0/5.
//
// In cases where the extended public key is at depth 4, meaning the external or
// internal chain is already included, deriveAddress will directly derive the
// address index at the existing depth.
//
// deriveAddress does not support hardened child indexes (anything greater than
// or equal to 2147483648, abbreviated as 0')
//
// The returned address will be a p2pkh/p2sh address for prefixes xpub and tpub,
// (i.e. prefixed by 1, m, or n), a p2wpkh-in-p2sh address for prefixes ypub or
// upub (i.e., prefixed by 3 or 2), and a bech32 p2wpkh address for prefixes
// zpub or vpub (i.e., prefixed by bc1 or tb1).
//
// See [BIP32], [BIP44], [BIP49], and [BIP84] for more on address derivation,
// particular paths, etc.
//
// [BIP32]: https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki
// [BIP44]: https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki
// [BIP49]: https://github.com/bitcoin/bips/blob/master/bip-0049.mediawiki
// [BIP84]: https://github.com/bitcoin/bips/blob/master/bip-0084.mediawiki
func DeriveAddress(extendedPublicKey string, addressIndex uint32) (string, error) {
	extendedKey, err := hdkeychain.NewKeyFromString(extendedPublicKey)
	if err != nil {
		return "", fmt.Errorf(
			"error parsing extended public key: [%s]",
			err,
		)
	}

	externalChain := extendedKey
	if externalChain.Depth() > 4 {
		return "", fmt.Errorf("extended public key is deeper than 4, depth: %d", externalChain.Depth())
	}
	for externalChain.Depth() < 4 {
		// Descend the hierarchy at /0 until the external chain path, `m/*/*/*/0`.
		// ex: If we get a `m/32'/5` extended key, we descend to `m/32'/5/0/0`.
		externalChain, err = externalChain.Derive(0)
		if err != nil {
			return "", fmt.Errorf(
				"error deriving external chain path /0 from extended key: [%s]",
				err,
			)
		}
	}

	requestedPublicKey, err := externalChain.Derive(addressIndex)
	if err != nil {
		return "", fmt.Errorf(
			"error deriving requested address index /0/%v from extended key: [%s]",
			addressIndex,
			err,
		)
	}

	// Now to decide how we want to serialize the address...
	var chainParams *chaincfg.Params

	publicKeyDescriptor := extendedPublicKey[0:4]
	switch publicKeyDescriptor {
	case "xpub", "ypub", "zpub":
		chainParams = &chaincfg.MainNetParams
	case "tpub", "upub", "vpub":
		chainParams = &chaincfg.TestNet3Params
	default:
		return "", fmt.Errorf(
			"unsupported public key format [%s]",
			publicKeyDescriptor,
		)
	}

	requestedAddress, err := requestedPublicKey.Address(chainParams)
	if err != nil {
		return "", fmt.Errorf(
			"error retrieving the requested address from the public key with extended key [%v]: [%s]",
			extendedPublicKey,
			err,
		)
	}

	var finalAddress btcutil.Address = requestedAddress
	switch publicKeyDescriptor {
	case "xpub", "tpub":
		// Noop, the address is already correct
	case "ypub", "upub":
		// p2wpkh-in-p2sh, constructed as per https://github.com/bitcoin/bips/blob/master/bip-0141.mediawiki#p2wpkh-nested-in-bip16-p2sh .
		scriptSig := append([]byte{0x00, 0x14}, requestedAddress.Hash160()[:]...)
		finalAddress, err = btcutil.NewAddressScriptHashFromHash(
			btcutil.Hash160(scriptSig),
			chainParams,
		)
	case "zpub", "vpub":
		// p2wpkh
		finalAddress, err = btcutil.NewAddressWitnessPubKeyHash(
			requestedAddress.Hash160()[:],
			chainParams,
		)
	}
	if err != nil {
		return "", fmt.Errorf(
			"failed to derive final address format from extended key: [%s]",
			err,
		)
	}

	return finalAddress.EncodeAddress(), nil
}

// ValidateAddress checks to see if the supplied btc address is valid on the
// supplied chain. We check both raw btc addresses and *pub extended keys.
func ValidateAddress(btcAddress string, chainParams *chaincfg.Params) error {
	decodedAddress, decodeErr := btcutil.DecodeAddress(btcAddress, chainParams)
	if decodeErr != nil {
		_, deriveErr := DeriveAddress(btcAddress, 0)
		if deriveErr != nil {
			return fmt.Errorf(
				"[%s] is not a valid btc address using chain [%s]: "+
					"decode address failed with [%v] and "+
					"derive address failed with [%v]",
				btcAddress,
				chainParams.Name,
				decodeErr,
				deriveErr,
			)
		}
		return nil
	}

	if !decodedAddress.IsForNet(chainParams) {
		return fmt.Errorf(
			"provided address [%s] is not a valid btc address for chain [%s]",
			btcAddress,
			chainParams.Name,
		)
	}

	return nil
}
