package tbtc

import (
	"bytes"
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-ecdsa/pkg/chain/local"
)

func TestRetrievePubkeyExtension_TimeoutElapsed(t *testing.T) {
	ctx := context.Background()
	tbtcChain := local.NewTBTCLocalChain()
	extensionsManager := &extensionsManager{tbtcChain}

	timeout := 500 * time.Millisecond
	depositAddress := "0xa5FA806723A7c7c8523F33c39686f20b52612877"

	err := extensionsManager.initializeRetrievePubkeyExtension(ctx, timeout)
	if err != nil {
		t.Fatal(err)
	}

	tbtcChain.CreateDeposit(depositAddress)

	keepAddress, err := tbtcChain.KeepAddress(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	var keepPubkey [64]byte
	rand.Read(keepPubkey[:])

	err = tbtcChain.SubmitKeepPublicKey(
		common.HexToAddress(keepAddress),
		keepPubkey,
	)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the timeout because the action must be completed
	time.Sleep(2 * timeout)

	depositPubkey, err := tbtcChain.DepositPubkey(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(keepPubkey[:], depositPubkey) {
		t.Fatalf(
			"unexpected public key\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			keepPubkey,
			depositPubkey,
		)
	}
}

// TODO: Implement other scenarios:
//  - stop event triggered
//  - keep closed
//  - keep terminated
//  - tx failed
//  - ctx cancelled
