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

	"github.com/keep-network/keep-core/pkg/operator"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ipfs/go-log/v2"
	"github.com/keep-network/keep-core/pkg/net"
	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-core/pkg/net/local"
	"github.com/keep-network/keep-ecdsa/internal/testdata"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss/params"
	"github.com/keep-network/keep-ecdsa/pkg/utils/testutils"
)

func TestGenerateKeyAndSign(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	groupSize := 5
	dishonestThreshold := uint(groupSize - 1)
	groupID := fmt.Sprintf("tss-test-%d", rand.Int())

	err := log.SetLogLevel("*", "DEBUG")
	if err != nil {
		t.Fatalf("logger initialization failed: [%v]", err)
	}

	errChan := make(chan error)

	groupMemberIDs, err := generateMemberKeys(groupSize)
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
	signers := make(map[string]*ThresholdSigner)

	keyGenDone := make(chan interface{})

	go func() {
		var keyGenWait sync.WaitGroup
		keyGenWait.Add(groupSize)

		var providersInitializedWg sync.WaitGroup
		providersInitializedWg.Add(groupSize)
		providersInitialized := make(chan struct{})

		go func() {
			providersInitializedWg.Wait()
			close(providersInitialized)
		}()

		for i, memberID := range groupMemberIDs {
			go func(memberID MemberID, index int) {
				memberPublicKey, err := memberID.PublicKey()
				if err != nil {
					errChan <- err
					return
				}

				networkPublicKey := key.NetworkPublic(*memberPublicKey)
				network := newTestNetProvider(&networkPublicKey)
				networkProviders.Store(memberID.String(), network)
				providersInitializedWg.Done()
				<-providersInitialized

				preParams := testData[index].LocalPreParams

				signer, err := GenerateThresholdSigner(
					ctx,
					groupID,
					memberID,
					groupMemberIDs,
					dishonestThreshold,
					network,
					params.NewBox(&preParams),
				)
				if err != nil {
					errChan <- fmt.Errorf("failed to generate signer: [%v]", err)
				}

				signersMutex.Lock()
				signers[memberID.String()] = signer
				signersMutex.Unlock()

				keyGenWait.Done()
			}(memberID, i)
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

	firstSigner := signers[groupMemberIDs[0].String()]
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

		for memberIDString, signer := range signers {
			memberID, _ := MemberIDFromString(memberIDString)

			go func(memberID MemberID, signer *ThresholdSigner) {
				value, loaded := networkProviders.Load(memberID.String())
				if !loaded {
					errChan <- fmt.Errorf("failed to load network provider")
					return
				}
				networkProvider := value.(net.Provider)

				signature, err := signer.CalculateSignature(
					ctx,
					digest[:],
					networkProvider,
				)
				if err != nil {
					errChan <- fmt.Errorf("failed to sign: [%v]", err)
					return
				}

				signatureMutex.Lock()
				signatures[memberID.String()] = signature
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

	firstSignature := signatures[groupMemberIDs[0].String()]
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

func generateMemberKeys(groupSize int) ([]MemberID, error) {
	memberIDs := []MemberID{}

	for i := 0; i < groupSize; i++ {
		_, publicKey, err := operator.GenerateKeyPair()
		if err != nil {
			return nil, fmt.Errorf("failed to generate operator key: [%v]", err)
		}

		memberIDs = append(memberIDs, MemberIDFromPublicKey(publicKey))
	}

	return memberIDs, nil
}

func newTestNetProvider(memberNetworkKey *key.NetworkPublic) net.Provider {
	return local.ConnectWithKey(memberNetworkKey)
}
