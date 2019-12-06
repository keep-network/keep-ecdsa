package tss

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-core/pkg/beacon/relay/group"
	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-tecdsa/internal/testdata"
	"github.com/keep-network/keep-tecdsa/pkg/net"
	"github.com/keep-network/keep-tecdsa/pkg/net/local"
)

func TestInitializeSignerAndGenerateKey(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	groupSize := 2
	threshold := groupSize

	err := log.SetLogLevel("*", "INFO")
	if err != nil {
		t.Errorf("logger initialization failed: [%v]", err)
	}

	completed := make(chan interface{})
	errChan := make(chan error)

	// go func() {
	groupMemberIDs := []group.MemberIndex{}
	groupMembersKeys := make(map[group.MemberIndex]*key.NetworkPublic, groupSize)

	for i := 0; i < groupSize; i++ {
		_, publicKey, err := key.GenerateStaticNetworkKey()
		if err != nil {
			t.Fatalf("failed to generate network key: [%v]", err)
		}

		memberIndex := group.MemberIndex(i + 1)

		groupMemberIDs = append(groupMemberIDs, memberIndex)
		groupMembersKeys[memberIndex] = publicKey
	}

	testData, err := testdata.LoadKeygenTestFixtures(groupSize)
	if err != nil {
		t.Fatalf("failed to load test data: [%v]", err)
	}

	signersMutex := &sync.Mutex{}
	signers := []*Signer{}

	// Signer initialization.
	for i, memberID := range groupMemberIDs {

		network, err := newTestNetProvider(memberID, groupMembersKeys, errChan)

		preParams := testData[i].LocalPreParams

		signer, err := InitializeSigner(
			memberID,
			groupSize,
			threshold,
			&preParams,
			network,
		)
		if err != nil {
			t.Fatalf("failed to initialize signer: [%v]", err)
		}

		signersMutex.Lock()
		signers = append(signers, signer)
		signersMutex.Unlock()
	}

	if len(signers) != len(groupMemberIDs) {
		t.Fatalf(
			"unexpected number of signers\nexpected: %d\nactual:   %d\n",
			len(groupMemberIDs),
			len(signers),
		)
	}

	// Key generaton.
	go func() {
		var keyGenWait sync.WaitGroup
		keyGenWait.Add(len(signers))

		for _, signer := range signers {
			go func(signer *Signer) {
				go func() {
					for {
						select {
						case err := <-signer.keygenErrChan:
							errChan <- err
							return
						}
					}
				}()

				err = signer.GenerateKey()
				if err != nil {
					errChan <- fmt.Errorf("failed to generate key: [%v]", err)
					return
				}

				keyGenWait.Done()
			}(signer)
		}

		keyGenWait.Wait()
		completed <- "DONE"
	}()

	select {
	case <-completed:
	case err := <-errChan:
		t.Fatalf("unexpected error on key generation: [%v]", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	firstPublicKey := signers[0].PublicKey()
	curve := secp256k1.S256()

	if !curve.IsOnCurve(firstPublicKey.X, firstPublicKey.Y) {
		t.Error("public key is not on curve")
	}

	for i, signer := range signers {
		publicKey := signer.PublicKey()
		if publicKey.X.Cmp(firstPublicKey.X) != 0 || publicKey.Y.Cmp(firstPublicKey.Y) != 0 {
			t.Errorf("public key for party [%v] doesn't match expected", i)
		}
	}

}

type testNetProvider struct {
}

func newTestNetProvider(
	memberID group.MemberIndex,
	membersNetworkKeys map[group.MemberIndex]*key.NetworkPublic,
	errChan chan error,
) (net.Provider, error) {
	provider := local.LocalProvider(
		memberID.Int().String(),
		membersNetworkKeys[memberID],
		errChan,
	)

	return provider, nil
}
