package cmd

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
)

// Signatures should match a signature format on mycrypto.com. A signing/verification
// tool available on https://mycrypto.com/sign-and-verify-message can be used to
// cross-check correctness of signatures provided by our implementation.

var validSignature = EthereumSignature{
	Address:   common.HexToAddress("0x4BCFC3099F12C53D01Da46695CC8776be584b946"),
	Message:   "verySecretMessage",
	Signature: "0xc8be189ab0ee691de7019eaa3de58558b84775085d9a0840908343ac690e02ca3f6e3d2dc70025b9b214d96c30e38c41f818cccd6f06b7a81c4afd26cbe6d6d600",
	Version:   "2",
}

func TestSign(t *testing.T) {
	message := "verySecretMessage"
	keyFilePath := "../internal/testdata/eth_key.json"
	keyFilePassword := "password"

	expectedResult := &EthereumSignature{
		Address:   common.HexToAddress("0x4BCFC3099F12C53D01Da46695CC8776be584b946"),
		Message:   message,
		Signature: "0xc8be189ab0ee691de7019eaa3de58558b84775085d9a0840908343ac690e02ca3f6e3d2dc70025b9b214d96c30e38c41f818cccd6f06b7a81c4afd26cbe6d6d600",
		Version:   "2",
	}

	ethereumKey, err := ethutil.DecryptKeyFile(keyFilePath, keyFilePassword)
	if err != nil {
		t.Fatalf(
			"failed to read key file [%s]: [%v]",
			keyFilePath,
			err,
		)
	}

	ethereumSignature, err := sign(ethereumKey, message)
	if err != nil {
		t.Errorf("signing failed: [%v]", err)
	}

	if !reflect.DeepEqual(ethereumSignature, expectedResult) {
		t.Errorf(
			"unexpected signature\nexpected: %v\nactual:   %v",
			expectedResult,
			ethereumSignature,
		)
	}
}

func TestVerify_V0(t *testing.T) {
	err := verify(&validSignature)
	if err != nil {
		t.Errorf("unexpected error: [%v]", err)
	}
}

func TestVerify_V27(t *testing.T) {
	// go-ethereum library produces a signature with V value of 0 or 1. In some
	// chains the V value is expected to be 27 or 28. Even ethereum is sometimes
	// inconsistent about that across their libraries. In our implementation we
	// expect V to be 0 or 1, we're not currently supporting 27 or 28.
	ethereumSignature := validSignature
	ethereumSignature.Signature = "0xc8be189ab0ee691de7019eaa3de58558b84775085d9a0840908343ac690e02ca3f6e3d2dc70025b9b214d96c30e38c41f818cccd6f06b7a81c4afd26cbe6d6d61b"

	expectedError := fmt.Errorf("could not recover public key from signature [invalid signature recovery id]")

	err := verify(&ethereumSignature)
	if !reflect.DeepEqual(expectedError, err) {
		t.Errorf("unexpected error\nexpected: [%v]\nactual:   [%v]", expectedError, err)
	}
}

func TestVerify_WrongAddress(t *testing.T) {
	ethereumSignature := validSignature
	ethereumSignature.Address = common.HexToAddress("0x93df7c54c41A9D7FB17C1E8039d387a2A924708c")

	expectedError := fmt.Errorf("invalid signer\n\texpected: 0x93df7c54c41A9D7FB17C1E8039d387a2A924708c\n\tactual:   0x4BCFC3099F12C53D01Da46695CC8776be584b946")

	err := verify(&ethereumSignature)
	if !reflect.DeepEqual(expectedError, err) {
		t.Errorf("unexpected error\nexpected: [%v]\nactual:   [%v]", expectedError, err)
	}
}

func TestVerify_WrongMessage(t *testing.T) {
	ethereumSignature := validSignature
	ethereumSignature.Message = "notTheSignedMessage"

	expectedError := fmt.Errorf("invalid signer\n\texpected: 0x4BCFC3099F12C53D01Da46695CC8776be584b946\n\tactual:   0x19882d7da145A10d5AEEFEe217Fd87dE679b4bb1")

	err := verify(&ethereumSignature)
	if !reflect.DeepEqual(expectedError, err) {
		t.Errorf("unexpected error\nexpected: [%v]\nactual:   [%v]", expectedError, err)
	}
}

func TestVerify_WrongSignature(t *testing.T) {
	ethereumSignature := validSignature
	ethereumSignature.Signature = "0xc8be189ab0ee691de7019eaa3de58558b84775085d9a0840908343ac690e02ca3f6e3d2dc70025b9b214d96c30e38c41f818cccd6f06b7a81c4afd26cbe6d6d601"

	expectedError := fmt.Errorf("invalid signer\n\texpected: 0x4BCFC3099F12C53D01Da46695CC8776be584b946\n\tactual:   0xb560e6c746138528509de08B782E3144E031a6B1")

	err := verify(&ethereumSignature)
	if !reflect.DeepEqual(expectedError, err) {
		t.Errorf("unexpected error\nexpected: [%v]\nactual:   [%v]", expectedError, err)
	}
}

func TestVerify_WrongVersion(t *testing.T) {
	ethereumSignature := validSignature
	ethereumSignature.Version = "1"

	expectedError := fmt.Errorf("unsupported ethereum signature version\n\texpected: 2\n\tactual:   1")

	err := verify(&ethereumSignature)
	if !reflect.DeepEqual(expectedError, err) {
		t.Errorf("unexpected error\nexpected: [%v]\nactual:   [%v]", expectedError, err)
	}
}
