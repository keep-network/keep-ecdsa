package firewall

import (
	"crypto/ecdsa"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-ecdsa/pkg/chain/local"
)

var cacheLifeTime = time.Second

func TestHasMinimumStake(t *testing.T) {
	chain := local.Connect()
	coreFirewall := newMockCoreFirewall()
	policy := &stakeOrActiveKeepPolicy{
		chain:                  chain,
		minimumStakePolicy:     coreFirewall,
		activeKeepMembersCache: newTimeCache(cacheLifeTime),
	}

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	coreFirewall.updatePeer(remotePeerPublicKey, true)

	if err := policy.Validate(
		key.NetworkKeyToECDSAKey(remotePeerPublicKey),
	); err != nil {
		t.Fatalf("validation should pass: [%v]", err)
	}
}

func TestNoMinimumStakeNoKeepsExist(t *testing.T) {
	chain := local.Connect()
	coreFirewall := newMockCoreFirewall()
	policy := &stakeOrActiveKeepPolicy{
		chain:                  chain,
		minimumStakePolicy:     coreFirewall,
		activeKeepMembersCache: newTimeCache(cacheLifeTime),
	}

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	if err := policy.Validate(
		key.NetworkKeyToECDSAKey(remotePeerPublicKey),
	); err != errNoMinStakeNoActiveKeep {
		t.Fatalf(
			"unexpected validation error\nactual:   [%v]\nexpected: [%v]",
			err,
			errNoMinStakeNoActiveKeep,
		)
	}
}

func TestNoMinimumStakeIsNotKeepMember(t *testing.T) {
	chain := local.Connect()
	coreFirewall := newMockCoreFirewall()
	policy := &stakeOrActiveKeepPolicy{
		chain:                  chain,
		minimumStakePolicy:     coreFirewall,
		activeKeepMembersCache: newTimeCache(cacheLifeTime),
	}

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	chain.OpenKeep(
		common.HexToAddress("0xD6e148Be1E36Fc4Be9FE5a1abD7b3103ED527256"),
		[]common.Address{
			common.HexToAddress("0x4f7C771Ab173bEc2BbE980497111866383a21172"),
			common.HexToAddress("0x65ea55c1f10491038425725dc00dffeab2a1e28a"),
		},
	)

	if err := policy.Validate(
		key.NetworkKeyToECDSAKey(remotePeerPublicKey),
	); err != errNoMinStakeNoActiveKeep {
		t.Fatalf(
			"unexpected validation error\nactual:   [%v]\nexpected: [%v]",
			err,
			errNoMinStakeNoActiveKeep,
		)
	}
}

func TestNoMinimumStakeIsActiveKeepMember(t *testing.T) {
	chain := local.Connect()
	coreFirewall := newMockCoreFirewall()
	policy := &stakeOrActiveKeepPolicy{
		chain:                  chain,
		minimumStakePolicy:     coreFirewall,
		activeKeepMembersCache: newTimeCache(cacheLifeTime),
	}

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	chain.OpenKeep(
		common.HexToAddress("0xD6e148Be1E36Fc4Be9FE5a1abD7b3103ED527256"),
		[]common.Address{
			common.HexToAddress("0x4f7C771Ab173bEc2BbE980497111866383a21172"),
			common.HexToAddress(key.NetworkPubKeyToEthAddress(remotePeerPublicKey)),
		},
	)

	if err := policy.Validate(
		key.NetworkKeyToECDSAKey(remotePeerPublicKey),
	); err != nil {
		t.Fatalf("validation should pass: [%v]", err)
	}
}

func TestNoMinimumStakeIsClosedKeepMember(t *testing.T) {
	chain := local.Connect()
	coreFirewall := newMockCoreFirewall()
	policy := &stakeOrActiveKeepPolicy{
		chain:                  chain,
		minimumStakePolicy:     coreFirewall,
		activeKeepMembersCache: newTimeCache(cacheLifeTime),
	}

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	keepAddress := common.HexToAddress("0xD6e148Be1E36Fc4Be9FE5a1abD7b3103ED527256")
	chain.OpenKeep(
		keepAddress,
		[]common.Address{
			common.HexToAddress("0x4f7C771Ab173bEc2BbE980497111866383a21172"),
			common.HexToAddress(key.NetworkPubKeyToEthAddress(remotePeerPublicKey)),
		},
	)
	if err := chain.CloseKeep(keepAddress); err != nil {
		t.Fatal(err)
	}

	if err := policy.Validate(
		key.NetworkKeyToECDSAKey(remotePeerPublicKey),
	); err != errNoMinStakeNoActiveKeep {
		t.Fatalf(
			"unexpected validation error\nactual:   [%v]\nexpected: [%v]",
			err,
			errNoMinStakeNoActiveKeep,
		)
	}
}

func TestNoMinimumStakeMultipleKeepsMember(t *testing.T) {
	chain := local.Connect()
	coreFirewall := newMockCoreFirewall()
	policy := &stakeOrActiveKeepPolicy{
		chain:                  chain,
		minimumStakePolicy:     coreFirewall,
		activeKeepMembersCache: newTimeCache(cacheLifeTime),
	}

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	keep1Address := common.HexToAddress("0xD6e148Be1E36Fc4Be9FE5a1abD7b3103ED527256")
	chain.OpenKeep(
		keep1Address,
		[]common.Address{
			common.HexToAddress("0x4f7C771Ab173bEc2BbE980497111866383a21172"),
			common.HexToAddress(key.NetworkPubKeyToEthAddress(remotePeerPublicKey)),
		},
	)
	keep2Address := common.HexToAddress("0x1Ca1EB1CafF6B3784Fe28a1b12266a10D04626A0")
	chain.OpenKeep(
		keep2Address,
		[]common.Address{
			common.HexToAddress("0xF9798F39CfEf21931d3B5F73aF67718ae569a73e"),
			common.HexToAddress(key.NetworkPubKeyToEthAddress(remotePeerPublicKey)),
		},
	)

	if err := chain.CloseKeep(keep1Address); err != nil {
		t.Fatal(err)
	}

	if err := policy.Validate(
		key.NetworkKeyToECDSAKey(remotePeerPublicKey),
	); err != nil {
		t.Fatalf("validation should pass: [%v]", err)
	}
}

func newMockCoreFirewall() *mockCoreFirewall {
	return &mockCoreFirewall{
		meetsCriteria: make(map[uint64]bool),
	}
}

type mockCoreFirewall struct {
	meetsCriteria map[uint64]bool
}

func (mf *mockCoreFirewall) Validate(remotePeerPublicKey *ecdsa.PublicKey) error {
	if !mf.meetsCriteria[remotePeerPublicKey.X.Uint64()] {
		return fmt.Errorf("remote peer does not meet firewall criteria")
	}
	return nil
}

func (mf *mockCoreFirewall) updatePeer(
	remotePeerPublicKey *key.NetworkPublic,
	meetsCriteria bool,
) {
	x := key.NetworkKeyToECDSAKey(remotePeerPublicKey).X.Uint64()
	mf.meetsCriteria[x] = meetsCriteria
}
