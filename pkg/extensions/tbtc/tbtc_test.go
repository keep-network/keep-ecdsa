package tbtc

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-ecdsa/pkg/utils/byteutils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-ecdsa/pkg/chain/local"
)

const (
	timeout        = 500 * time.Millisecond
	depositAddress = "0xa5FA806723A7c7c8523F33c39686f20b52612877"
)

func TestRetrievePubkey_TimeoutElapsed(t *testing.T) {
	ctx := context.Background()
	tbtcChain := local.NewTBTCLocalChain()
	tbtc := newTBTC(tbtcChain)

	err := tbtc.monitorRetrievePubKey(
		ctx,
		constantBackoff,
		timeout,
	)
	if err != nil {
		t.Fatal(err)
	}

	tbtcChain.CreateDeposit(depositAddress)

	keepPubkey, err := submitKeepPublicKey(depositAddress, tbtcChain)
	if err != nil {
		t.Fatal(err)
	}

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedRetrieveSignerPubkeyCalls := 1
	actualRetrieveSignerPubkeyCalls := tbtcChain.Logger().RetrieveSignerPubkeyCalls()
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
		t.Errorf("unexpected error while fetching deposit pubkey: [%v]", err)
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
	ctx := context.Background()
	tbtcChain := local.NewTBTCLocalChain()
	tbtc := newTBTC(tbtcChain)

	err := tbtc.monitorRetrievePubKey(
		ctx,
		constantBackoff,
		timeout,
	)
	if err != nil {
		t.Fatal(err)
	}

	tbtcChain.CreateDeposit(depositAddress)

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
	actualRetrieveSignerPubkeyCalls := tbtcChain.Logger().RetrieveSignerPubkeyCalls()
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
		t.Errorf("unexpected error while fetching deposit pubkey: [%v]", err)
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
	ctx := context.Background()
	tbtcChain := local.NewTBTCLocalChain()
	tbtc := newTBTC(tbtcChain)

	err := tbtc.monitorRetrievePubKey(
		ctx,
		constantBackoff,
		timeout,
	)
	if err != nil {
		t.Fatal(err)
	}

	tbtcChain.CreateDeposit(depositAddress)

	_, err = submitKeepPublicKey(depositAddress, tbtcChain)
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
	actualRetrieveSignerPubkeyCalls := tbtcChain.Logger().RetrieveSignerPubkeyCalls()
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

	expectedError := fmt.Errorf("no pubkey for deposit [%v]", depositAddress)
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
	ctx := context.Background()
	tbtcChain := local.NewTBTCLocalChain()
	tbtc := newTBTC(tbtcChain)

	err := tbtc.monitorRetrievePubKey(
		ctx,
		constantBackoff,
		timeout,
	)
	if err != nil {
		t.Fatal(err)
	}

	tbtcChain.CreateDeposit(depositAddress)

	_, err = submitKeepPublicKey(depositAddress, tbtcChain)
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
	actualRetrieveSignerPubkeyCalls := tbtcChain.Logger().RetrieveSignerPubkeyCalls()
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

	expectedError := fmt.Errorf("no pubkey for deposit [%v]", depositAddress)
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
	ctx := context.Background()
	tbtcChain := local.NewTBTCLocalChain()
	tbtc := newTBTC(tbtcChain)

	err := tbtc.monitorRetrievePubKey(
		ctx,
		constantBackoff,
		timeout,
	)
	if err != nil {
		t.Fatal(err)
	}

	tbtcChain.CreateDeposit(depositAddress)

	// do not submit the keep public key intentionally to cause
	// the action error

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedRetrieveSignerPubkeyCalls := 3
	actualRetrieveSignerPubkeyCalls := tbtcChain.Logger().RetrieveSignerPubkeyCalls()
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
	tbtcChain := local.NewTBTCLocalChain()
	tbtc := newTBTC(tbtcChain)

	err := tbtc.monitorRetrievePubKey(
		ctx,
		constantBackoff,
		timeout,
	)
	if err != nil {
		t.Fatal(err)
	}

	// cancel the context before any start event occurs
	cancelCtx()

	tbtcChain.CreateDeposit(depositAddress)

	// wait a bit longer than the monitoring timeout
	// to make sure the potential transaction completes
	time.Sleep(2 * timeout)

	expectedRetrieveSignerPubkeyCalls := 0
	actualRetrieveSignerPubkeyCalls := tbtcChain.Logger().RetrieveSignerPubkeyCalls()
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
	tbtcChain := local.NewTBTCLocalChain()
	tbtc := newTBTC(tbtcChain)

	err := tbtc.monitorRetrievePubKey(
		ctx,
		constantBackoff,
		timeout,
	)
	if err != nil {
		t.Fatal(err)
	}

	tbtcChain.CreateDeposit(depositAddress)

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
	actualRetrieveSignerPubkeyCalls := tbtcChain.Logger().RetrieveSignerPubkeyCalls()
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

func TestProvideRedemptionSignature_TimeoutElapsed(t *testing.T) {
	ctx := context.Background()
	tbtcChain := local.NewTBTCLocalChain()
	tbtc := newTBTC(tbtcChain)

	err := tbtc.monitorProvideRedemptionSignature(
		ctx,
		constantBackoff,
		timeout,
	)
	if err != nil {
		t.Fatal(err)
	}

	tbtcChain.CreateDeposit(depositAddress)

	_, err = submitKeepPublicKey(depositAddress, tbtcChain)
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
	actualProvideRedemptionSignatureCalls := tbtcChain.Logger().ProvideRedemptionSignatureCalls()
	if expectedProvideRedemptionSignatureCalls != actualProvideRedemptionSignatureCalls {
		t.Errorf(
			"unexpected number of ProvideRedemptionSignature calls\n"+
				"expected: [%v]\n"+
				"actual:   [%v]",
			expectedProvideRedemptionSignatureCalls,
			actualProvideRedemptionSignatureCalls,
		)
	}

	depositSignature, err := tbtcChain.DepositSignature(depositAddress)
	if err != nil {
		t.Errorf("unexpected error while fetching deposit pubkey: [%v]", err)
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

func submitKeepPublicKey(
	depositAddress string,
	tbtcChain *local.TBTCLocalChain,
) ([64]byte, error) {
	keepAddress, err := tbtcChain.KeepAddress(depositAddress)
	if err != nil {
		return [64]byte{}, err
	}

	var keepPubkey [64]byte
	rand.Read(keepPubkey[:])

	err = tbtcChain.SubmitKeepPublicKey(
		common.HexToAddress(keepAddress),
		keepPubkey,
	)
	if err != nil {
		return [64]byte{}, err
	}

	return keepPubkey, nil
}

func submitKeepSignature(
	depositAddress string,
	tbtcChain *local.TBTCLocalChain,
) (*local.Signature, error) {
	keepAddress, err := tbtcChain.KeepAddress(depositAddress)
	if err != nil {
		return nil, err
	}

	signature := &ecdsa.Signature{
		R:          new(big.Int).SetUint64(rand.Uint64()),
		S:          new(big.Int).SetUint64(rand.Uint64()),
		RecoveryID: rand.Intn(4),
	}

	err = tbtcChain.SubmitSignature(
		common.HexToAddress(keepAddress),
		signature,
	)
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
	keepAddress, err := tbtcChain.KeepAddress(depositAddress)
	if err != nil {
		return err
	}

	err = tbtcChain.CloseKeep(common.HexToAddress(keepAddress))
	if err != nil {
		return err
	}

	return nil
}

func terminateKeep(
	depositAddress string,
	tbtcChain *local.TBTCLocalChain,
) error {
	keepAddress, err := tbtcChain.KeepAddress(depositAddress)
	if err != nil {
		return err
	}

	err = tbtcChain.TerminateKeep(common.HexToAddress(keepAddress))
	if err != nil {
		return err
	}

	return nil
}

func constantBackoff(_ int) time.Duration {
	return time.Millisecond
}
