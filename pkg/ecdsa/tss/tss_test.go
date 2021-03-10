package tss

import (
	"context"
	cecdsa "crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/keep-network/keep-ecdsa/pkg/utils/byteutils"

	"github.com/keep-network/keep-core/pkg/operator"

	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-core/pkg/net"
	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-core/pkg/net/local"
	"github.com/keep-network/keep-ecdsa/internal/testdata"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss/params"
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

	pubKeyToAddressFn := func(publicKey cecdsa.PublicKey) []byte {
		return elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
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
					pubKeyToAddressFn,
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

	// S256 function is defined in respective `tss-*-test.go` files. The file
	// in use depends on the passed build tags.
	curve := S256()

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
					pubKeyToAddressFn,
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
		firstPublicKey,
		digest[:],
		firstSignature.R,
		firstSignature.S,
	) {
		t.Errorf("invalid signature: [%+v]", firstSignature)
	}

	verifyEthereumSignature(t, digest[:], firstSignature, firstPublicKey)
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

// verifyEthereumSignature validates that signature in form (r, s, recoveryID)
// is a valid ethereum signature. `SigToPub` is a wrapper on `Ecrecover` that
// allows us to validate a signature in the same way as it's done on-chain, we
// extract public key from the signature and compare it with signer's public key.
func verifyEthereumSignature(
	t *testing.T,
	hash []byte,
	signature *ecdsa.Signature,
	expectedPublicKey *cecdsa.PublicKey,
) {

	serializedSignature, err := serializeSignature(signature)
	if err != nil {
		t.Fatalf("failed to serialize signature: [%v]", err)
	}

	// SigToPub function is defined in respective `tss-*-test.go` files. The
	// file in use depends on the passed build tags.
	publicKey, err := SigToPub(hash, serializedSignature)
	if err != nil {
		t.Fatalf("failed to get public key from signature: [%v]", err)
	}

	if expectedPublicKey.X.Cmp(publicKey.X) != 0 ||
		expectedPublicKey.Y.Cmp(publicKey.Y) != 0 {
		t.Errorf(
			"invalid public key:\nexpected: [%x]\nactual:   [%x]\n",
			expectedPublicKey,
			publicKey,
		)
	}
}

func serializeSignature(signature *ecdsa.Signature) ([]byte, error) {
	var serializedBytes []byte

	r, err := byteutils.LeftPadTo32Bytes(signature.R.Bytes())
	if err != nil {
		return nil, err
	}

	s, err := byteutils.LeftPadTo32Bytes(signature.S.Bytes())
	if err != nil {
		return nil, err
	}

	recoveryID := byte(signature.RecoveryID)

	serializedBytes = append(serializedBytes, r...)
	serializedBytes = append(serializedBytes, s...)
	serializedBytes = append(serializedBytes, recoveryID)

	return serializedBytes, nil
}
