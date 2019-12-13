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
	dishonestThreshold := uint(groupSize - 1)
	groupID := fmt.Sprintf("tss-test-%d", rand.Int())

	err := log.SetLogLevel("*", "INFO")
	if err != nil {
		t.Fatalf("logger initialization failed: [%v]", err)
	}

	errChan := make(chan error)

	groupMemberIDs, groupMembersKeys, err := generateMemberKeys(groupSize)
	if err != nil {
		t.Fatalf("failed to generate members keys: [%v]", err)
	}

	testData, err := testdata.LoadKeygenTestFixtures(groupSize)
	if err != nil {
		t.Fatalf("failed to load test data: [%v]", err)
	}

	networkProviders := &sync.Map{} // < MemberID, net.Provider >

	// Key generation.
	signersMutex := sync.Mutex{}
	signers := make(map[MemberID]*ThresholdSigner)

	keyGenDone := make(chan interface{})
	KeyGenSync.Add(groupSize) // TODO: Temp Sync

	go func() {
		var keyGenWait sync.WaitGroup
		keyGenWait.Add(groupSize)

		for i, memberID := range groupMemberIDs {
			go func(memberID MemberID) {
				keygenErrChan := make(chan error)
				go func() {
					for {
						select {
						case err := <-keygenErrChan:
							errChan <- err
							return
						}
					}
				}()

				network := newTestNetProvider(memberID, groupMembersKeys, keygenErrChan)
				networkProviders.Store(memberID, network)

				preParams := testData[i].LocalPreParams

				signer, err := GenerateThresholdSigner(
					groupID,
					memberID,
					groupMemberIDs,
					dishonestThreshold,
					network,
					&preParams,
				)
				if err != nil {
					keygenErrChan <- fmt.Errorf("failed to generate signer: [%v]", err)
				}

				signersMutex.Lock()
				signers[memberID] = signer
				signersMutex.Unlock()

				keyGenWait.Done()
			}(memberID)
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

	// Signing.
	message := []byte("message to sign")
	digest := sha256.Sum256(message)

	signatureMutex := sync.Mutex{}
	signatures := make(map[MemberID]*ecdsa.Signature)

	signingDone := make(chan interface{})

	SigningSync.Add(groupSize) // TODO: Temp Sync

	go func() {
		var signingWait sync.WaitGroup
		signingWait.Add(groupSize)

		for memberID, signer := range signers {
			go func(memberID MemberID, signer *ThresholdSigner) {
				signingErrChan := make(chan error)

				go func() {
					for {
						select {
						case err := <-signingErrChan:
							errChan <- err
							return
						}
					}
				}()

				value, loaded := networkProviders.Load(memberID)
				if !loaded {
					errChan <- fmt.Errorf("failed to load network provider")
					return
				}
				networkProvider := value.(net.Provider)

				signature, err := signer.CalculateSignature(
					digest[:],
					networkProvider,
				)
				if err != nil {
					errChan <- fmt.Errorf("failed to sign: [%v]", err)
					return
				}

				signatureMutex.Lock()
				signatures[memberID] = signature
				signatureMutex.Unlock()

				signingWait.Done()
			}(memberID, signer)
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

func generateMemberKeys(groupSize int) ([]MemberID, map[MemberID]*key.NetworkPublic, error) {
	memberIDs := []MemberID{}
	groupMembersKeys := make(map[MemberID]*key.NetworkPublic, groupSize)

	for i := 0; i < groupSize; i++ {
		_, publicKey, err := key.GenerateStaticNetworkKey()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate network key: [%v]", err)
		}

		memberID := MemberID(key.NetworkPubKeyToEthAddress(publicKey))

		memberIDs = append(memberIDs, memberID)
		groupMembersKeys[memberID] = publicKey
	}

	return memberIDs, groupMembersKeys, nil
}

func newTestNetProvider(
	memberID MemberID,
	membersNetworkKeys map[MemberID]*key.NetworkPublic,
	errChan chan error,
) net.Provider {
	return local.LocalProvider(
		membersNetworkKeys[memberID],
		errChan,
	)
}
