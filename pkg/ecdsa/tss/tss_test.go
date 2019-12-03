package tss

import (
	"context"
	cecdsa "crypto/ecdsa"
	"crypto/sha256"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-core/pkg/beacon/relay/group"
	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-tecdsa/internal/testdata"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-tecdsa/pkg/net"
	"github.com/keep-network/keep-tecdsa/pkg/net/local"
	"github.com/keep-network/keep-tecdsa/pkg/utils/testutils"
)

func TestGenerateKeyAndSign(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	groupSize := 5
	threshold := groupSize - 1

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

	networkProviders := []net.Provider{}

	// Members initialization.
	members := []*Member{}

	for i, memberID := range groupMemberIDs {
		network, err := newTestNetProvider(memberID, groupMembersKeys, errChan)

		networkProviders = append(networkProviders, network)

		preParams := testData[i].LocalPreParams

		member, err := InitializeKeyGeneration(
			memberID,
			groupSize,
			threshold,
			&preParams,
			network,
		)
		if err != nil {
			t.Fatalf("failed to initialize signer: [%v]", err)
		}

		members = append(members, member)
	}

	if len(members) != len(groupMemberIDs) {
		t.Fatalf(
			"unexpected number of signers\nexpected: %d\nactual:   %d\n",
			len(groupMemberIDs),
			len(members),
		)
	}

	// Key generaton.
	signersMutex := sync.Mutex{}
	signers := []*Signer{}

	go func() {
		var keyGenWait sync.WaitGroup
		keyGenWait.Add(len(members))

		for _, member := range members {
			go func(member *Member) {
				go func() {
					for {
						select {
						case err := <-member.keygenErrChan:
							errChan <- err
							return
						}
					}
				}()

				signer, err := member.GenerateKey()
				if err != nil {
					errChan <- fmt.Errorf("failed to generate key: [%v]", err)
					return
				}

				signersMutex.Lock()
				signers = append(signers, signer)
				signersMutex.Unlock()

				keyGenWait.Done()
			}(member)
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

	firstSigner := signers[0]
	firstPublicKey := firstSigner.PublicKey()
	curve := secp256k1.S256()

	if !curve.IsOnCurve(firstPublicKey.X, firstPublicKey.Y) {
		t.Error("public key is not on curve")
	}

	for i, signer := range signers {
		publicKey := signer.PublicKey()
		if publicKey.X.Cmp(firstPublicKey.X) != 0 || publicKey.Y.Cmp(firstPublicKey.Y) != 0 {
			t.Errorf("public key for signer [%d] doesn't match expected", i)
		}
	}

	// Signing initialization.
	message := []byte("message to sign")
	digest := sha256.Sum256(message)

	signingSigners := []*SigningSigner{}

	var initSigningWait sync.WaitGroup
	initSigningWait.Add(len(groupMembersKeys))

	for i, signer := range signers {
		signingSigner := signer.InitializeSigning(
			digest[:],
			networkProviders[i],
		)

		signingSigners = append(signingSigners, signingSigner)
	}

	// Signing.
	signatures := []*ecdsa.Signature{}
	signaturesMutex := &sync.Mutex{}

	var signingWait sync.WaitGroup
	signingWait.Add(len(groupMembersKeys))

	for _, signingSigner := range signingSigners {
		go func() {
			signature, err := signingSigner.Sign()
			if err != nil {
				t.Errorf("failed to sign: [%v]", err)
			}

			signaturesMutex.Lock()
			signatures = append(signatures, signature)
			signaturesMutex.Unlock()

			signingWait.Done()
		}()
	}

	signingWait.Wait()

	if len(signatures) != groupSize {
		t.Errorf("invalid number of signatures\nexpected: %d\nactual:   %d", groupSize, len(signatures))
	}

	firstSignature := signatures[0]
	for i, signature := range signatures {
		if !reflect.DeepEqual(firstSignature, signature) {
			t.Errorf(
				"signature for party [%v] doesn't match expected\nexpected: [%v]\nactual: [%v]",
				i,
				firstSignature,
				signature,
			)
		}
	}

	if !cecdsa.Verify(
		(*cecdsa.PublicKey)(firstPublicKey),
		digest[:],
		firstSignature.R,
		firstSignature.S,
	) {
		t.Errorf("invalid signature: [%+v]", firstSignature)
	}

	testutils.VerifyEthereumSignature(t, digest[:], firstSignature, firstPublicKey)
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
