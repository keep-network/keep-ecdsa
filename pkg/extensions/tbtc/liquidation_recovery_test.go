package tbtc

import (
	cecdsa "crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
)

func Test_PublicKeyToP2WPKHScriptCode_Works(t *testing.T) {
	curve := elliptic.P224()
	privateKey, _ := cecdsa.GenerateKey(curve, rand.Reader)
	scriptCodeBytes, _ := PublicKeyToP2WPKHScriptCode(&privateKey.PublicKey, &chaincfg.TestNet3Params)

	if len(scriptCodeBytes) != 25 {
		t.Errorf("The script code must be exactly 26 bytes long. Instead, it was %v", len(scriptCodeBytes))
	}
}
