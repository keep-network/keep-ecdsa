package firewall

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/cache"
	coreNet "github.com/keep-network/keep-core/pkg/net"
	"github.com/keep-network/keep-core/pkg/net/key"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/local"
)

var cacheLifeTime = time.Second

// Has minimum stake.
// Should allow to connect.
func TestHasMinimumStake(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	chain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(chain, coreFirewall)

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

	chain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(chain, coreFirewall)

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

	chain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(chain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	remotePeerAddress := common.HexToAddress(
		key.NetworkPubKeyToEthAddress(remotePeerPublicKey),
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

	chain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(chain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	remotePeerAddress := common.HexToAddress(
		key.NetworkPubKeyToEthAddress(remotePeerPublicKey),
	)

	chain.AuthorizeOperator(remotePeerAddress)

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

	chain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(chain, coreFirewall)

	_, remotePeer1PublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}
	remotePeer1Address := common.HexToAddress(
		key.NetworkPubKeyToEthAddress(remotePeer1PublicKey),
	)

	_, remotePeer2PublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}
	remotePeer2Address := common.HexToAddress(
		key.NetworkPubKeyToEthAddress(remotePeer2PublicKey),
	)

	policy.authorizedOperatorsCache.Add(remotePeer1Address.String())

	policy.nonAuthorizedOperatorsCache.Add(remotePeer2Address.String())
	chain.AuthorizeOperator(remotePeer2Address)

	err = policy.validateAuthorization(remotePeer1Address.String())
	if err != nil {
		t.Errorf("expected no valdation error; has: [%v]", err)
	}

	err = policy.validateAuthorization(remotePeer2Address.String())
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

	chain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(chain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	chain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToEthAddress(remotePeerPublicKey),
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

	chain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(chain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	chain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToEthAddress(remotePeerPublicKey),
	))

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

// Has no minimum stake.
// Has authorization.
// Is a member of an active keep
// Should allow to connect.
func TestNoMinimumStakeIsActiveKeepMember(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	chain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(chain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	chain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToEthAddress(remotePeerPublicKey),
	))

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

// Has no minimum stake.
// Has authorization.
// Is a member of a closed keep
// Should NOT allow to connect.
func TestNoMinimumStakeIsClosedKeepMember(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	chain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(chain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	chain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToEthAddress(remotePeerPublicKey),
	))

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

// Has no minimum stake.
// Has authorization.
// There are multiple keeps.
// Is a member of an active keep
// Should allow to connect.
func TestNoMinimumStakeMultipleKeepsMember(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	chain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(chain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	chain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToEthAddress(remotePeerPublicKey),
	))

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

// Has no minimum stake.
// Has authorization.
// There are multiple keeps.
// Is not a member of an active keep.
// Should NOT allow to connect but should cache all active keep members in-memory.
func TestCachesAllActiveKeepMembers(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	chain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(chain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	chain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToEthAddress(remotePeerPublicKey),
	))

	activeKeepMembers := []common.Address{
		common.HexToAddress("0xD6e148Be1E36Fc4Be9FE5a1abD7b3103ED527256"),
		common.HexToAddress("0x4f7C771Ab173bEc2BbE980497111866383a21172"),
	}
	closedKeepMembers := []common.Address{
		common.HexToAddress("0x1AD7E510d9AAA24588cB23De4F14fE57D42A5385"),
		common.HexToAddress("0x18e67aF1a54BF713Bc04EF811a7779b5AC0ef0eC"),
	}

	chain.OpenKeep(
		common.HexToAddress("0xCFEF2DC492E44a2747B2712f92d82527964B4b8F"),
		activeKeepMembers,
	)

	closedKeepAddress := common.HexToAddress("0x1Ca1EB1CafF6B3784Fe28a1b12266a10D04626A0")
	chain.OpenKeep(closedKeepAddress, closedKeepMembers)
	if err := chain.CloseKeep(closedKeepAddress); err != nil {
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
	if !policy.noActiveKeepMembersCache.Has(key.NetworkPubKeyToEthAddress(remotePeerPublicKey)) {
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

	chain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(chain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	chain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToEthAddress(remotePeerPublicKey),
	))

	keepAddress := common.HexToAddress("0x1Ca1EB1CafF6B3784Fe28a1b12266a10D04626A0")
	chain.OpenKeep(
		keepAddress,
		[]common.Address{
			common.HexToAddress(key.NetworkPubKeyToEthAddress(remotePeerPublicKey)),
		},
	)

	if err := policy.Validate(
		key.NetworkKeyToECDSAKey(remotePeerPublicKey),
	); err != nil {
		t.Fatalf("validation should pass: [%v]", err)
	}

	if err := chain.CloseKeep(keepAddress); err != nil {
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

	chain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(chain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	chain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToEthAddress(remotePeerPublicKey),
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
	chain.OpenKeep(
		keepAddress,
		[]common.Address{
			common.HexToAddress(key.NetworkPubKeyToEthAddress(remotePeerPublicKey)),
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

func TestIsKeepActiveCaching(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	chain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(chain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	chain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToEthAddress(remotePeerPublicKey),
	))

	keep1Address := common.HexToAddress("0xD6e148Be1E36Fc4Be9FE5a1abD7b3103ED527256")
	chain.OpenKeep(
		keep1Address,
		[]common.Address{
			common.HexToAddress("0x4f7C771Ab173bEc2BbE980497111866383a21172"),
		},
	)
	keep2Address := common.HexToAddress("0x1Ca1EB1CafF6B3784Fe28a1b12266a10D04626A0")
	chain.OpenKeep(
		keep2Address,
		[]common.Address{
			common.HexToAddress("0xF9798F39CfEf21931d3B5F73aF67718ae569a73e"),
		},
	)
	chain.CloseKeep(keep2Address)

	// first check, result should be put into the cache
	isActive, err := policy.isKeepActive(keep1Address)
	if err != nil {
		t.Fatal(err)
	}
	if !isActive {
		t.Fatal("keep is active")
	}
	isActive, err = policy.isKeepActive(keep2Address)
	if err != nil {
		t.Fatal(err)
	}
	if isActive {
		t.Fatal("keep is not active")
	}

	// result is read from the cache, should be the same as the original one
	isActive, err = policy.isKeepActive(keep1Address)
	if err != nil {
		t.Fatal(err)
	}
	if !isActive {
		t.Fatal("keep is active")
	}
	isActive, err = policy.isKeepActive(keep2Address)
	if err != nil {
		t.Fatal(err)
	}
	if isActive {
		t.Fatal("keep is not active")
	}

	// close active keep and see it's been updated properly
	chain.CloseKeep(keep1Address)
	isActive, err = policy.isKeepActive(keep1Address)
	if err != nil {
		t.Fatal(err)
	}
	if isActive {
		t.Fatal("keep is not active")
	}
}

func TestGetKeepMembersCaching(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	chain := local.Connect(ctx)
	coreFirewall := newMockCoreFirewall()
	policy := createNewPolicy(chain, coreFirewall)

	_, remotePeerPublicKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	chain.AuthorizeOperator(common.HexToAddress(
		key.NetworkPubKeyToEthAddress(remotePeerPublicKey),
	))

	keep1Address := common.HexToAddress("0xD6e148Be1E36Fc4Be9FE5a1abD7b3103ED527256")
	chain.OpenKeep(
		keep1Address,
		[]common.Address{
			common.HexToAddress("0x4f7C771Ab173bEc2BbE980497111866383a21172"),
			common.HexToAddress("0xA04Ba34b0689D1b1b5670a774a8EC5538C77FfaF"),
		},
	)
	keep2Address := common.HexToAddress("0x1Ca1EB1CafF6B3784Fe28a1b12266a10D04626A0")
	chain.OpenKeep(
		keep2Address,
		[]common.Address{
			common.HexToAddress("0xF9798F39CfEf21931d3B5F73aF67718ae569a73e"),
			common.HexToAddress("0x4f7C771Ab173bEc2BbE980497111866383a21172"),
		},
	)

	keep1ExpectedMembers := []string{
		"0x4f7C771Ab173bEc2BbE980497111866383a21172",
		"0xA04Ba34b0689D1b1b5670a774a8EC5538C77FfaF",
	}
	keep2ExpectedMembers := []string{
		"0xF9798F39CfEf21931d3B5F73aF67718ae569a73e",
		"0x4f7C771Ab173bEc2BbE980497111866383a21172",
	}

	// first check, result should be put into the cache
	members, err := policy.getKeepMembers(keep1Address)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(members, keep1ExpectedMembers) {
		t.Fatal("unexpected members")
	}
	members, err = policy.getKeepMembers(keep2Address)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(members, keep2ExpectedMembers) {
		t.Fatal("unexpected members")
	}

	// result is read from the cache, should be the same as the original one
	members, err = policy.getKeepMembers(keep1Address)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(members, keep1ExpectedMembers) {
		t.Fatal("unexpected members")
	}
	members, err = policy.getKeepMembers(keep2Address)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(members, keep2ExpectedMembers) {
		t.Fatal("unexpected members")
	}
}

func createNewPolicy(
	chain eth.Handle,
	coreFirewall coreNet.Firewall,
) *stakeOrActiveKeepPolicy {
	return &stakeOrActiveKeepPolicy{
		chain:                       chain,
		minimumStakePolicy:          coreFirewall,
		authorizedOperatorsCache:    cache.NewTimeCache(cacheLifeTime),
		nonAuthorizedOperatorsCache: cache.NewTimeCache(cacheLifeTime),
		activeKeepMembersCache:      cache.NewTimeCache(cacheLifeTime),
		noActiveKeepMembersCache:    cache.NewTimeCache(cacheLifeTime),
		keepInfoCache:               newKeepInfoCache(),
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
