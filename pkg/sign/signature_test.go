package sign

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/keep-network/keep-tecdsa/pkg/utils/byteutils"
)

func TestCalculateSignature(t *testing.T) {
	signer := newTestSigner()
	hash, _ := hex.DecodeString("54a6483b8aca55c9df2a35baf71d9965ddfd623468d81d51229bd5eb7d1e1c1b")

	for i := 0; i < 100; i++ {
		signature, err := signer.CalculateSignature(crand.Reader, hash)
		if err != nil {
			t.Fatal(err)
		}

		serializedSignature, err := serializeSignature(signature)
		if err != nil {
			t.Fatal(err)
		}

		publicKey, err := crypto.SigToPub(hash, serializedSignature[:])
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
}
func TestCalculateECDSASignature(t *testing.T) {
	signer := newTestSigner()

	hash := []byte("test hash")

	sigR, sigS, err := signer.calculateECDSASignature(crand.Reader, hash)
	if err != nil {
		t.Fatal(err)
	}

	if !ecdsa.Verify(
		&signer.privateKey.PublicKey,
		hash,
		sigR,
		sigS,
	) {
		t.Errorf("signature is invalid")
	}
}

func TestFindRecoveryID(t *testing.T) {
	signer := newTestSigner()
	sigR, _ := new(big.Int).SetString("9b32c3623b6a16e87b4d3a56cd67c666c9897751e24a51518136185403b1cba2", 16)
	sigS, _ := new(big.Int).SetString("90838891021e1c7d0d1336613f24ecab703dee5ff1b6c8881bccc2c011606a35", 16)
	hash, _ := hex.DecodeString("54a6483b8aca55c9df2a35baf71d9965ddfd623468d81d51229bd5eb7d1e1c1b")

	expectedRecoveryID := 1

	recoveryID, err := signer.findRecoveryID(sigR, sigS, hash)
	if err != nil {
		t.Fatal(err)
	}

	if recoveryID != expectedRecoveryID {
		t.Fatalf(
			"unexpected recovery ID:\nexpected: [%d]\nactual:   [%d]\n",
			expectedRecoveryID,
			recoveryID,
		)
	}
}

func TestRecoverPublicKeyFromSignature(t *testing.T) {
	signer := newTestSigner()
	sigR, _ := new(big.Int).SetString("9b32c3623b6a16e87b4d3a56cd67c666c9897751e24a51518136185403b1cba2", 16)
	sigS, _ := new(big.Int).SetString("90838891021e1c7d0d1336613f24ecab703dee5ff1b6c8881bccc2c011606a35", 16)
	hash, _ := hex.DecodeString("54a6483b8aca55c9df2a35baf71d9965ddfd623468d81d51229bd5eb7d1e1c1b")

	x0, _ := new(big.Int).SetString("657282135ed640b0f5a280874c7e7ade110b5c3db362e0552e6b7fff2cc84593", 16)
	y0, _ := new(big.Int).SetString("28850039b734db7629c31567d7fc5677536b7fc504e967dc11f3f2289d3d4051", 16)
	expectedPublicKey0 := &PublicKey{signer.Curve(), x0, y0}

	x1, _ := new(big.Int).SetString("2f01e5e15cca351daff3843fb70f3c2f0a1bdd05e5af888a67784ef3e10a2a01", 16)
	y1, _ := new(big.Int).SetString("5c4da8a741539949293d082a132d13b4c2e213d6ba5b7617b5da2cb76cbde904", 16)
	expectedPublicKey1 := &PublicKey{signer.Curve(), x1, y1}

	iteration := int64(0)

	publicKey, err := recoverPublicKeyFromSignature(signer.Curve(), sigR, sigS, hash, iteration)
	if err != nil {
		t.Fatal(err)
	}

	if expectedPublicKey0.X.Cmp(publicKey.X) != 0 ||
		expectedPublicKey0.Y.Cmp(publicKey.Y) != 0 {
		t.Fatalf(
			"unexpected public key for iteration 0\nexpected: [%x]\nactual:   [%x]\n",
			expectedPublicKey0,
			publicKey,
		)
	}

	iteration = 1

	publicKey, err = recoverPublicKeyFromSignature(signer.Curve(), sigR, sigS, hash, iteration)
	if err != nil {
		t.Fatal(err)
	}

	if expectedPublicKey1.X.Cmp(publicKey.X) != 0 ||
		expectedPublicKey1.Y.Cmp(publicKey.Y) != 0 {
		t.Fatalf(
			"unexpected public key for iteration 1\nexpected: [%x]\nactual:   [%x]\n",
			expectedPublicKey0,
			publicKey,
		)
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

func newTestSigner() *Signer {
	curve := secp256k1.S256()
	k := big.NewInt(8)

	privateKey := new(ecdsa.PrivateKey)
	privateKey.PublicKey.Curve = curve
	privateKey.D = k
	privateKey.PublicKey.X, privateKey.PublicKey.Y = curve.ScalarBaseMult(k.Bytes())

	return NewSigner(privateKey)
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

	v := byte(signature.V)

	serializedBytes = append(serializedBytes, r...)
	serializedBytes = append(serializedBytes, s...)
	serializedBytes = append(serializedBytes, v)

	// copy(serialized[:], serializedBytes)

	return serializedBytes, nil
}
