package ecdsa

import (
	cecdsa "crypto/ecdsa"
	crand "crypto/rand"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/keep-network/keep-tecdsa/pkg/utils/byteutils"
)

func TestCalculateSignature(t *testing.T) {
	privateKey, err := GenerateKey(crand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer := NewSigner(privateKey)

	hash, _ := hex.DecodeString("54a6483b8aca55c9df2a35baf71d9965ddfd623468d81d51229bd5eb7d1e1c1b")

	signature, err := signer.CalculateSignature(crand.Reader, hash)
	if err != nil {
		t.Fatal(err)
	}

	serializedSignature, err := serializeSignature(signature)
	if err != nil {
		t.Fatal(err)
	}

	publicKey, err := crypto.SigToPub(hash, serializedSignature)
	if err != nil {
		t.Fatal(err)
	}

	if signer.PublicKey().X.Cmp(publicKey.X) != 0 ||
		signer.PublicKey().Y.Cmp(publicKey.Y) != 0 {
		t.Fatalf(
			"unexpected public key:\nexpected: [%x]\nactual:   [%x]\n",
			signer.PublicKey(),
			publicKey,
		)
	}
}

func TestFindRecoveryID(t *testing.T) {
	curve := secp256k1.S256()
	k := big.NewInt(8)

	privateKey := new(cecdsa.PrivateKey)
	privateKey.PublicKey.Curve = curve
	privateKey.D = k
	privateKey.PublicKey.X, privateKey.PublicKey.Y = curve.ScalarBaseMult(k.Bytes())

	signer := NewSigner(privateKey)

	hash, _ := hex.DecodeString("54a6483b8aca55c9df2a35baf71d9965ddfd623468d81d51229bd5eb7d1e1c1b")

	var tests = map[string]struct {
		sigR               string
		sigS               string
		expectedRecoveryID int
	}{
		"recovers recovery ID 0": {
			sigR:               "c896280f80f74055910baa4e4e6cfd52d73e1f7deba19e0dcbb01218dc49a729",
			sigS:               "3628a6016d1afb333f6fe7ea65a3e255219b23734d2be50088318ca0e2b2e119",
			expectedRecoveryID: 0,
		},
		"recovers recovery ID 1": {
			sigR:               "9b32c3623b6a16e87b4d3a56cd67c666c9897751e24a51518136185403b1cba2",
			sigS:               "90838891021e1c7d0d1336613f24ecab703dee5ff1b6c8881bccc2c011606a35",
			expectedRecoveryID: 1,
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			sigR, _ := new(big.Int).SetString(test.sigR, 16)
			sigS, _ := new(big.Int).SetString(test.sigS, 16)

			recoveryID, err := signer.findRecoveryID(sigR, sigS, hash)
			if err != nil {
				t.Fatal(err)
			}

			if recoveryID != test.expectedRecoveryID {
				t.Fatalf(
					"unexpected recovery ID:\nexpected: [%d]\nactual:   [%d]\n",
					test.expectedRecoveryID,
					recoveryID,
				)
			}
		})
	}
}

func TestCalculateY(t *testing.T) {
	curve := secp256k1.S256()

	x, _ := new(big.Int).SetString("657282135ed640b0f5a280874c7e7ade110b5c3db362e0552e6b7fff2cc84593", 16)
	expectedY, _ := new(big.Int).SetString("d77affc648cb2489d63cea982803a988ac94803afb169823ee0c0dd662c2bbde", 16)

	y := calculateY(curve, x)

	if !curve.IsOnCurve(x, y) {
		t.Errorf("point [%x, %x] is not on the curve", x, y)
	}

	if expectedY.Cmp(y) != 0 {
		t.Fatalf(
			"unexpected y coordinate\nexpected: [%x]\nactual:   [%x]\n",
			expectedY,
			y,
		)
	}
}

func serializeSignature(signature *Signature) ([]byte, error) {
	var serializedBytes []byte

	r, err := byteutils.LeftPadTo32Bytes(signature.R.Bytes())
	if err != nil {
		return nil, err
	}

	s, err := byteutils.LeftPadTo32Bytes(signature.S.Bytes())
	if err != nil {
		return nil, err
	}

	recoveryID := byte(signature.RecoveryID)

	serializedBytes = append(serializedBytes, r...)
	serializedBytes = append(serializedBytes, s...)
	serializedBytes = append(serializedBytes, recoveryID)

	return serializedBytes, nil
}
