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

	err := log.SetLogLevel("*", "ERROR")
	if err != nil {
		t.Fatalf("logger initialization failed: [%v]", err)
	}

	errChan := make(chan error)

	groupMemberIDs := []group.MemberIndex{}
	groupMembersKeys := make(map[group.MemberIndex]*key.NetworkPublic, groupSize)
	networkProviders := make(map[group.MemberIndex]net.Provider, groupSize)

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

	// Members initialization.
	members := make(map[group.MemberIndex]*Member)

	for i, memberID := range groupMemberIDs {
		network, err := newTestNetProvider(memberID, groupMembersKeys, errChan)

		networkProviders[memberID] = network

		preParams := testData[i].LocalPreParams

		member, err := InitializeKeyGeneration(
			memberID,
			groupSize,
			threshold,
			&preParams,
			network,
		)
		if err != nil {
			t.Fatalf("failed to initialize member: [%v]", err)
		}

		members[memberID] = member
	}

	if len(members) != len(groupMemberIDs) {
		t.Fatalf(
			"unexpected number of signers\nexpected: %d\nactual:   %d\n",
			len(groupMemberIDs),
			len(members),
		)
	}

	// Key generation.
	signersMutex := sync.Mutex{}
	signers := make(map[group.MemberIndex]*Signer)

	keyGenDone := make(chan interface{})

	go func() {
		var keyGenWait sync.WaitGroup
		keyGenWait.Add(groupSize)

		for memberID, member := range members {
			go func(memberID group.MemberIndex, member *Member) {
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
				signers[memberID] = signer
				signersMutex.Unlock()

				keyGenWait.Done()
			}(memberID, member)
		}

		keyGenWait.Wait()
		close(keyGenDone)
	}()

	select {
	case <-keyGenDone:
	case err := <-errChan:
		t.Fatalf("unexpected error on key generation: [%v]", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	firstSigner := signers[groupMemberIDs[0]]
	firstPublicKey := firstSigner.PublicKey()
	curve := secp256k1.S256()

	if !curve.IsOnCurve(firstPublicKey.X, firstPublicKey.Y) {
		t.Error("public key is not on curve")
	}

	for _, signer := range signers {
		publicKey := signer.PublicKey()
		if publicKey.X.Cmp(firstPublicKey.X) != 0 || publicKey.Y.Cmp(firstPublicKey.Y) != 0 {
			t.Errorf(
				"public key doesn't match expected\nexpected: [%v]\nactual: [%v]",
				firstPublicKey,
				publicKey,
			)
		}
	}

	// Give it some time to clean up after key generation run.
	time.Sleep(100 * time.Millisecond)

	// Signing initialization.
	message := []byte("message to sign")
	digest := sha256.Sum256(message)

	signingSigners := make(map[group.MemberIndex]*SigningSigner)

	var initSigningWait sync.WaitGroup
	initSigningWait.Add(len(groupMembersKeys))

	for memberID, signer := range signers {
		signingSigner, err := signer.InitializeSigning(
			digest[:],
			networkProviders[memberID],
		)
		if err != nil {
			t.Fatalf("failed to initialize signer: [%v]", err)
		}

		signingSigners[memberID] = signingSigner
	}

	// Signing.
	signatureMutex := sync.Mutex{}
	signatures := []*ecdsa.Signature{}

	signingDone := make(chan interface{})

	go func() {
		var signingWait sync.WaitGroup
		signingWait.Add(groupSize)

		for memberID, signingSigner := range signingSigners {
			go func(memberID group.MemberIndex, signingSigner *SigningSigner) {
				go func() {
					for {
						select {
						case err := <-signingSigner.signingErrChan:
							errChan <- err
							return
						}
					}
				}()

				signature, err := signingSigner.Sign()
				if err != nil {
					errChan <- fmt.Errorf("failed to sign: [%v]", err)
					return
				}

				signatureMutex.Lock()
				signatures = append(signatures, signature)
				signatureMutex.Unlock()

				signingWait.Done()
			}(memberID, signingSigner)
		}

		signingWait.Wait()
		close(signingDone)
	}()

	select {
	case <-signingDone:
	case err := <-errChan:
		t.Fatalf("unexpected error on key generation: [%v]", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	if len(signatures) != groupSize {
		t.Errorf("invalid number of signatures\nexpected: %d\nactual:   %d", groupSize, len(signatures))
	}

	firstSignature := signatures[groupMemberIDs[0]]
	for _, signature := range signatures {
		if !reflect.DeepEqual(firstSignature, signature) {
			t.Errorf(
				"signature doesn't match expected\nexpected: [%v]\nactual: [%v]",
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
