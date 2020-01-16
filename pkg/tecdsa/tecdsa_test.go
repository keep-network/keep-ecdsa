package tecdsa

import (
	cecdsa "crypto/ecdsa"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/local"
)

var (
	keepAddress = common.HexToAddress("0x770a9E2F2Aa1eC2d3Ca916Fc3e6A55058A898632")
)

func TestGenerateSignerForKeep(t *testing.T) {
	tecdsa, chain, err := initialize()
	if err != nil {
		t.Fatalf("failed to initialize: [%v]", err)
	}

	signer, err := tecdsa.GenerateSignerForKeep(keepAddress)
	if err != nil {
		t.Fatalf("failed to generate signer: [%v]", err)
	}

	expectedPublicKey, err := eth.SerializePublicKey(signer.PublicKey())
	if err != nil {
		t.Fatalf("failed to serialize public key: [%v]", err)
	}

	publicKey, err := chain.GetKeepPublicKey(keepAddress)
	if err != nil {
		t.Fatalf("failed to get public key: [%v]", err)
	}

	if !reflect.DeepEqual(expectedPublicKey, publicKey) {
		t.Errorf(
			"unexpected public key\nexpected: [%x]\nactual:   [%x]",
			expectedPublicKey,
			publicKey,
		)
	}
}

func TestRegisterForSignEvents(t *testing.T) {
	digest := [32]byte{7, 8, 9}

	tecdsa, chain, err := initialize()
	if err != nil {
		t.Fatalf("failed to initialize: [%v]", err)
	}

	signer, err := generateSigner()
	if err != nil {
		t.Fatalf("failed to generate signer: [%v]", err)
	}

	tecdsa.RegisterForSignEvents(keepAddress, signer)

	err = chain.RequestSignature(keepAddress, digest)
	if err != nil {
		t.Fatalf("failed to request signature: [%v]", err)
	}

	time.Sleep(100 * time.Millisecond)

	signature, err := chain.GetSignature(keepAddress, digest)
	if err != nil {
		t.Fatalf("failed to get signature: [%v]", err)
	}

	if !cecdsa.Verify(
		(*cecdsa.PublicKey)(signer.PublicKey()),
		digest[:],
		signature.R,
		signature.S,
	) {
		t.Fatal("invalid signature")
	}
}

func initialize() (*TECDSA, *local.LocalChain, error) {
	chain := local.Connect().(*local.LocalChain)

	keepMembers := []common.Address{chain.Address()}

	if err := chain.CreateKeep(keepAddress, keepMembers); err != nil {
		return nil, nil, fmt.Errorf("failed to create keep: [%v]", err)
	}

	return &TECDSA{EthereumChain: chain}, chain, nil
}
