package recovery

import (
	"bytes"
	cecdsa "crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
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
	if len(script) != 25 {
		return nil, fmt.Errorf(
			"error deriving p2wpkh scriptCode from public key: [unexpected scriptCode length: %v]",
			len(script),
		)
	}

	// End goal here is a scriptCode that looks like
	// 0x1976a914{20-byte-pubkey-hash}88ac . 0x19 should be the length of the
	// script.
	return append([]byte{byte(len(script))}, script...), nil
}

// ConstructUnsignedTransaction produces an unsigned transaction
func ConstructUnsignedTransaction(
	previousTransactionHashHex string,
	previousOutputIndex uint32,
	previousOutputValue int64,
	feePerVbyte int64,
	recipientAddresses []string,
	chainParams *chaincfg.Params,
) (*wire.MsgTx, error) {
	// If the previous output transaction hash is passed as a []byte, can use
	// chainhash.NewHash.
	previousOutputTransactionHash, err := chainhash.NewHashFromStr(
		previousTransactionHashHex,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error decoding outpoint transaction hash: [%s]",
			previousOutputTransactionHash,
		)
	}

	// The witness signature field is the DER signature followed by the hash type.
	// We write a dummy signature with 73 0 bytes. DER signatures vary in encoding
	// between 71, 72, and 73 bytes, so we choose the longest for fee purposes.
	// We then add one more dummy byte for the SigHashType for a total of 74 bytes.
	dummySignatureForWitness := bytes.Repeat([]byte{0}, 74)

	// The compressed public key requires 33 bytes.
	dummyCompressedPublicKeyForWitness := bytes.Repeat([]byte{0}, 33)

	tx := wire.NewMsgTx(wire.TxVersion)
	txIn := wire.NewTxIn(
		wire.NewOutPoint(previousOutputTransactionHash, previousOutputIndex),
		[]byte{}, // scriptSig is empty here
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
				"error decoding recipient address [%s]: [%s]",
				recipientAddress,
				err,
			)
		}
		outputScript, err := txscript.PayToAddrScript(address)
		if err != nil {
			return nil, fmt.Errorf(
				"error constructing script from recipient address [%s]: [%s]",
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
	vsize := mempool.GetTxVirtualSize(btcutil.NewTx(tx))
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
	publicKey *cecdsa.PublicKey,
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
	// around a strings. Builder to get a hex string.
	transactionHexBuilder := &strings.Builder{}
	transactionWriter := hex.NewEncoder(transactionHexBuilder)
	// We use BtcEncode instead of Serialize here since we're preparing for the
	// transaction to be sent out of our network or executed on the bitcoin
	// blockchain, rather than persisting the information. For more information,
	// check out the btcsuite/btcd/wire/msgtx.go documentation.
	signedTransaction.BtcEncode(
		transactionWriter,
		wire.ProtocolVersion,
		wire.WitnessEncoding,
	)

	return transactionHexBuilder.String(), nil
}