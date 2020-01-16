package testdata

import (
	cecdsa "crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

// Sample keep addresses.
var (
	KeepAddress1 = (eth.KeepAddress)(common.HexToAddress("0x770a9E2F2Aa1eC2d3Ca916Fc3e6A55058A898632"))
	KeepAddress2 = (eth.KeepAddress)(common.HexToAddress("0x8B3BccB3A3994681A1C1584DE4b4E8b23ed1Ed6d"))
)

// KeepSigners holds signers initialized for each registered keep address.
var KeepSigners = map[eth.KeepAddress]*ecdsa.Signer{
	KeepAddress1: newTestSigner(big.NewInt(10)),
	KeepAddress2: newTestSigner(big.NewInt(20)),
}

func newTestSigner(privateKeyD *big.Int) *ecdsa.Signer {
	curve := secp256k1.S256()

	privateKey := new(cecdsa.PrivateKey)
	privateKey.PublicKey.Curve = curve
	privateKey.D = privateKeyD
	privateKey.PublicKey.X, privateKey.PublicKey.Y = curve.ScalarBaseMult(privateKeyD.Bytes())

	return ecdsa.NewSigner(privateKey)
}
