package btc

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/keep-network/keep-tecdsa/pkg/sign"
)

// SetSignatureWitnessToTransaction sets a pay-to-witness-public-key-hash (P2WPKH)
// witness for a transaction. Witness contains a signature and a public
// key according to [BIP-141].
//
// [BIP-141]: https://github.com/bitcoin/bips/blob/master/bip-0141.mediawiki#p2wpkh
func SetSignatureWitnessToTransaction(
	signature *sign.Signature,
	publicKey *sign.PublicKey,
	inputIndex int,
	msgTx *wire.MsgTx,
) {
	hashType := txscript.SigHashAll

	btcecSignature := &btcec.Signature{R: signature.R, S: signature.S}

	sig := append(btcecSignature.Serialize(), byte(hashType))

	pkData := (*btcec.PublicKey)(publicKey).SerializeCompressed()

	txWitness := wire.TxWitness{sig, pkData}

	msgTx.TxIn[inputIndex].Witness = txWitness
}
