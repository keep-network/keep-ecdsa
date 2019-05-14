package btc

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/keep-network/keep-tecdsa/internal/testdata"
	"github.com/keep-network/keep-tecdsa/pkg/btc/chain/local"
	"github.com/keep-network/keep-tecdsa/pkg/sign"
)

func TestSignAndPublishTransaction(t *testing.T) {
	witnessSignatureHash, _ := hex.DecodeString(testdata.ValidTx.WitnessSignatureHash)
	transactionPreimage, _ := hex.DecodeString(testdata.ValidTx.UnsignedRaw)

	signer, err := signer()
	if err != nil {
		t.Fatal(err)
	}

	chain := local.Connect()
	err = SignAndPublishTransaction(crand.Reader, chain, signer, witnessSignatureHash, transactionPreimage)
	if err != nil {
		t.Error(err)
	}
}

func signer() (*sign.Signer, error) {
	wif, err := btcutil.DecodeWIF("923CjseKgQf7Xz185dmYUJer9i8jsb9Cd18Rtec4DFKeiBZg3wi")
	if err != nil {
		return nil, fmt.Errorf("cannot decode WIF [%s]", err)
	}
	privateKey := (*ecdsa.PrivateKey)(wif.PrivKey)

	return sign.NewSigner(privateKey), nil
}
