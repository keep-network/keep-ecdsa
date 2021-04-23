package tbtc

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/keep-network/keep-common/pkg/subscription"

	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/local"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-ecdsa/pkg/utils/byteutils"
)

const (
	timeout                        = 500 * time.Millisecond
	depositAddress                 = "0xa5FA806723A7c7c8523F33c39686f20b52612877"
	defaultLocalBlockConfirmations = 0
)

func newTestTBTC(
	localChain *local.TBTCLocalChain,
) *tbtc {
	tbtc := newTBTC(
		localChain,
		localChain.BlockCounter(),
		localChain.BlockTimestamp,
	)

	tbtc.blockConfirmations = defaultLocalBlockConfirmations

	return tbtc
}

func TestRetrievePubkey_TimeoutElapsed(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorRetrievePubKey(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	keepPubkey, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedRetrieveSignerPubkeyCalls := 1
	actualRetrieveSignerPubkeyCalls := tbtcChain.Logger().
		RetrieveSignerPubkeyCalls()
	if expectedRetrieveSignerPubkeyCalls != actualRetrieveSignerPubkeyCalls {
		t.Errorf(
			"unexpected number of RetrieveSignerPubkey calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedRetrieveSignerPubkeyCalls,
			actualRetrieveSignerPubkeyCalls,
		)
	}

	depositPubkey, err := tbtcChain.DepositPubkey(depositAddress)
	if err != nil {
		t.Errorf(
			"unexpected error while fetching deposit pubkey: [%v]",
			err,
		)
	}

	if !bytes.Equal(keepPubkey[:], depositPubkey) {
		t.Errorf(
			"unexpected public key\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			keepPubkey,
			depositPubkey,
		)
	}
}

func TestRetrievePubkey_StopEventOccurred(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorRetrievePubKey(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	keepPubkey, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// wait a while before triggering the stop event because the
	// extension must have time to handle the start event
	time.Sleep(100 * time.Millisecond)

	// invoke the action which will trigger the stop event in result
	err = tbtcChain.RetrieveSignerPubkey(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedRetrieveSignerPubkeyCalls := 1
	actualRetrieveSignerPubkeyCalls := tbtcChain.Logger().
		RetrieveSignerPubkeyCalls()
	if expectedRetrieveSignerPubkeyCalls != actualRetrieveSignerPubkeyCalls {
		t.Errorf(
			"unexpected number of RetrieveSignerPubkey calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedRetrieveSignerPubkeyCalls,
			actualRetrieveSignerPubkeyCalls,
		)
	}

	depositPubkey, err := tbtcChain.DepositPubkey(depositAddress)
	if err != nil {
		t.Errorf(
			"unexpected error while fetching deposit pubkey: [%v]",
			err,
		)
	}

	if !bytes.Equal(keepPubkey[:], depositPubkey) {
		t.Errorf(
			"unexpected public key\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			keepPubkey,
			depositPubkey,
		)
	}
}

func TestRetrievePubkey_KeepClosedEventOccurred(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorRetrievePubKey(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// wait a while before triggering the keep closed event because the
	// extension must have time to handle the start event
	time.Sleep(100 * time.Millisecond)

	err = closeKeep(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedRetrieveSignerPubkeyCalls := 0
	actualRetrieveSignerPubkeyCalls := tbtcChain.Logger().
		RetrieveSignerPubkeyCalls()
	if expectedRetrieveSignerPubkeyCalls != actualRetrieveSignerPubkeyCalls {
		t.Errorf(
			"unexpected number of RetrieveSignerPubkey calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedRetrieveSignerPubkeyCalls,
			actualRetrieveSignerPubkeyCalls,
		)
	}

	_, err = tbtcChain.DepositPubkey(depositAddress)

	expectedError := fmt.Errorf(
		"no pubkey for deposit [%v]",
		depositAddress,
	)
	if !reflect.DeepEqual(expectedError, err) {
		t.Errorf(
			"unexpected error\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedError,
			err,
		)
	}
}

func TestRetrievePubkey_KeepTerminatedEventOccurred(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorRetrievePubKey(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// wait a while before triggering the keep terminated event because the
	// extension must have time to handle the start event
	time.Sleep(100 * time.Millisecond)

	err = terminateKeep(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedRetrieveSignerPubkeyCalls := 0
	actualRetrieveSignerPubkeyCalls := tbtcChain.Logger().
		RetrieveSignerPubkeyCalls()
	if expectedRetrieveSignerPubkeyCalls != actualRetrieveSignerPubkeyCalls {
		t.Errorf(
			"unexpected number of RetrieveSignerPubkey calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedRetrieveSignerPubkeyCalls,
			actualRetrieveSignerPubkeyCalls,
		)
	}

	_, err = tbtcChain.DepositPubkey(depositAddress)

	expectedError := fmt.Errorf(
		"no pubkey for deposit [%v]",
		depositAddress,
	)
	if !reflect.DeepEqual(expectedError, err) {
		t.Errorf(
			"unexpected error\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedError,
			err,
		)
	}
}

func TestRetrievePubkey_ActionFailed(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorRetrievePubKey(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	// do not submit the keep public key intentionally to cause
	// the action error

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedRetrieveSignerPubkeyCalls := 3
	actualRetrieveSignerPubkeyCalls := tbtcChain.Logger().
		RetrieveSignerPubkeyCalls()
	if expectedRetrieveSignerPubkeyCalls != actualRetrieveSignerPubkeyCalls {
		t.Errorf(
			"unexpected number of RetrieveSignerPubkey calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedRetrieveSignerPubkeyCalls,
			actualRetrieveSignerPubkeyCalls,
		)
	}
}

func TestRetrievePubkey_ContextCancelled_WithoutWorkingMonitoring(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorRetrievePubKey(
		ctx,
		constantBackoff,
		timeout,
	)

	// cancel the context before any start event occurs
	cancelCtx()

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedRetrieveSignerPubkeyCalls := 0
	actualRetrieveSignerPubkeyCalls := tbtcChain.Logger().
		RetrieveSignerPubkeyCalls()
	if expectedRetrieveSignerPubkeyCalls != actualRetrieveSignerPubkeyCalls {
		t.Errorf(
			"unexpected number of RetrieveSignerPubkey calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedRetrieveSignerPubkeyCalls,
			actualRetrieveSignerPubkeyCalls,
		)
	}
}

func TestRetrievePubkey_ContextCancelled_WithWorkingMonitoring(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorRetrievePubKey(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	// wait a while before cancelling the context because the
	// extension must have time to handle the start event
	time.Sleep(100 * time.Millisecond)

	// cancel the context once the start event is handled and
	// the monitoring process is running
	cancelCtx()

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedRetrieveSignerPubkeyCalls := 0
	actualRetrieveSignerPubkeyCalls := tbtcChain.Logger().
		RetrieveSignerPubkeyCalls()
	if expectedRetrieveSignerPubkeyCalls !=
		actualRetrieveSignerPubkeyCalls {
		t.Errorf(
			"unexpected number of RetrieveSignerPubkey calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedRetrieveSignerPubkeyCalls,
			actualRetrieveSignerPubkeyCalls,
		)
	}
}

func TestRetrievePubkey_OperatorNotInSigningGroup(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorRetrievePubKey(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := local.RandomSigningGroup(3)

	tbtcChain.CreateDeposit(depositAddress, signers)

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedRetrieveSignerPubkeyCalls := 0
	actualRetrieveSignerPubkeyCalls := tbtcChain.Logger().
		RetrieveSignerPubkeyCalls()
	if expectedRetrieveSignerPubkeyCalls !=
		actualRetrieveSignerPubkeyCalls {
		t.Errorf(
			"unexpected number of RetrieveSignerPubkey calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedRetrieveSignerPubkeyCalls,
			actualRetrieveSignerPubkeyCalls,
		)
	}
}

func TestProvideRedemptionSignature_TimeoutElapsed(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionSignature(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	keepSignature, err := submitKeepSignature(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedProvideRedemptionSignatureCalls := 1
	actualProvideRedemptionSignatureCalls := tbtcChain.Logger().
		ProvideRedemptionSignatureCalls()
	if expectedProvideRedemptionSignatureCalls !=
		actualProvideRedemptionSignatureCalls {
		t.Errorf(
			"unexpected number of ProvideRedemptionSignature calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedProvideRedemptionSignatureCalls,
			actualProvideRedemptionSignatureCalls,
		)
	}

	depositSignature, err := tbtcChain.DepositRedemptionSignature(
		depositAddress,
	)
	if err != nil {
		t.Errorf(
			"unexpected error while fetching deposit signature: [%v]",
			err,
		)
	}

	if !areChainSignaturesEqual(keepSignature, depositSignature) {
		t.Errorf(
			"unexpected signature\n"+
				"expected: [%+v]\n"+
				"actual:   [%+v]",
			keepSignature,
			depositSignature,
		)
	}
}

func TestProvideRedemptionSignature_StopEventOccurred_DepositGotRedemptionSignature(
	t *testing.T,
) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionSignature(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	keepSignature, err := submitKeepSignature(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// wait a while before triggering the stop event because the
	// extension must have time to handle the start event
	time.Sleep(100 * time.Millisecond)

	// invoke the action which will trigger the stop event in result
	err = tbtcChain.ProvideRedemptionSignature(
		depositAddress,
		keepSignature.V,
		keepSignature.R,
		keepSignature.S,
	)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedProvideRedemptionSignatureCalls := 1
	actualProvideRedemptionSignatureCalls := tbtcChain.Logger().
		ProvideRedemptionSignatureCalls()
	if expectedProvideRedemptionSignatureCalls !=
		actualProvideRedemptionSignatureCalls {
		t.Errorf(
			"unexpected number of ProvideRedemptionSignature calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedProvideRedemptionSignatureCalls,
			actualProvideRedemptionSignatureCalls,
		)
	}

	depositSignature, err := tbtcChain.DepositRedemptionSignature(
		depositAddress,
	)
	if err != nil {
		t.Errorf(
			"unexpected error while fetching deposit signature: [%v]",
			err,
		)
	}

	if !areChainSignaturesEqual(keepSignature, depositSignature) {
		t.Errorf(
			"unexpected signature\n"+
				"expected: [%+v]\n"+
				"actual:   [%+v]",
			keepSignature,
			depositSignature,
		)
	}
}

func TestProvideRedemptionSignature_StopEventOccurred_DepositRedeemed(
	t *testing.T,
) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionSignature(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	_, err = submitKeepSignature(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// wait a while before triggering the stop event because the
	// extension must have time to handle the start event
	time.Sleep(100 * time.Millisecond)

	// invoke the action which will trigger the stop event in result
	err = tbtcChain.ProvideRedemptionProof(
		depositAddress,
		[4]uint8{},
		nil,
		nil,
		[4]uint8{},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedProvideRedemptionSignatureCalls := 0
	actualProvideRedemptionSignatureCalls := tbtcChain.Logger().
		ProvideRedemptionSignatureCalls()
	if expectedProvideRedemptionSignatureCalls !=
		actualProvideRedemptionSignatureCalls {
		t.Errorf(
			"unexpected number of ProvideRedemptionSignature calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedProvideRedemptionSignatureCalls,
			actualProvideRedemptionSignatureCalls,
		)
	}

	depositProof, err := tbtcChain.DepositRedemptionProof(depositAddress)
	if err != nil {
		t.Errorf("unexpected error while fetching deposit proof: [%v]", err)
	}

	if depositProof == nil {
		t.Errorf("deposit proof should be provided")
	}
}

func TestProvideRedemptionSignature_KeepClosedEventOccurred(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionSignature(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	_, err = submitKeepSignature(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// wait a while before triggering the keep closed event because the
	// extension must have time to handle the start event
	time.Sleep(100 * time.Millisecond)

	err = closeKeep(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedProvideRedemptionSignatureCalls := 0
	actualProvideRedemptionSignatureCalls := tbtcChain.Logger().
		ProvideRedemptionSignatureCalls()
	if expectedProvideRedemptionSignatureCalls !=
		actualProvideRedemptionSignatureCalls {
		t.Errorf(
			"unexpected number of ProvideRedemptionSignature calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedProvideRedemptionSignatureCalls,
			actualProvideRedemptionSignatureCalls,
		)
	}

	_, err = tbtcChain.DepositRedemptionSignature(depositAddress)

	expectedError := fmt.Errorf(
		"no redemption signature for deposit [%v]",
		depositAddress,
	)
	if !reflect.DeepEqual(expectedError, err) {
		t.Errorf(
			"unexpected error\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedError,
			err,
		)
	}
}

func TestProvideRedemptionSignature_KeepTerminatedEventOccurred(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionSignature(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	_, err = submitKeepSignature(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// wait a while before triggering the keep terminated event because the
	// extension must have time to handle the start event
	time.Sleep(100 * time.Millisecond)

	err = terminateKeep(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedProvideRedemptionSignatureCalls := 0
	actualProvideRedemptionSignatureCalls := tbtcChain.Logger().
		ProvideRedemptionSignatureCalls()
	if expectedProvideRedemptionSignatureCalls !=
		actualProvideRedemptionSignatureCalls {
		t.Errorf(
			"unexpected number of ProvideRedemptionSignature calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedProvideRedemptionSignatureCalls,
			actualProvideRedemptionSignatureCalls,
		)
	}

	_, err = tbtcChain.DepositRedemptionSignature(depositAddress)

	expectedError := fmt.Errorf(
		"no redemption signature for deposit [%v]",
		depositAddress,
	)
	if !reflect.DeepEqual(expectedError, err) {
		t.Errorf(
			"unexpected error\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedError,
			err,
		)
	}
}

func TestProvideRedemptionSignature_ActionFailed(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionSignature(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	_, err = submitKeepSignature(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// simulate a situation when `ProvideRedemptionSignature` fails on-chain
	tbtcChain.SetAlwaysFailingTransactions("ProvideRedemptionSignature")

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedProvideRedemptionSignatureCalls := 3
	actualProvideRedemptionSignatureCalls := tbtcChain.Logger().
		ProvideRedemptionSignatureCalls()
	if expectedProvideRedemptionSignatureCalls !=
		actualProvideRedemptionSignatureCalls {
		t.Errorf(
			"unexpected number of ProvideRedemptionSignature calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedProvideRedemptionSignatureCalls,
			actualProvideRedemptionSignatureCalls,
		)
	}
}

func TestProvideRedemptionSignature_ContextCancelled_WithoutWorkingMonitoring(
	t *testing.T,
) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionSignature(
		ctx,
		constantBackoff,
		timeout,
	)

	// cancel the context before any start event occurs
	cancelCtx()

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedProvideRedemptionSignatureCalls := 0
	actualProvideRedemptionSignatureCalls := tbtcChain.Logger().
		ProvideRedemptionSignatureCalls()
	if expectedProvideRedemptionSignatureCalls !=
		actualProvideRedemptionSignatureCalls {
		t.Errorf(
			"unexpected number of ProvideRedemptionSignature calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedProvideRedemptionSignatureCalls,
			actualProvideRedemptionSignatureCalls,
		)
	}
}

func TestProvideRedemptionSignature_ContextCancelled_WithWorkingMonitoring(
	t *testing.T,
) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionSignature(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	// wait a while before cancelling the context because the
	// extension must have time to handle the start event
	time.Sleep(100 * time.Millisecond)

	// cancel the context once the start event is handled and
	// the monitoring process is running
	cancelCtx()

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedProvideRedemptionSignatureCalls := 0
	actualProvideRedemptionSignatureCalls := tbtcChain.Logger().
		ProvideRedemptionSignatureCalls()
	if expectedProvideRedemptionSignatureCalls !=
		actualProvideRedemptionSignatureCalls {
		t.Errorf(
			"unexpected number of ProvideRedemptionSignature calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedProvideRedemptionSignatureCalls,
			actualProvideRedemptionSignatureCalls,
		)
	}
}

func TestProvideRedemptionSignature_OperatorNotInSigningGroup(
	t *testing.T,
) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionSignature(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := local.RandomSigningGroup(3)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	_, err = submitKeepSignature(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedProvideRedemptionSignatureCalls := 0
	actualProvideRedemptionSignatureCalls := tbtcChain.Logger().
		ProvideRedemptionSignatureCalls()
	if expectedProvideRedemptionSignatureCalls !=
		actualProvideRedemptionSignatureCalls {
		t.Errorf(
			"unexpected number of ProvideRedemptionSignature calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedProvideRedemptionSignatureCalls,
			actualProvideRedemptionSignatureCalls,
		)
	}
}

func TestProvideRedemptionProof_TimeoutElapsed(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionProof(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	initialDepositRedemptionFee, err := tbtcChain.DepositRedemptionFee(
		depositAddress,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Increase the fee at least 3 times to test fee increase step determination
	// from the two latests events.
	for i := 1; i <= 3; i++ {
		keepSignature, err := submitKeepSignature(depositAddress, tbtcChain)
		if err != nil {
			t.Fatal(err)
		}

		err = tbtcChain.ProvideRedemptionSignature(
			depositAddress,
			keepSignature.V,
			keepSignature.R,
			keepSignature.S,
		)
		if err != nil {
			t.Fatal(err)
		}

		// wait a bit longer than the monitoring timeout
		// to make sure the potential transaction completes
		time.Sleep(2 * timeout)

		expectedIncreaseRedemptionFeeCalls := i
		actualIncreaseRedemptionFeeCalls := tbtcChain.Logger().
			IncreaseRedemptionFeeCalls()
		if expectedIncreaseRedemptionFeeCalls != actualIncreaseRedemptionFeeCalls {
			t.Errorf(
				"unexpected number of IncreaseRedemptionFee calls after [%d] increase\n"+
					"expected: [%v]\n"+
					"actual:   [%v]",
				i,
				expectedIncreaseRedemptionFeeCalls,
				actualIncreaseRedemptionFeeCalls,
			)
		}

		expectedDepositRedemptionFee := new(big.Int).Mul(
			new(big.Int).Add(big.NewInt(1), big.NewInt(int64(i))),
			initialDepositRedemptionFee,
		)

		actualDepositRedemptionFee, err := tbtcChain.DepositRedemptionFee(
			depositAddress,
		)
		if err != nil {
			t.Fatal(err)
		}

		if expectedDepositRedemptionFee.Cmp(actualDepositRedemptionFee) != 0 {
			t.Errorf(
				"unexpected redemption fee value after [%d] increase\n"+
					"expected: [%v]\n"+
					"actual:   [%v]",
				i,
				expectedDepositRedemptionFee.Text(10),
				actualDepositRedemptionFee.Text(10),
			)
		}
	}
}

func TestProvideRedemptionProof_StopEventOccurred_DepositRedemptionRequested(
	t *testing.T,
) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionProof(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	initialDepositRedemptionFee, err := tbtcChain.DepositRedemptionFee(
		depositAddress,
	)
	if err != nil {
		t.Fatal(err)
	}

	keepSignature, err := submitKeepSignature(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.ProvideRedemptionSignature(
		depositAddress,
		keepSignature.V,
		keepSignature.R,
		keepSignature.S,
	)
	if err != nil {
		t.Fatal(err)
	}

	// wait a while before triggering the stop event because the
	// extension must have time to handle the start event
	time.Sleep(100 * time.Millisecond)

	// invoke the action which will trigger the stop event in result
	err = tbtcChain.IncreaseRedemptionFee(
		depositAddress,
		toLittleEndianBytes(big.NewInt(990)),
		toLittleEndianBytes(big.NewInt(980)),
	)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	// Expect exactly one call of `IncreaseRedemptionFee` coming from the
	// explicit invocation placed above. The monitoring routine is not expected
	// to do any calls.
	expectedIncreaseRedemptionFeeCalls := 1
	actualIncreaseRedemptionFeeCalls := tbtcChain.Logger().
		IncreaseRedemptionFeeCalls()
	if expectedIncreaseRedemptionFeeCalls != actualIncreaseRedemptionFeeCalls {
		t.Errorf(
			"unexpected number of IncreaseRedemptionFee calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedIncreaseRedemptionFeeCalls,
			actualIncreaseRedemptionFeeCalls,
		)
	}

	expectedDepositRedemptionFee := new(big.Int).Mul(
		big.NewInt(2),
		initialDepositRedemptionFee,
	)

	actualDepositRedemptionFee, err := tbtcChain.DepositRedemptionFee(
		depositAddress,
	)
	if err != nil {
		t.Fatal(err)
	}

	if expectedDepositRedemptionFee.Cmp(actualDepositRedemptionFee) != 0 {
		t.Errorf(
			"unexpected redemption fee value\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedDepositRedemptionFee.Text(10),
			actualDepositRedemptionFee.Text(10),
		)
	}
}

func TestProvideRedemptionProof_StopEventOccurred_DepositRedeemed(
	t *testing.T,
) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionProof(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	keepSignature, err := submitKeepSignature(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.ProvideRedemptionSignature(
		depositAddress,
		keepSignature.V,
		keepSignature.R,
		keepSignature.S,
	)
	if err != nil {
		t.Fatal(err)
	}

	// wait a while before triggering the stop event because the
	// extension must have time to handle the start event
	time.Sleep(100 * time.Millisecond)

	// invoke the action which will trigger the stop event in result
	err = tbtcChain.ProvideRedemptionProof(
		depositAddress,
		[4]uint8{},
		nil,
		nil,
		[4]uint8{},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedIncreaseRedemptionFeeCalls := 0
	actualIncreaseRedemptionFeeCalls := tbtcChain.Logger().
		IncreaseRedemptionFeeCalls()
	if expectedIncreaseRedemptionFeeCalls != actualIncreaseRedemptionFeeCalls {
		t.Errorf(
			"unexpected number of IncreaseRedemptionFee calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedIncreaseRedemptionFeeCalls,
			actualIncreaseRedemptionFeeCalls,
		)
	}

	depositProof, err := tbtcChain.DepositRedemptionProof(depositAddress)
	if err != nil {
		t.Errorf("unexpected error while fetching deposit proof: [%v]", err)
	}

	if depositProof == nil {
		t.Errorf("deposit proof should be provided")
	}
}

func TestProvideRedemptionProof_KeepClosedEventOccurred(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionProof(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	initialDepositRedemptionFee, err := tbtcChain.DepositRedemptionFee(
		depositAddress,
	)
	if err != nil {
		t.Fatal(err)
	}

	keepSignature, err := submitKeepSignature(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.ProvideRedemptionSignature(
		depositAddress,
		keepSignature.V,
		keepSignature.R,
		keepSignature.S,
	)
	if err != nil {
		t.Fatal(err)
	}

	// wait a while before triggering the keep closed event because the
	// extension must have time to handle the start event
	time.Sleep(100 * time.Millisecond)

	err = closeKeep(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedIncreaseRedemptionFeeCalls := 0
	actualIncreaseRedemptionFeeCalls := tbtcChain.Logger().
		IncreaseRedemptionFeeCalls()
	if expectedIncreaseRedemptionFeeCalls != actualIncreaseRedemptionFeeCalls {
		t.Errorf(
			"unexpected number of IncreaseRedemptionFee calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedIncreaseRedemptionFeeCalls,
			actualIncreaseRedemptionFeeCalls,
		)
	}

	actualDepositRedemptionFee, err := tbtcChain.DepositRedemptionFee(
		depositAddress,
	)
	if err != nil {
		t.Fatal(err)
	}

	if initialDepositRedemptionFee.Cmp(actualDepositRedemptionFee) != 0 {
		t.Errorf(
			"unexpected redemption fee value\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			initialDepositRedemptionFee.Text(10),
			actualDepositRedemptionFee.Text(10),
		)
	}
}

func TestProvideRedemptionProof_KeepTerminatedEventOccurred(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionProof(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	initialDepositRedemptionFee, err := tbtcChain.DepositRedemptionFee(
		depositAddress,
	)
	if err != nil {
		t.Fatal(err)
	}

	keepSignature, err := submitKeepSignature(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.ProvideRedemptionSignature(
		depositAddress,
		keepSignature.V,
		keepSignature.R,
		keepSignature.S,
	)
	if err != nil {
		t.Fatal(err)
	}

	// wait a while before triggering the keep terminated event because the
	// extension must have time to handle the start event
	time.Sleep(100 * time.Millisecond)

	err = terminateKeep(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedIncreaseRedemptionFeeCalls := 0
	actualIncreaseRedemptionFeeCalls := tbtcChain.Logger().
		IncreaseRedemptionFeeCalls()
	if expectedIncreaseRedemptionFeeCalls != actualIncreaseRedemptionFeeCalls {
		t.Errorf(
			"unexpected number of IncreaseRedemptionFee calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedIncreaseRedemptionFeeCalls,
			actualIncreaseRedemptionFeeCalls,
		)
	}

	actualDepositRedemptionFee, err := tbtcChain.DepositRedemptionFee(
		depositAddress,
	)
	if err != nil {
		t.Fatal(err)
	}

	if initialDepositRedemptionFee.Cmp(actualDepositRedemptionFee) != 0 {
		t.Errorf(
			"unexpected redemption fee value\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			initialDepositRedemptionFee.Text(10),
			actualDepositRedemptionFee.Text(10),
		)
	}
}

func TestProvideRedemptionProof_ActionFailed(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionProof(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	keepSignature, err := submitKeepSignature(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.ProvideRedemptionSignature(
		depositAddress,
		keepSignature.V,
		keepSignature.R,
		keepSignature.S,
	)
	if err != nil {
		t.Fatal(err)
	}

	// simulate a situation when `IncreaseRedemptionFee` fails on-chain
	tbtcChain.SetAlwaysFailingTransactions("IncreaseRedemptionFee")

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedIncreaseRedemptionFeeCalls := 3
	actualIncreaseRedemptionFeeCalls := tbtcChain.Logger().
		IncreaseRedemptionFeeCalls()
	if expectedIncreaseRedemptionFeeCalls != actualIncreaseRedemptionFeeCalls {
		t.Errorf(
			"unexpected number of IncreaseRedemptionFee calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedIncreaseRedemptionFeeCalls,
			actualIncreaseRedemptionFeeCalls,
		)
	}
}

func TestProvideRedemptionProof_ContextCancelled_WithoutWorkingMonitoring(
	t *testing.T,
) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionProof(
		ctx,
		constantBackoff,
		timeout,
	)

	// cancel the context before any start event occurs
	cancelCtx()

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	keepSignature, err := submitKeepSignature(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.ProvideRedemptionSignature(
		depositAddress,
		keepSignature.V,
		keepSignature.R,
		keepSignature.S,
	)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedIncreaseRedemptionFeeCalls := 0
	actualIncreaseRedemptionFeeCalls := tbtcChain.Logger().
		IncreaseRedemptionFeeCalls()
	if expectedIncreaseRedemptionFeeCalls != actualIncreaseRedemptionFeeCalls {
		t.Errorf(
			"unexpected number of IncreaseRedemptionFee calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedIncreaseRedemptionFeeCalls,
			actualIncreaseRedemptionFeeCalls,
		)
	}
}

func TestProvideRedemptionProof_ContextCancelled_WithWorkingMonitoring(
	t *testing.T,
) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionProof(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	keepSignature, err := submitKeepSignature(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.ProvideRedemptionSignature(
		depositAddress,
		keepSignature.V,
		keepSignature.R,
		keepSignature.S,
	)
	if err != nil {
		t.Fatal(err)
	}

	// wait a while before cancelling the context because the
	// extension must have time to handle the start event
	time.Sleep(100 * time.Millisecond)

	// cancel the context once the start event is handled and
	// the monitoring process is running
	cancelCtx()

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedIncreaseRedemptionFeeCalls := 0
	actualIncreaseRedemptionFeeCalls := tbtcChain.Logger().
		IncreaseRedemptionFeeCalls()
	if expectedIncreaseRedemptionFeeCalls != actualIncreaseRedemptionFeeCalls {
		t.Errorf(
			"unexpected number of IncreaseRedemptionFee calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedIncreaseRedemptionFeeCalls,
			actualIncreaseRedemptionFeeCalls,
		)
	}
}

func TestProvideRedemptionProof_OperatorNotInSigningGroup(
	t *testing.T,
) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.monitorProvideRedemptionProof(
		ctx,
		constantBackoff,
		timeout,
	)

	signers := local.RandomSigningGroup(3)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.RedeemDeposit(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	keepSignature, err := submitKeepSignature(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	err = tbtcChain.ProvideRedemptionSignature(
		depositAddress,
		keepSignature.V,
		keepSignature.R,
		keepSignature.S,
	)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedIncreaseRedemptionFeeCalls := 0
	actualIncreaseRedemptionFeeCalls := tbtcChain.Logger().
		IncreaseRedemptionFeeCalls()
	if expectedIncreaseRedemptionFeeCalls != actualIncreaseRedemptionFeeCalls {
		t.Errorf(
			"unexpected number of IncreaseRedemptionFee calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedIncreaseRedemptionFeeCalls,
			actualIncreaseRedemptionFeeCalls,
		)
	}
}

func TestMonitorAndActDeduplication(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	monitoringName := "monitoring"

	shouldMonitorFn := func(depositAddress string) bool {
		return true
	}

	monitoringStartFn := func(
		handler depositEventHandler,
	) subscription.EventSubscription {
		for i := 0; i < 5; i++ {
			handler("deposit") // simulate multiple start events
		}

		return subscription.NewEventSubscription(func() {})
	}

	monitoringStopFn := func(
		handler depositEventHandler,
	) subscription.EventSubscription {
		return subscription.NewEventSubscription(func() {})
	}

	keepClosedFn := func(depositAddress string) (chan struct{}, func(), error) {
		return make(chan struct{}), func() {}, nil
	}

	var actCounter uint64
	actFn := func(depositAddress string) error {
		atomic.AddUint64(&actCounter, 1)
		return nil
	}

	timeoutFn := func(depositAddress string) (duration time.Duration, e error) {
		return timeout, nil
	}

	monitoringSubscription := tbtc.monitorAndAct(
		ctx,
		monitoringName,
		shouldMonitorFn,
		monitoringStartFn,
		monitoringStopFn,
		keepClosedFn,
		actFn,
		constantBackoff,
		timeoutFn,
	)
	defer monitoringSubscription.Unsubscribe()

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedActCounter := uint64(1)
	if actCounter != expectedActCounter {
		t.Errorf(
			"unexpected number of action invocations\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedActCounter,
			actCounter,
		)
	}
}

func TestAcquireMonitoringLock(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	if !tbtc.acquireMonitoringLock("0xAA", "monitoring one") {
		t.Errorf("monitoring wasn't started before; should be locked successfully")
	}

	if !tbtc.acquireMonitoringLock("0xBB", "monitoring one") {
		t.Errorf("monitoring wasn't started before; should be locked successfully")
	}

	if !tbtc.acquireMonitoringLock("0xAA", "monitoring two") {
		t.Errorf("monitoring wasn't started before; should be locked successfully")
	}

	if !tbtc.acquireMonitoringLock("0xBB", "monitoring two") {
		t.Errorf("monitoring wasn't started before; should be locked successfully")
	}
}

func TestAcquireMonitoringLock_Duplicate(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	if !tbtc.acquireMonitoringLock("0xAA", "monitoring one") {
		t.Errorf("monitoring wasn't started before; should be locked successfully")
	}

	if tbtc.acquireMonitoringLock("0xAA", "monitoring one") {
		t.Errorf("monitoring was started before; lock attempt should be rejected")
	}
}

func TestReleaseMonitoringLock(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	if !tbtc.acquireMonitoringLock("0xAA", "monitoring one") {
		t.Errorf("monitoring wasn't started before; should be locked successfully")
	}

	tbtc.releaseMonitoringLock("0xAA", "monitoring one")

	if !tbtc.acquireMonitoringLock("0xAA", "monitoring one") {
		t.Errorf("monitoring lock has been released; should be locked successfully")
	}
}

func TestReleaseMonitoringLock_WhenEmpty(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	tbtc.releaseMonitoringLock("0xAA", "monitoring one")

	if !tbtc.acquireMonitoringLock("0xAA", "monitoring one") {
		t.Errorf("monitoring wasn't started before; should be locked successfully")
	}
}

func TestShouldMonitorDeposit_ExpectedInitialState(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	// create a signing group which contains the operator
	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)
	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}
	tbtcChain.RetrieveSignerPubkey(depositAddress)

	// Deposit has just retrieved signer public key and is in
	// AwaitingSignerSetup state. shouldMonitorDeposit should return false for
	// any other state than AwaitingSignerSetup.
	shouldMonitor := tbtc.shouldMonitorDeposit(
		5*time.Second,
		depositAddress,
		chain.AwaitingBtcFundingProof,
	)
	if !shouldMonitor {
		t.Errorf("should monitor the deposit")
	}
}

func TestShouldMonitorDeposit_UnexpectedInitialState(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	// create a signing group which contains the operator
	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}
	tbtcChain.RetrieveSignerPubkey(depositAddress)

	// Deposit has just retrieved signer public key and is in
	// AwaitingSignerSetup state. shouldMonitorDeposit should return true
	// for this state.
	shouldMonitor := tbtc.shouldMonitorDeposit(
		5*time.Second,
		depositAddress,
		chain.AwaitingSignerSetup,
	)
	if shouldMonitor {
		t.Errorf("should not monitor the deposit")
	}
}

func TestShouldMonitorDeposit_InitialStateChange(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	// create a signing group which contains the operator
	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)
	_, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// Deposit has just been created and it is in AwaitingSignerSetup state.
	// This first check will setup *some* caching in shouldMonitorDeposit and
	// further checks will make sure this caching works as expected when the
	// state changed.
	shouldMonitor := tbtc.shouldMonitorDeposit(
		5*time.Second,
		depositAddress,
		chain.AwaitingSignerSetup,
	)
	if !shouldMonitor {
		t.Errorf("should monitor the deposit for AwaitingSignerSetup state")
	}

	tbtcChain.RetrieveSignerPubkey(depositAddress)

	// Deposit state has changed and it is now in AwaitingBtcFundingProof.
	// Those checks make sure the caching inside shouldMonitorDeposit does not
	// affect the result for changed state.
	shouldMonitor = tbtc.shouldMonitorDeposit(
		5*time.Second,
		depositAddress,
		chain.AwaitingBtcFundingProof,
	)
	if !shouldMonitor {
		t.Errorf("should monitor the deposit for AwaitingBtcFundingProof state")
	}
	shouldMonitor = tbtc.shouldMonitorDeposit(
		5*time.Second,
		depositAddress,
		chain.AwaitingSignerSetup,
	)
	if shouldMonitor {
		t.Errorf("should not monitor the deposit for AwaitingSignerSetup state")
	}
}

func TestShouldMonitorDeposit_MemberCache(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	// create a signing group which contains the operator
	signers := append(
		[]common.Address{tbtcChain.OperatorAddress()},
		local.RandomSigningGroup(2)...,
	)

	tbtcChain.CreateDeposit(depositAddress, signers)

	// all calls should be `true`
	const stateConfirmTimeout = 1 * time.Second
	call1 := tbtc.shouldMonitorDeposit(
		stateConfirmTimeout,
		depositAddress,
		chain.AwaitingSignerSetup,
	)
	call2 := tbtc.shouldMonitorDeposit(
		stateConfirmTimeout,
		depositAddress,
		chain.AwaitingSignerSetup,
	)
	call3 := tbtc.shouldMonitorDeposit(
		stateConfirmTimeout,
		depositAddress,
		chain.AwaitingSignerSetup,
	)

	if !(call1 && call2 && call3) {
		t.Errorf("should monitor deposit calls results are not same")
	}

	expectedChainCalls := 1
	actualChainCalls := tbtcChain.Logger().KeepAddressCalls()
	if expectedChainCalls != actualChainCalls {
		t.Errorf(
			"unexpected number of chain calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedChainCalls,
			actualChainCalls,
		)
	}
}

func TestShouldMonitorDeposit_NotMemberCache(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := local.NewTBTCLocalChain(ctx)
	tbtc := newTestTBTC(tbtcChain)

	// create a signing group which DOES NOT contain the operator
	signers := local.RandomSigningGroup(3)

	tbtcChain.CreateDeposit(depositAddress, signers)

	// all calls should be `false`
	const stateConfirmTimeout = 1 * time.Second
	call1 := tbtc.shouldMonitorDeposit(
		stateConfirmTimeout,
		depositAddress,
		chain.AwaitingSignerSetup,
	)
	call2 := tbtc.shouldMonitorDeposit(
		stateConfirmTimeout,
		depositAddress,
		chain.AwaitingSignerSetup,
	)
	call3 := tbtc.shouldMonitorDeposit(
		stateConfirmTimeout,
		depositAddress,
		chain.AwaitingSignerSetup,
	)

	if call1 || call2 || call3 {
		t.Errorf("should monitor deposit calls results are not same")
	}

	expectedChainCalls := 1
	actualChainCalls := tbtcChain.Logger().KeepAddressCalls()
	if expectedChainCalls != actualChainCalls {
		t.Errorf(
			"unexpected number of chain calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedChainCalls,
			actualChainCalls,
		)
	}
}

func TestGetSignerActionDelay(t *testing.T) {
	var tests = map[string]struct {
		signerIndex               int
		signersCount              int
		expectedSignerActionDelay time.Duration
	}{
		"signer index is `0`": {
			signerIndex:               0,
			signersCount:              3,
			expectedSignerActionDelay: 0 * time.Minute,
		},
		"signer index is `1`": {
			signerIndex:               1,
			signersCount:              3,
			expectedSignerActionDelay: 5 * time.Minute,
		},
		"signer index is `2`": {
			signerIndex:               2,
			signersCount:              3,
			expectedSignerActionDelay: 10 * time.Minute,
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			ctx, cancelCtx := context.WithCancel(context.Background())
			defer cancelCtx()

			tbtcChain := local.NewTBTCLocalChain(ctx)
			tbtc := newTestTBTC(tbtcChain)

			signers := local.RandomSigningGroup(test.signersCount)
			signers[test.signerIndex] = tbtcChain.OperatorAddress()

			tbtcChain.CreateDeposit(depositAddress, signers)

			actualSignerActionDelay, err := tbtc.getSignerActionDelay(
				depositAddress,
			)
			if err != nil {
				t.Fatal(err)
			}

			if test.expectedSignerActionDelay != actualSignerActionDelay {
				t.Errorf(
					"unexpected signer action delay\n"+
						"expected: [%v]\n"+
						"actual:   [%v]",
					test.expectedSignerActionDelay,
					actualSignerActionDelay,
				)
			}
		})
	}
}

func submitKeepPublicKey(
	depositAddress string,
	tbtcChain *local.TBTCLocalChain,
) ([64]byte, error) {
	keep, err := tbtcChain.Keep(depositAddress)
	if err != nil {
		return [64]byte{}, err
	}

	var keepPubkey [64]byte
	rand.Read(keepPubkey[:])

	err = keep.SubmitKeepPublicKey(keepPubkey)
	if err != nil {
		return [64]byte{}, err
	}

	return keepPubkey, nil
}

func submitKeepSignature(
	depositAddress string,
	tbtcChain *local.TBTCLocalChain,
) (*local.Signature, error) {
	keep, err := tbtcChain.Keep(depositAddress)
	if err != nil {
		return nil, err
	}

	signature := &ecdsa.Signature{
		R:          new(big.Int).SetUint64(rand.Uint64()),
		S:          new(big.Int).SetUint64(rand.Uint64()),
		RecoveryID: rand.Intn(4),
	}

	err = keep.SubmitSignature(signature)
	if err != nil {
		return nil, err
	}

	return toChainSignature(signature)
}

func toChainSignature(signature *ecdsa.Signature) (*local.Signature, error) {
	v := uint8(27 + signature.RecoveryID)

	r, err := byteutils.BytesTo32Byte(signature.R.Bytes())
	if err != nil {
		return nil, err
	}

	s, err := byteutils.BytesTo32Byte(signature.S.Bytes())
	if err != nil {
		return nil, err
	}

	return &local.Signature{
		V: v,
		R: r,
		S: s,
	}, nil
}

func areChainSignaturesEqual(signature1, signature2 *local.Signature) bool {
	if signature1.V != signature2.V {
		return false
	}

	if !bytes.Equal(signature1.R[:], signature2.R[:]) {
		return false
	}

	if !bytes.Equal(signature1.S[:], signature2.S[:]) {
		return false
	}

	return true
}

func closeKeep(
	depositAddress string,
	tbtcChain *local.TBTCLocalChain,
) error {
	keep, err := tbtcChain.Keep(depositAddress)
	if err != nil {
		return err
	}

	err = tbtcChain.CloseKeep(common.HexToAddress(keep.ID().String()))
	if err != nil {
		return err
	}

	return nil
}

func terminateKeep(
	depositAddress string,
	tbtcChain *local.TBTCLocalChain,
) error {
	keep, err := tbtcChain.Keep(depositAddress)
	if err != nil {
		return err
	}

	err = tbtcChain.TerminateKeep(common.HexToAddress(keep.ID().String()))
	if err != nil {
		return err
	}

	return nil
}

func constantBackoff(_ int) time.Duration {
	return time.Millisecond
}
