package local

import (
	"context"
	"math/big"
	"reflect"
	"testing"

	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

const (
	depositAddress = "0xa5FA806723A7c7c8523F33c39686f20b52612877"
)

func TestFundingInfo(t *testing.T) {
	expectedFundingInfo := &chain.FundingInfo{
		UtxoValueBytes:  [8]uint8{128, 150, 152, 0, 0, 0, 0, 0},
		FundedAt:        big.NewInt(1615172517),
		TransactionHash: "c27c3bfa8293ac6b303b9f7455ae23b7c24b8814915a6511976027064efc4d51",
		OutputIndex:     1,
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := NewTBTCLocalChain(ctx)

	tbtcChain.CreateDeposit(depositAddress, RandomSigningGroup(3))
	tbtcChain.FundDeposit(depositAddress)

	fundingInfo, err := tbtcChain.FundingInfo(depositAddress)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expectedFundingInfo, fundingInfo) {
		t.Errorf(
			"funding info does not match\nexpected: %+v\nactual:   %+v",
			expectedFundingInfo,
			fundingInfo,
		)
	}
}

func TestFundingInfo_NotFunded(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := NewTBTCLocalChain(ctx)

	tbtcChain.CreateDeposit(depositAddress, RandomSigningGroup(3))

	fundingInfo, err := tbtcChain.FundingInfo(depositAddress)
	if err != chain.ErrDepositNotFunded {
		t.Errorf(
			"unexpected error\nexpected: %v\nactual:   %v",
			chain.ErrDepositNotFunded,
			err,
		)
	}

	if fundingInfo != nil {
		t.Errorf(
			"funding info does not match\nexpected: %+v\nactual:   %+v",
			nil,
			fundingInfo,
		)
	}
}

func TestGetOwner(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	tbtcChain := NewTBTCLocalChain(ctx)

	signers := RandomSigningGroup(3)

	tbtcChain.CreateDeposit(depositAddress, signers)
	keep, err := tbtcChain.Keep(depositAddress)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	owner, err := keep.GetOwner()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if owner.String() != depositAddress {
		t.Errorf(
			"unexpected owner address\nexpected: %s\nactual:   %s",
			depositAddress,
			owner.String(),
		)
	}
}
