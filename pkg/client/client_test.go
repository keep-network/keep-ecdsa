// Package client defines ECDSA keep client.
package client

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-tecdsa/internal/testdata"
	"github.com/keep-network/keep-tecdsa/internal/testutils/mock"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/local"
)

func TestRegistersAsCandidateMember(t *testing.T) {
	chain, persistence, err := initializeTest()
	if err != nil {
		t.Fatalf("failed to initialize test: [%v]", err)
	}

	Initialize(chain, persistence)

	expectedMemberCandidates := []common.Address{chain.Address()}
	memberCandidates := chain.GetMemberCandidates()

	if !reflect.DeepEqual(expectedMemberCandidates, memberCandidates) {
		t.Errorf(
			"unexpected registered member candidates\nexpected: [%x]\nactual:   [%x]",
			expectedMemberCandidates,
			memberCandidates,
		)
	}
}

func TestSigningForSignerLoadedFromStorage(t *testing.T) {
	keepAddress := testdata.KeepAddress1
	digest := [32]byte{1, 2, 3}

	chain, persistence, err := initializeTest()
	if err != nil {
		t.Fatalf("failed to initialize test: [%v]", err)
	}

	Initialize(chain, persistence)

	validateSignatureRequest(t, chain, keepAddress, digest, true)
}

func TestSigningForNewKeepWhereClientIsMember(t *testing.T) {
	keepAddress := (eth.KeepAddress)(common.HexToAddress("0xCFDB3668cFd46CbeA5a89Db36820e53f7B01A4d9"))
	digest := [32]byte{4, 5, 6}

	chain, persistence, err := initializeTest()
	if err != nil {
		t.Fatalf("failed to initialize test: [%v]", err)
	}

	Initialize(chain, persistence)

	keepMembers := []common.Address{chain.Address()}

	if err := chain.CreateKeep(keepAddress, keepMembers); err != nil {
		t.Fatalf(
			"failed to create keep [%s]: [%v]",
			keepAddress.String(), err,
		)
	}

	time.Sleep(100 * time.Millisecond)

	validateSignatureRequest(t, chain, keepAddress, digest, true)
}

func TestSigningForNewKeepWhereClientIsNotMember(t *testing.T) {
	keepAddress := (eth.KeepAddress)(common.HexToAddress("0xCFDB3668cFd46CbeA5a89Db36820e53f7B01A4d9"))
	digest := [32]byte{4, 5, 6}

	chain, persistence, err := initializeTest()
	if err != nil {
		t.Fatalf("failed to initialize test: [%v]", err)
	}

	Initialize(chain, persistence)

	keepMembers := []common.Address{}

	if err := chain.CreateKeep(keepAddress, keepMembers); err != nil {
		t.Fatalf(
			"failed to create keep [%s]: [%v]",
			keepAddress.String(), err,
		)
	}

	time.Sleep(100 * time.Millisecond)

	validateSignatureRequest(t, chain, keepAddress, digest, false)
}

func validateSignatureRequest(
	t *testing.T,
	chain *local.LocalChain,
	keepAddress eth.KeepAddress,
	digest [32]byte,
	expectedSignature bool,
) {
	signature, err := chain.GetSignature(keepAddress, digest)
	if err != nil {
		t.Fatalf("failed to get signature: [%v]", err)
	}
	if signature != nil {
		t.Error("signature has been already submitted")
	}

	err = chain.RequestSignature(keepAddress, digest)
	if err != nil {
		t.Fatalf("failed to request signature: [%v]", err)
	}

	time.Sleep(100 * time.Millisecond)

	signature, err = chain.GetSignature(keepAddress, digest)
	if err != nil {
		t.Fatalf("failed to get signature: [%v]", err)
	}

	if expectedSignature && signature == nil {
		t.Errorf(
			"expected signature presence: [%v]\nbut found signature: [%v]",
			expectedSignature,
			signature,
		)
	}
}

func initializeTest() (*local.LocalChain, *mock.PersistenceHandle, error) {
	persistence := &mock.PersistenceHandle{}

	chain := local.Connect().(*local.LocalChain)

	for keepAddress := range testdata.KeepSigners {
		keepMembers := []common.Address{chain.Address()}

		if err := chain.CreateKeep(keepAddress, keepMembers); err != nil {
			return nil, nil, fmt.Errorf(
				"failed to create keep [%s]: [%v]",
				keepAddress.String(), err,
			)
		}
	}

	return chain, persistence, nil
}
