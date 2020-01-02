// Package node defines a node executing the TSS protocol.
package node

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-core/pkg/net/key"

	testdata "github.com/keep-network/keep-tecdsa/internal/testdata/signinggroup"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	ethLocal "github.com/keep-network/keep-tecdsa/pkg/chain/eth/local"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss"
	netLocal "github.com/keep-network/keep-tecdsa/pkg/net/local"
)

func TestGenerateSignerForKeep(t *testing.T) {
	groupSize := 5

	memberIDs, memberKeys, err := generateMemberKeys(groupSize)
	if err != nil {
		t.Fatal(err)
	}

	keepAddress := eth.KeepAddress([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1})

	localChain := ethLocal.Connect()
	localChain.(*ethLocal.LocalChain).CreateKeep(keepAddress)

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(groupSize)

	var signerPublicKey *ecdsa.PublicKey
	signers := make(map[string]*tss.ThresholdSigner)

	for i, memberID := range memberIDs {
		go func(i int, memberID common.Address) {
			defer waitGroup.Done()

			localChain := ethLocal.ConnectWithKey(memberKeys[memberID.String()])
			localNetwork := netLocal.LocalProvider(memberKeys[memberID.String()])

			node := NewNode(localChain, localNetwork)
			node.tssParamsPool = newTestPool(1)
			go node.tssParamsPool.pumpPool()

			signer, err := node.GenerateSignerForKeep(
				keepAddress,
				memberIDs,
			)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			signers[memberID.String()] = signer

		}(i, memberID)
	}

	waitGroup.Wait()

	signerPublicKey = signers[memberIDs[0].String()].PublicKey()

	keepPublicKey, err := localChain.(*ethLocal.LocalChain).GetKeepPublicKey(keepAddress)
	if err != nil {
		t.Fatalf("failed to get public key for keep: [%v]", err)
	}

	expectedPublicKey, err := eth.SerializePublicKey(signerPublicKey)
	if err != nil {
		t.Fatalf("failed to serialize public key: [%v]", err)
	}

	if keepPublicKey != expectedPublicKey {
		t.Errorf(
			"invalid keep public key\nexpected: %v\nactual:   %v\n",
			expectedPublicKey,
			keepPublicKey,
		)
	}
}

func TestCalculateSignature(t *testing.T) {
	digest := sha256.Sum256([]byte("message to sign"))
	groupSize := 5

	_, signer, err := testdata.LoadSigner(0)
	if err != nil {
		t.Fatal(err)

	}
	keepAddress := common.HexToAddress(signer.GroupID())

	localChain := ethLocal.Connect()
	localChain.(*ethLocal.LocalChain).CreateKeep(keepAddress)

	serializedPublicKey, err := eth.SerializePublicKey(signer.PublicKey())
	if err != nil {
		t.Fatal(err)
	}
	if err := localChain.(*ethLocal.LocalChain).SubmitKeepPublicKey(keepAddress, serializedPublicKey); err != nil {
		t.Fatal(err)
	}

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(groupSize)

	for i := 0; i < groupSize; i++ {
		go func(i int) {
			defer waitGroup.Done()

			networkKey, signer, err := testdata.LoadSigner(i)
			if err != nil {
				t.Errorf("failed to load signer: [%v]", err)
				return
			}

			localChain := ethLocal.ConnectWithKey(networkKey)
			localNetwork := netLocal.LocalProvider(networkKey)

			node := NewNode(localChain, localNetwork)
			node.tssParamsPool = newTestPool(1)
			go node.tssParamsPool.pumpPool()

			if err := node.CalculateSignature(
				signer,
				digest,
			); err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
		}(i)
	}

	waitGroup.Wait()

	signatures, err := localChain.(*ethLocal.LocalChain).GetSignatures(keepAddress)
	if err != nil {
		t.Fatalf("failed to get signatures for keep: [%v]", err)
	}

	if len(signatures) != 1 {
		t.Errorf(
			"invalid number of submitted signatures\nexpected: [%d]\nactual:   [%d]",
			1,
			len(signatures),
		)
	}

	for _, signature := range signatures {
		if !reflect.DeepEqual(signature, signatures[0]) {
			t.Errorf(
				"invalid submitted signature\nexpected: [%d]\nactual:   [%d]",
				signatures[0],
				signature,
			)
		}

	}
}

func generateMemberKeys(groupSize int) ([]common.Address, map[string]*key.NetworkPublic, error) {
	memberIDs := []common.Address{}
	groupMembersKeys := make(map[string]*key.NetworkPublic, groupSize)

	for i := 0; i < groupSize; i++ {
		_, publicKey, err := key.GenerateStaticNetworkKey()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate network key: [%v]", err)
		}

		memberID := common.HexToAddress(key.NetworkPubKeyToEthAddress(publicKey))

		memberIDs = append(memberIDs, memberID)
		groupMembersKeys[memberID.String()] = publicKey
	}

	return memberIDs, groupMembersKeys, nil
}
