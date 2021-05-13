package firewall

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/cache"
	coreNet "github.com/keep-network/keep-core/pkg/net"
	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/local"
)

var cacheLifeTime = time.Second

// Has minimum stake.
// Should allow to connect.
func TestHasMinimumStake(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	localChain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(localChain, coreFirewall)

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

// Has no minimum stake.
// Has no authorization.
// Should NOT allow to connect.
func TestNoAuthorization(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	localChain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(localChain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	if err := policy.Validate(
		key.NetworkKeyToECDSAKey(remotePeerPublicKey),
	); err != errNoAuthorization {
		t.Fatalf(
			"unexpected validation error\nactual:   [%v]\nexpected: [%v]",
			err,
			errNoMinStakeNoActiveKeep,
		)
	}
}

// Has no minimum stake
// Has no authorization
// Should cache the information operator is not authorized
func TestCachesNotAuthorizedOperators(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	localChain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(localChain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	remotePeerAddress := common.HexToAddress(
		key.NetworkPubKeyToChainAddress(remotePeerPublicKey),
	).String()

	policy.Validate(key.NetworkKeyToECDSAKey(remotePeerPublicKey))

	if policy.authorizedOperatorsCache.Has(remotePeerAddress) {
		t.Errorf("should not cache operator with no authorization")
	}
	if !policy.nonAuthorizedOperatorsCache.Has(remotePeerAddress) {
		t.Errorf("should cache operator with no authorization")
	}
}

// Has no minimum stake
// Has authorization
// Should cache the information operator is authorized.
func TestCachesAuthorizedOperators(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	localChain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(localChain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	remotePeerAddress := common.HexToAddress(
		key.NetworkPubKeyToChainAddress(remotePeerPublicKey),
	)

	localChain.AuthorizeOperator(remotePeerAddress)

	policy.Validate(key.NetworkKeyToECDSAKey(remotePeerPublicKey))

	if !policy.authorizedOperatorsCache.Has(remotePeerAddress.String()) {
		t.Errorf("should cache operator with no authorization")
	}
	if policy.nonAuthorizedOperatorsCache.Has(remotePeerAddress.String()) {
		t.Errorf("should not cache operator with no authorization")
	}
}

func TestConsultsAuthorizedOperatorsCache(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	localChain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(localChain, coreFirewall)

	_, remotePeer1PublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}
	remotePeer1Address := common.HexToAddress(
		key.NetworkPubKeyToChainAddress(remotePeer1PublicKey),
	)

	_, remotePeer2PublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}
	remotePeer2Address := common.HexToAddress(
		key.NetworkPubKeyToChainAddress(remotePeer2PublicKey),
	)

	policy.authorizedOperatorsCache.Add(remotePeer1Address.String())

	policy.nonAuthorizedOperatorsCache.Add(remotePeer2Address.String())
	localChain.AuthorizeOperator(remotePeer2Address)

	err = policy.validateAuthorization(localChain.PublicKeyToOperatorID(
		(*ecdsa.PublicKey)(remotePeer1PublicKey),
	))
	if err != nil {
		t.Errorf("expected no valdation error; has: [%v]", err)
	}

	err = policy.validateAuthorization(localChain.PublicKeyToOperatorID(
		(*ecdsa.PublicKey)(remotePeer2PublicKey),
	))
	if err != errNoAuthorization {
		t.Errorf("expected error about no authorization; has: [%v]", err)
	}
}

// Has no minimum stake.
// Has authorization.
// No keeps exist.
// Should NOT allow to connect.
func TestNoMinimumStakeNoKeepsExist(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	localChain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(localChain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	localChain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToChainAddress(remotePeerPublicKey),
	))

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

// Has no minimum stake.
// Has authorization.
// It not a member of a keep.
// Should NOT allow to connect.
func TestNoMinimumStakeIsNotKeepMember(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	localChain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(localChain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	localChain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToChainAddress(remotePeerPublicKey),
	))

	localChain.OpenKeep(
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

// Has no minimum stake.
// Has authorization.
// Is a member of an active keep
// Should allow to connect.
func TestNoMinimumStakeIsActiveKeepMember(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	localChain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(localChain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	localChain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToChainAddress(remotePeerPublicKey),
	))

	localChain.OpenKeep(
		common.HexToAddress("0xD6e148Be1E36Fc4Be9FE5a1abD7b3103ED527256"),
		[]common.Address{
			common.HexToAddress("0x4f7C771Ab173bEc2BbE980497111866383a21172"),
			common.HexToAddress(key.NetworkPubKeyToChainAddress(remotePeerPublicKey)),
		},
	)

	if err := policy.Validate(
		key.NetworkKeyToECDSAKey(remotePeerPublicKey),
	); err != nil {
		t.Fatalf("validation should pass: [%v]", err)
	}
}

// Has no minimum stake.
// Has authorization.
// Is a member of a closed keep
// Should NOT allow to connect.
func TestNoMinimumStakeIsClosedKeepMember(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	localChain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(localChain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	localChain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToChainAddress(remotePeerPublicKey),
	))

	keepAddress := common.HexToAddress("0xD6e148Be1E36Fc4Be9FE5a1abD7b3103ED527256")
	localChain.OpenKeep(
		keepAddress,
		[]common.Address{
			common.HexToAddress("0x4f7C771Ab173bEc2BbE980497111866383a21172"),
			common.HexToAddress(key.NetworkPubKeyToChainAddress(remotePeerPublicKey)),
		},
	)
	if err := localChain.CloseKeep(keepAddress); err != nil {
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

// Has no minimum stake.
// Has authorization.
// There are multiple keeps.
// Is a member of an active keep
// Should allow to connect.
func TestNoMinimumStakeMultipleKeepsMember(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	localChain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(localChain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	localChain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToChainAddress(remotePeerPublicKey),
	))

	keep1Address := common.HexToAddress("0xD6e148Be1E36Fc4Be9FE5a1abD7b3103ED527256")
	localChain.OpenKeep(
		keep1Address,
		[]common.Address{
			common.HexToAddress("0x4f7C771Ab173bEc2BbE980497111866383a21172"),
			common.HexToAddress(key.NetworkPubKeyToChainAddress(remotePeerPublicKey)),
		},
	)
	keep2Address := common.HexToAddress("0x1Ca1EB1CafF6B3784Fe28a1b12266a10D04626A0")
	localChain.OpenKeep(
		keep2Address,
		[]common.Address{
			common.HexToAddress("0xF9798F39CfEf21931d3B5F73aF67718ae569a73e"),
			common.HexToAddress(key.NetworkPubKeyToChainAddress(remotePeerPublicKey)),
		},
	)

	if err := localChain.CloseKeep(keep1Address); err != nil {
		t.Fatal(err)
	}

	if err := policy.Validate(
		key.NetworkKeyToECDSAKey(remotePeerPublicKey),
	); err != nil {
		t.Fatalf("validation should pass: [%v]", err)
	}
}

// Has no minimum stake.
// Has authorization.
// There are multiple keeps.
// Is not a member of an active keep.
// Should NOT allow to connect but should cache all active keep members in-memory.
func TestCachesAllActiveKeepMembers(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	localChain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(localChain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	localChain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToChainAddress(remotePeerPublicKey),
	))

	activeKeepMembers := []common.Address{
		common.HexToAddress("0xD6e148Be1E36Fc4Be9FE5a1abD7b3103ED527256"),
		common.HexToAddress("0x4f7C771Ab173bEc2BbE980497111866383a21172"),
	}
	closedKeepMembers := []common.Address{
		common.HexToAddress("0x1AD7E510d9AAA24588cB23De4F14fE57D42A5385"),
		common.HexToAddress("0x18e67aF1a54BF713Bc04EF811a7779b5AC0ef0eC"),
	}

	localChain.OpenKeep(
		common.HexToAddress("0xCFEF2DC492E44a2747B2712f92d82527964B4b8F"),
		activeKeepMembers,
	)

	closedKeepAddress := common.HexToAddress("0x1Ca1EB1CafF6B3784Fe28a1b12266a10D04626A0")
	localChain.OpenKeep(closedKeepAddress, closedKeepMembers)
	if err := localChain.CloseKeep(closedKeepAddress); err != nil {
		t.Fatal(err)
	}

	policy.Validate(key.NetworkKeyToECDSAKey(remotePeerPublicKey))

	if !policy.activeKeepMembersCache.Has(activeKeepMembers[0].String()) {
		t.Errorf("should cache active keep members")
	}
	if !policy.activeKeepMembersCache.Has(activeKeepMembers[1].String()) {
		t.Errorf("should cache active keep members")
	}
	if policy.activeKeepMembersCache.Has(closedKeepMembers[0].String()) {
		t.Errorf("should not cache non-active keep members")
	}
	if policy.activeKeepMembersCache.Has(closedKeepMembers[1].String()) {
		t.Errorf("should not cache non-active keep members")
	}

	// We don't put members of an active keep inside noActiveKeepMembersCache.
	// We don't put members of an inactive keep inside noActiveKeepMembersCache
	// because those members can belong to another active keep. We put there
	// only those members for which we executed the check.
	if policy.noActiveKeepMembersCache.Has(activeKeepMembers[0].String()) {
		t.Errorf("should not cache member until all keeps are checked")
	}
	if policy.noActiveKeepMembersCache.Has(activeKeepMembers[1].String()) {
		t.Errorf("should not cache member until all keeps are checked")
	}
	if policy.noActiveKeepMembersCache.Has(closedKeepMembers[0].String()) {
		t.Errorf("should not cache member until all keeps are checked")
	}
	if policy.noActiveKeepMembersCache.Has(closedKeepMembers[1].String()) {
		t.Errorf("should not cache member until all keeps are checked")
	}
	if !policy.noActiveKeepMembersCache.Has(key.NetworkPubKeyToChainAddress(remotePeerPublicKey)) {
		t.Errorf("should be in the no active keep members cache")
	}
}

// Has no minimum stake.
// Has authorization.
// There are multiple keeps.
// Is a member of an active keep
// Should allow to connect.
// After some time, the keep gets closed.
// It should no longer allow to connect.
func TestSweepsActiveKeepMembersCache(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	localChain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(localChain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	localChain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToChainAddress(remotePeerPublicKey),
	))

	keepAddress := common.HexToAddress("0x1Ca1EB1CafF6B3784Fe28a1b12266a10D04626A0")
	localChain.OpenKeep(
		keepAddress,
		[]common.Address{
			common.HexToAddress(key.NetworkPubKeyToChainAddress(remotePeerPublicKey)),
		},
	)

	if err := policy.Validate(
		key.NetworkKeyToECDSAKey(remotePeerPublicKey),
	); err != nil {
		t.Fatalf("validation should pass: [%v]", err)
	}

	if err := localChain.CloseKeep(keepAddress); err != nil {
		t.Fatal(err)
	}

	// still caching the old result
	if err := policy.Validate(
		key.NetworkKeyToECDSAKey(remotePeerPublicKey),
	); err != nil {
		t.Fatalf("validation should pass: [%v]", err)
	}

	time.Sleep(cacheLifeTime)

	// no longer caches the previous result
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

// Has no minimum stake.
// Has authorization.
// Is not a member of an active keep.
// Should NOT allow to connect.
// After some time, has a minimum stake again and becomes a member of an active keep.
// Shortly after that, the minimum stake drops below the required minimum but
// the membership in an active keep remains.
func TestSweepsNoActiveKeepMembersCache(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	localChain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(localChain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	localChain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToChainAddress(remotePeerPublicKey),
	))

	if err := policy.Validate(
		key.NetworkKeyToECDSAKey(remotePeerPublicKey),
	); err != errNoMinStakeNoActiveKeep {
		t.Fatalf(
			"unexpected validation error\nactual:   [%v]\nexpected: [%v]",
			err,
			errNoMinStakeNoActiveKeep,
		)
	}

	keepAddress := common.HexToAddress("0x1Ca1EB1CafF6B3784Fe28a1b12266a10D04626A0")
	localChain.OpenKeep(
		keepAddress,
		[]common.Address{
			common.HexToAddress(key.NetworkPubKeyToChainAddress(remotePeerPublicKey)),
		},
	)

	// still caching the old result
	if err := policy.Validate(
		key.NetworkKeyToECDSAKey(remotePeerPublicKey),
	); err != errNoMinStakeNoActiveKeep {
		t.Fatalf(
			"unexpected validation error\nactual:   [%v]\nexpected: [%v]",
			err,
			errNoMinStakeNoActiveKeep,
		)
	}

	time.Sleep(cacheLifeTime)

	// no longer caches the previous result
	if err := policy.Validate(
		key.NetworkKeyToECDSAKey(remotePeerPublicKey),
	); err != nil {
		t.Fatalf("validation should pass: [%v]", err)
	}
}

func TestGetKeepAtIndexCaching(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	localChain := withCallCounter(local.Connect(ctx))
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(localChain, coreFirewall)

	keep1Address := common.HexToAddress("0xD6e148Be1E36Fc4Be9FE5a1abD7b3103ED527256")
	keep1 := localChain.OpenKeep(
		keep1Address,
		[]common.Address{
			common.HexToAddress("0x4f7C771Ab173bEc2BbE980497111866383a21172"),
		},
	)
	keep2Address := common.HexToAddress("0x1Ca1EB1CafF6B3784Fe28a1b12266a10D04626A0")
	keep2 := localChain.OpenKeep(
		keep2Address,
		[]common.Address{
			common.HexToAddress("0xF9798F39CfEf21931d3B5F73aF67718ae569a73e"),
		},
	)

	// first check, result should be put into the cache
	keep, err := policy.getKeepAtIndex(big.NewInt(0))
	if err != nil {
		t.Fatal(err)
	}
	if keep.ID() != keep1.ID() {
		t.Fatal("unexpected keep at index 0")
	}
	keep, err = policy.getKeepAtIndex(big.NewInt(1))
	if err != nil {
		t.Fatal(err)
	}
	if keep.ID() != keep2.ID() {
		t.Fatal("unexpected keep at index 1")
	}

	// result is read from the cache, should be the same as the original one
	keep, err = policy.getKeepAtIndex(big.NewInt(0))
	if err != nil {
		t.Fatal(err)
	}
	if keep.ID() != keep1.ID() {
		t.Fatal("unexpected keep at index 0")
	}
	keep, err = policy.getKeepAtIndex(big.NewInt(1))
	if err != nil {
		t.Fatal(err)
	}
	if keep.ID() != keep2.ID() {
		t.Fatal("unexpected keep at index 1")
	}

	// we do cache result for each on-chain GetKeepAtIndex call so
	// there should be only two calls - one for each keep
	if localChain.getKeepAtIndexCallCount != 2 {
		t.Fatalf(
			"chain GetKeepAtIndex should be called 2 times; was [%v]",
			localChain.getKeepAtIndexCallCount,
		)
	}
}

func TestIsKeepActiveCaching(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	localChain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(localChain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	localChain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToChainAddress(remotePeerPublicKey),
	))

	keep1Address := common.HexToAddress("0xD6e148Be1E36Fc4Be9FE5a1abD7b3103ED527256")
	keep1 := withKeepCallCounter(localChain.OpenKeep(
		keep1Address,
		[]common.Address{
			common.HexToAddress("0x4f7C771Ab173bEc2BbE980497111866383a21172"),
		},
	))
	keep2Address := common.HexToAddress("0x1Ca1EB1CafF6B3784Fe28a1b12266a10D04626A0")
	keep2 := withKeepCallCounter(localChain.OpenKeep(
		keep2Address,
		[]common.Address{
			common.HexToAddress("0xF9798F39CfEf21931d3B5F73aF67718ae569a73e"),
		},
	))
	localChain.CloseKeep(keep2Address)

	// first check, result should be put into the cache
	isActive, err := policy.isKeepActive(keep1)
	if err != nil {
		t.Fatal(err)
	}
	if !isActive {
		t.Fatal("keep is active")
	}
	isActive, err = policy.isKeepActive(keep2)
	if err != nil {
		t.Fatal(err)
	}
	if isActive {
		t.Fatal("keep is not active")
	}

	// result is read from the cache, should be the same as the original one
	isActive, err = policy.isKeepActive(keep1)
	if err != nil {
		t.Fatal(err)
	}
	if !isActive {
		t.Fatal("keep is active")
	}
	isActive, err = policy.isKeepActive(keep2)
	if err != nil {
		t.Fatal(err)
	}
	if isActive {
		t.Fatal("keep is not active")
	}

	// close active keep and see it's been updated properly after caching period
	// elapsed
	localChain.CloseKeep(keep1Address)
	time.Sleep(cacheLifeTime)
	isActive, err = policy.isKeepActive(keep1)
	if err != nil {
		t.Fatal(err)
	}
	if isActive {
		t.Fatal("keep is not active")
	}

	// we do time-cache information that the keep is active and there should be
	// two on-chain isActive checks:
	// - first keep1 isActive check that yields true and is time-cached
	// - second keep1 isActive check after time cache elapsed that yields false
	if keep1.isActiveCallCount != 2 {
		t.Fatalf(
			"chain isActive should be called 2 time; was [%v]",
			keep1.isActiveCallCount,
		)
	}

	// we do cache information that the keep is inactive, so there should be
	// only one isActive check to the chain for a keep inactive during that
	// first check
	if keep2.isActiveCallCount != 1 {
		t.Fatalf(
			"chain isActive should be called only one time; was [%v]",
			keep2.isActiveCallCount,
		)
	}
}

func TestGetKeepMembersCaching(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	localChain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(localChain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	localChain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToChainAddress(remotePeerPublicKey),
	))

	keep1Address := common.HexToAddress("0xD6e148Be1E36Fc4Be9FE5a1abD7b3103ED527256")
	keep1 := withKeepCallCounter(localChain.OpenKeep(
		keep1Address,
		[]common.Address{
			common.HexToAddress("0x4f7C771Ab173bEc2BbE980497111866383a21172"),
			common.HexToAddress("0xA04Ba34b0689D1b1b5670a774a8EC5538C77FfaF"),
		},
	))
	keep2Address := common.HexToAddress("0x1Ca1EB1CafF6B3784Fe28a1b12266a10D04626A0")
	keep2 := withKeepCallCounter(localChain.OpenKeep(
		keep2Address,
		[]common.Address{
			common.HexToAddress("0xF9798F39CfEf21931d3B5F73aF67718ae569a73e"),
			common.HexToAddress("0x4f7C771Ab173bEc2BbE980497111866383a21172"),
		},
	))

	keep1ExpectedMembers := []string{
		"0x4f7C771Ab173bEc2BbE980497111866383a21172",
		"0xA04Ba34b0689D1b1b5670a774a8EC5538C77FfaF",
	}
	keep2ExpectedMembers := []string{
		"0xF9798F39CfEf21931d3B5F73aF67718ae569a73e",
		"0x4f7C771Ab173bEc2BbE980497111866383a21172",
	}

	// first check, result should be put into the cache
	members, err := policy.getKeepMembers(keep1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(members, keep1ExpectedMembers) {
		t.Fatal("unexpected members")
	}
	members, err = policy.getKeepMembers(keep2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(members, keep2ExpectedMembers) {
		t.Fatal("unexpected members")
	}

	// result is read from the cache, should be the same as the original one
	members, err = policy.getKeepMembers(keep1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(members, keep1ExpectedMembers) {
		t.Fatal("unexpected members")
	}
	members, err = policy.getKeepMembers(keep2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(members, keep2ExpectedMembers) {
		t.Fatal("unexpected members")
	}

	// we do cache result for each on-chain getMembers call so there should be
	// only one on-chain getMembers call for each keep
	if keep1.getMembersCallCount != 1 {
		t.Fatalf(
			"chain getMembers should be called only one time; was [%v]",
			keep1.getMembersCallCount,
		)
	}
	if keep2.getMembersCallCount != 1 {
		t.Fatalf(
			"chain getMembers should be called only one time; was [%v]",
			keep2.getMembersCallCount,
		)
	}
}

func createNewPolicy(
	chainHandle chain.Handle,
	coreFirewall coreNet.Firewall,
) *stakeOrActiveKeepPolicy {
	return &stakeOrActiveKeepPolicy{
		chain:                       chainHandle,
		minimumStakePolicy:          coreFirewall,
		authorizedOperatorsCache:    cache.NewTimeCache(cacheLifeTime),
		nonAuthorizedOperatorsCache: cache.NewTimeCache(cacheLifeTime),
		activeKeepMembersCache:      cache.NewTimeCache(cacheLifeTime),
		noActiveKeepMembersCache:    cache.NewTimeCache(cacheLifeTime),
		keepInfoCache:               newKeepInfoCache(cacheLifeTime),
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

func withCallCounter(handle local.Chain) *chainWithCallCounter {
	return &chainWithCallCounter{
		Chain: handle,
		getKeepAtIndexCallCount: 0,
	}
}

type chainWithCallCounter struct {
	local.Chain
	getKeepAtIndexCallCount uint64
}

func (cwcc *chainWithCallCounter) GetKeepAtIndex(
	keepIndex *big.Int,
) (chain.BondedECDSAKeepHandle, error) {
	atomic.AddUint64(&cwcc.getKeepAtIndexCallCount, 1)
	return cwcc.Chain.GetKeepAtIndex(keepIndex)
}

func withKeepCallCounter(handle chain.BondedECDSAKeepHandle) *keepHandleWithCounter {
	return &keepHandleWithCounter{
		BondedECDSAKeepHandle: handle,
		isActiveCallCount:     0,
		getMembersCallCount:   0,
	}
}

type keepHandleWithCounter struct {
	chain.BondedECDSAKeepHandle

	isActiveCallCount   uint64
	getMembersCallCount uint64
}

func (khwc *keepHandleWithCounter) IsActive() (bool, error) {
	atomic.AddUint64(&khwc.isActiveCallCount, 1)
	return khwc.BondedECDSAKeepHandle.IsActive()
}

func (khwc *keepHandleWithCounter) GetMembers() ([]chain.ID, error) {
	atomic.AddUint64(&khwc.getMembersCallCount, 1)
	return khwc.BondedECDSAKeepHandle.GetMembers()
}
