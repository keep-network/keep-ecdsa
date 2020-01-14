package tss

import (
	"context"
	cecdsa "crypto/ecdsa"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-tecdsa/internal/testdata"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-tecdsa/pkg/net"
	"github.com/keep-network/keep-tecdsa/pkg/net/local"
	"github.com/keep-network/keep-tecdsa/pkg/utils/testutils"
)

func TestGenerateKeyAndSign(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	groupSize := 5
	localCount := 3
	dishonestThreshold := uint(groupSize - 1)

	groupID := fmt.Sprintf("tss-test-%d", rand.Int())

	err := log.SetLogLevel("*", "INFO")
	if err != nil {
		t.Fatalf("logger initialization failed: [%v]", err)
	}

	errChan := make(chan error)

	testMembers, err := generateTestMembers(groupSize, localCount)
	if err != nil {
		t.Fatalf("failed to generate test members: [%v]", err)
	}

	testData, err := testdata.LoadKeygenTestFixtures(groupSize)
	if err != nil {
		t.Fatalf("failed to load test data: [%v]", err)
	}

	// Key generation.
	signersMutex := sync.Mutex{}
	signers := make(map[string]*ThresholdSigner)

	keyGenDone := make(chan interface{})

	go func() {
		var keyGenWait sync.WaitGroup
		keyGenWait.Add(groupSize)

		for i, tm := range testMembers {
			go func(tm *testMember) {
				preParams := testData[i].LocalPreParams

				signer, err := GenerateThresholdSigner(
					groupID,
					tm.memberID,
					testMembers.memberIDs(),
					dishonestThreshold,
					testMembers.groupNetworkIDs(),
					tm.networkProvider,
					&preParams,
				)
				if err != nil {
					errChan <- fmt.Errorf("failed to generate signer: [%v]", err)
				}

				signersMutex.Lock()
				signers[tm.memberID.String()] = signer
				signersMutex.Unlock()

				keyGenWait.Done()
			}(tm)
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

	firstSigner := signers[testMembers[0].memberID.String()]
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

	// Signing.
	message := []byte("message to sign")
	digest := sha256.Sum256(message)

	signatureMutex := sync.Mutex{}
	signatures := make(map[string]*ecdsa.Signature)

	signingDone := make(chan interface{})

	go func() {
		var signingWait sync.WaitGroup
		signingWait.Add(groupSize)

		for _, tm := range testMembers {
			go func(tm *testMember) {
				signer := signers[tm.memberID.String()]

				signature, err := signer.CalculateSignature(
					digest[:],
					tm.networkProvider,
				)
				if err != nil {
					errChan <- fmt.Errorf("failed to sign: [%v]", err)
					return
				}

				signatureMutex.Lock()
				signatures[tm.memberID.String()] = signature
				signatureMutex.Unlock()

				signingWait.Done()
			}(tm)
		}

		signingWait.Wait()
		close(signingDone)
	}()

	select {
	case <-signingDone:
	case err := <-errChan:
		t.Fatalf("unexpected error on signing: [%v]", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	if len(signatures) != groupSize {
		t.Errorf("invalid number of signatures\nexpected: %d\nactual:   %d", groupSize, len(signatures))
	}

	firstSignature := signatures[testMembers[0].memberID.String()]
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

type testMember struct {
	memberID        MemberID
	networkProvider net.Provider
	networkID       net.TransportIdentifier
}

type testMembers []*testMember

func (tms testMembers) memberIDs() []MemberID {
	memberIDs := make([]MemberID, len(tms))

	for i, tm := range tms {
		memberIDs[i] = tm.memberID
	}

	return memberIDs
}

// Generates test members group of given size. The group contains specific number
// of members operating with the same network public key which simulates running
// multiple members by one operator.
func generateTestMembers(groupSize int, localMembersCount int) (testMembers, error) {
	members := make(testMembers, groupSize)

	for i := 0; i < groupSize-localMembersCount; i++ {
		_, publicKey, err := key.GenerateStaticNetworkKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate network key: [%v]", err)
		}

		networkProvider := newTestNetProvider(publicKey)
		memberID := MemberID(i + 1)

		members[i] = &testMember{
			memberID:        memberID,
			networkProvider: networkProvider,
		}
	}

	if localMembersCount > 0 {
		_, publicKey, err := key.GenerateStaticNetworkKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate network key: [%v]", err)
		}
		networkProvider := newTestNetProvider(publicKey)

		for i := groupSize - localMembersCount; i < groupSize; i++ {
			memberID := MemberID(i + 1)

			members[i] = &testMember{
				memberID:        memberID,
				networkProvider: networkProvider,
			}
		}
	}

	return members, nil
}

func (tms testMembers) groupNetworkIDs() map[string]net.TransportIdentifier {
	networkIDs := make(map[string]net.TransportIdentifier, len(tms))

	for _, tm := range tms {
		networkIDs[tm.memberID.String()] = tm.networkProvider.ID()
	}

	return networkIDs
}

func newTestNetProvider(memberNetworkKey *key.NetworkPublic) net.Provider {
	return local.LocalProvider(memberNetworkKey)
}
