package tbtc

import (
	cecdsa "crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
)

// PublicKeyToP2WPKHScriptCode converts a public key to a Bitcion p2wpkh
// witness scriptCode that can spend an output sent to that public key's
// corresponding address.
//
// [BIP143]: https://github.com/bitcoin/bips/blob/master/bip-0143.mediawiki
func PublicKeyToP2WPKHScriptCode(
	publicKey *cecdsa.PublicKey,
	chainParams *chaincfg.Params,
) ([]byte, error) {
	// ecdsa.PublicKey and btcec.PublicKey are both method attachments to
	// Go's crypto/ecdsa.PublicKey, so we can cast.
	publicKeyBytes := (*btcec.PublicKey)(publicKey).SerializeCompressed()
	// Note that the scriptCode for a p2wpkh address is the equivalent of the
	// p2pkh scriptPubKey.
	pubKeyAddress, err := btcutil.NewAddressPubKey(publicKeyBytes, chainParams)
	if err != nil {
		return nil, fmt.Errorf(
			"error deriving p2wpkh scriptCode from public key: [%s]",
			err,
		)
	}
	pkhAddress := pubKeyAddress.AddressPubKeyHash()

	script, err := txscript.PayToAddrScript(pkhAddress)
	if err != nil {
		return nil, fmt.Errorf(
			"error deriving p2wpkh scriptCode from public key: [%s]",
			err,
		)
	}
	if len(script) > 255 {
		return nil, fmt.Errorf(
			"error deriving p2wpkh scriptCode from public key: [scriptCode too long: %v]",
			len(script),
		)
	}

	// End goal here is a scriptCode that looks like
	// 0x1976a914{20-byte-pubkey-hash}88ac . 0x19 should be the length of the
	// script.
	return append([]byte{byte(len(script))}, script...), nil
}

// DeriveAddress uses the specified extended public key and address index to
// derive an address string in the appropriate format at the specified address
// index. The extended public key is expected to already be at the account path
// level (e.g. m/44'/0'/0' for [BIP44] xpubs), and addresses are derived within
// the external chain subpath (`/0`). Thus, calling DeriveAddress with an xpub
// generated at m/44'/0'/0' and passing the address index 5 will produce the
// address at path m/44'/0'/0'/0/5.
//
// In cases where the extended public key is at depth 4, meaning the external or
// internal chain is already included, DeriveAddress will directly derive the
// address index at the existing depth.
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
func DeriveAddress(extendedPublicKey string, addressIndex int) (string, error) {
	extendedKey, err := hdkeychain.NewKeyFromString(extendedPublicKey)
	if err != nil {
		return "", fmt.Errorf(
			"error parsing extended public key: [%s]",
			err,
		)
	}
	// For later usage---this is xpub/ypub/zpub/...
	publicKeyDescriptor := extendedPublicKey[0:4]

	externalChain := extendedKey
	if externalChain.Depth() < 4 {
		// Descend to the external chain path, /0.
		externalChain, err = extendedKey.Child(0)
		if err != nil {
			return "", fmt.Errorf(
				"error deriving external chain path /0 from extended key: [%s]",
				err,
			)
		}
	}

	requestedPublicKey, err := externalChain.Child(uint32(addressIndex))
	if err != nil {
		return "", fmt.Errorf(
			"error deriving requested address index /0/%v from extended key: [%s]",
			addressIndex,
			err,
		)
	}

	// Now to decide how we want to serialize the address...
	var chainParams *chaincfg.Params
	switch publicKeyDescriptor {
	case "xpub", "ypub", "zpub":
		chainParams = &chaincfg.MainNetParams
	case "tpub", "upub", "vpub":
		chainParams = &chaincfg.TestNet3Params
	}

	requestedAddress, err := requestedPublicKey.Address(chainParams)

	var finalAddress btcutil.Address = requestedAddress
	switch publicKeyDescriptor {
	case "xpub", "tpub":
		// Noop, the address is already correct
	case "ypub", "upub":
		// p2wpkh-in-p2sh, constructed as per https://github.com/bitcoin/bips/blob/master/bip-0141.mediawiki#p2wpkh-nested-in-bip16-p2sh .
		scriptSig := append([]byte{0x00, 0x14}, requestedAddress.Hash160()[:]...)
		finalAddress, err = btcutil.NewAddressWitnessScriptHash(
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

// ConstructUnsignedTransaction produces an unsigned transaction
func ConstructUnsignedTransaction(
	previousOutputTransactionHashHex string,
	previousOutputIndex uint32,
	previousOutputValue int64,
	feePerVbyte int64,
	recipientAddresses []string,
	chainParams *chaincfg.Params,
) (*wire.MsgTx, error) {
	// If the previous output transaction hash is passed as a []byte, can use
	// chainhash.NewHash.
	previousOutputTransactionHash, err := chainhash.NewHashFromStr(
		previousOutputTransactionHashHex,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error extracting outpoint transaction hash: [%s]",
			previousOutputTransactionHash,
		)
	}

	// The witness signature field is the DER signature followed by the hash type.
	// We write a dummy signature with 73 0 bytes. DER signatures vary in encoding
	// between 71, 72, and 73 bytes, so we choose the longest for fee purposes.
	dummySignatureForWitness := []byte{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0,
		0, // one more dummy byte for the SigHashType
	}
	// The compressed public key requires 33 bytes.
	dummyCompressedPublicKeyForWitness := []byte{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0,
	}

	tx := wire.NewMsgTx(wire.TxVersion)
	txIn := wire.NewTxIn(
		wire.NewOutPoint(previousOutputTransactionHash, previousOutputIndex),
		nil, // scriptSig is empty here
		[][]byte{
			dummySignatureForWitness,
			dummyCompressedPublicKeyForWitness,
		},
	)
	txIn.Sequence = 0
	tx.AddTxIn(txIn)

	for _, recipientAddress := range recipientAddresses {
		address, err := btcutil.DecodeAddress(recipientAddress, chainParams)
		if err != nil {
			return nil, fmt.Errorf(
				"error decoding output address [%s]: [%s]",
				recipientAddress,
				err,
			)
		}
		outputScript, err := txscript.PayToAddrScript(address)
		if err != nil {
			return nil, fmt.Errorf(
				"error constructing script from output address [%s]: [%s]",
				recipientAddress,
				err,
			)
		}

		tx.AddTxOut(wire.NewTxOut(
			int64(0), // value is filled in after fee is computed below
			outputScript,
		))
	}

	// Compute weight and vsize per [BIP141], except vsize is truncated
	// instead of rounded up, then compute the final fee and set the
	// per-recipient value. Could result in a fractionally low fee.
	weight := tx.SerializeSizeStripped()*3 + tx.SerializeSize()
	vsize := weight / 4
	fee := feePerVbyte * int64(vsize)
	perRecipientValue := (previousOutputValue - fee) / int64(len(recipientAddresses))
	for _, txOut := range tx.TxOut {
		txOut.Value = perRecipientValue
	}

	return tx, nil
}

// BuildSignedTransactionHexString generates the final transaction hex string
// that can then be submitted to the chain
func BuildSignedTransactionHexString(
	unsignedTransaction *wire.MsgTx,
	signature *ecdsa.Signature,
	publicKey *ecdsa.PublicKey,
) (string, error) {
	// For safety's sake, work on a deep copy, as mutations follow.
	signedTransaction := unsignedTransaction.Copy()

	btcSignature := &btcec.Signature{R: signature.R, S: signature.S}

	// The witness is for the first input, since this is known to be a
	// single-input transaction.
	signedTransaction.TxIn[0].Witness = wire.TxWitness{
		// The witness signature field is the DER signature followed by the hash type.
		append(btcSignature.Serialize(), byte(txscript.SigHashAll)),
		// The second part of the witness is the compressed public key.
		(*btcec.PublicKey)(publicKey).SerializeCompressed(),
	}

	// BtcEncode writes bytes, we wrap it in an hex encoder wrapped
	// around a strings.Builder to get a hex string.
	transactionHexBuilder := &strings.Builder{}
	transactionWriter := hex.NewEncoder(transactionHexBuilder)
	signedTransaction.BtcEncode(
		transactionWriter,
		wire.ProtocolVersion,
		wire.WitnessEncoding,
	)

	return transactionHexBuilder.String(), nil
}
