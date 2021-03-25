package recovery

import (
	"bytes"
	"context"
	cecdsa "crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-core/pkg/net/local"
	"github.com/keep-network/keep-core/pkg/operator"
	"github.com/keep-network/keep-ecdsa/internal/testdata"
	lc "github.com/keep-network/keep-ecdsa/pkg/chain/local"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss/params"
	"gotest.tools/v3/assert"
)

func TestPublicKeyToP2WPKHScriptCode(t *testing.T) {
	// Test based on test values from BIP143:
	// https://github.com/bitcoin/bips/blob/master/bip-0143.mediawiki#native-p2wpkh
	serializedPublicKey, _ := hex.DecodeString("025476c2e83188368da1ff3e292e7acafcdb3566bb0ad253f62fc70f07aeee6357")
	expectedScriptCode, _ := hex.DecodeString("1976a9141d0f172a0ecb48aee1be1f2687d2963ae33f71a188ac")

	publicKey, _ := btcec.ParsePubKey(serializedPublicKey, btcec.S256())

	scriptCodeBytes, err := publicKeyToP2WPKHScriptCode(publicKey.ToECDSA(), &chaincfg.MainNetParams)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bytes.Compare(expectedScriptCode, scriptCodeBytes) != 0 {
		t.Errorf(
			"unexpected script code\nexpected: %x\nactual:   %x",
			expectedScriptCode,
			scriptCodeBytes,
		)
	}
}

func TestConstructUnsignedTransaction(t *testing.T) {
	recipientAddresses := []string{
		"bcrt1q5sz7jly79m76a5e8py6kv402q07p725vm4s0zl",
		"bcrt1qlxt5a04pefwkl90mna2sn79nu7asq3excx60h0",
		"bcrt1qjhpgmmhaxfwj6t7zf3dvs2fhdhx02g8qn3xwsf",
	}

	previousOutputValue := int64(100000000)

	expectedTxHex := "01000000000101f19194baa0d12141a177f41ea218d93d10e2cf96921e009199215f65a9de990b000000000000000000039003fc0100000000160014a405e97c9e2efdaed32709356655ea03fc1f2a8c9003fc0100000000160014f9974ebea1ca5d6f95fb9f5509f8b3e7bb0047269003fc010000000016001495c28deefd325d2d2fc24c5ac829376dccf520e0024a00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002100000000000000000000000000000000000000000000000000000000000000000000000000"
	expectedTxBytes, err := hex.DecodeString(expectedTxHex)
	if err != nil {
		t.Fatal(err)
	}
	expectedTx := wire.NewMsgTx(0)
	expectedTx.Deserialize(bytes.NewReader(expectedTxBytes))

	actualTx, err := constructUnsignedTransaction(
		"0b99dea9655f219991001e9296cfe2103dd918a21ef477a14121d1a0ba9491f1",
		uint32(0),
		previousOutputValue,
		int64(700),
		recipientAddresses,
		&chaincfg.TestNet3Params,
	)
	if err != nil {
		t.Error(err)
	}

	assert.DeepEqual(t, actualTx, expectedTx)
}

func TestBuildSignedTransactionHexString(t *testing.T) {
	unsignedTxHex := "01000000000101f19194baa0d12141a177f41ea218d93d10e2cf96921e009199215f65a9de990b000000000000000000039003fc0100000000160014a405e97c9e2efdaed32709356655ea03fc1f2a8c9003fc0100000000160014f9974ebea1ca5d6f95fb9f5509f8b3e7bb0047269003fc010000000016001495c28deefd325d2d2fc24c5ac829376dccf520e0024a00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002100000000000000000000000000000000000000000000000000000000000000000000000000"
	expectedSignedTx := "01000000000101f19194baa0d12141a177f41ea218d93d10e2cf96921e009199215f65a9de990b000000000000000000039003fc0100000000160014a405e97c9e2efdaed32709356655ea03fc1f2a8c9003fc0100000000160014f9974ebea1ca5d6f95fb9f5509f8b3e7bb0047269003fc010000000016001495c28deefd325d2d2fc24c5ac829376dccf520e0020930060201030201070121020000000007de3ebb640d2b021590c09d5e739597d02d939224d227a17403607500000000"

	publicKey := &cecdsa.PublicKey{
		Curve: elliptic.P224(),
		X:     bigIntFromString(t, "828612351041249926199933036276541218289243364325366441967565889653"),
		Y:     bigIntFromString(t, "985040320797760939221216987624001720525496952574017416820319442840"),
	}

	signature := &ecdsa.Signature{
		R:          big.NewInt(int64(3)),
		S:          big.NewInt(int64(7)),
		RecoveryID: 1,
	}

	signedTxHex, err := buildSignedTransactionHexString(
		decodeTransaction(t, unsignedTxHex),
		signature,
		publicKey,
	)
	if err != nil {
		t.Fatalf("failed to build signed transaction: %v", err)
	}

	if signedTxHex != expectedSignedTx {
		t.Errorf(
			"invalid signed transaction\n- actual\n+ expected\n%s",
			cmp.Diff(decodeTransaction(t, signedTxHex), decodeTransaction(t, expectedSignedTx)))
	}
}

func decodeTransaction(t *testing.T, txHex string) *wire.MsgTx {
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		t.Fatalf("failed to decode transaction [%s]: [%v]", txHex, err)
	}
	tx := wire.NewMsgTx(0)
	tx.BtcDecode(bytes.NewReader(txBytes), wire.ProtocolVersion, wire.WitnessEncoding)

	return tx
}

func bigIntFromString(t *testing.T, s string) *big.Int {
	bigInt, ok := new(big.Int).SetString(s, 0)
	if !ok {
		t.Errorf("Something went wrong creating a big int from %s", s)
	}
	return bigInt
}

func TestBuildBitcoinTransaction(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	localChain := lc.Connect(ctx)
	tbtcHandle := lc.NewTBTCLocalChain(ctx)

	err := log.SetLogLevel("*", "INFO")
	if err != nil {
		t.Fatalf("logger initialization failed: [%v]", err)
	}

	groupSize := 3

	groupMembers, err := generateMemberKeys(groupSize)
	if err != nil {
		t.Fatalf("failed to generate members keys: [%v]", err)
	}

	testData, err := testdata.LoadKeygenTestFixtures(groupSize)
	if err != nil {
		t.Fatalf("failed to load test data: [%v]", err)
	}

	pubKeyToAddressFn := func(publicKey cecdsa.PublicKey) []byte {
		return elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
	}

	groupMemberAddresses := make([]string, groupSize)
	for i, member := range groupMembers {
		pubKey, err := member.PublicKey()
		if err != nil {
			t.Fatalf("could not get member pubkey: [%v]", err)
		}
		groupMemberAddresses[i] = hex.EncodeToString(pubKeyToAddressFn(*pubKey))
	}

	errChan := make(chan error)

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(groupSize)

	keepAddress := common.Address([20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	memberAddresses := make([]common.Address, groupSize)
	for i, memberAddress := range groupMemberAddresses {
		memberAddresses[i] = common.HexToAddress(memberAddress)
	}
	depositAddressString := "0xa5FA806723A7c7c8523F33c39686f20b52612877"
	depositAddress := common.HexToAddress(depositAddressString)
	tbtcHandle.CreateDeposit(depositAddressString, memberAddresses)
	keep := tbtcHandle.OpenKeep(keepAddress, depositAddress, memberAddresses)

	btcAddresses := []string{
		"1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx",
		"1EEX8qZnTw1thadyxsueV748v3Y6tTMccc",
		"1EZuKz6RrJ6XmBPvFwJiEcREpaEVhUVAt5",
	}
	maxFeePerVByte := int32(73)

	mutex := &sync.RWMutex{}

	result := make(map[string]string)

	for i, memberID := range groupMembers {
		go func(memberID tss.MemberID, index int) {
			memberPublicKey, err := memberID.PublicKey()
			if err != nil {
				errChan <- err
				return
			}

			memberNetworkKey := key.NetworkPublic(*memberPublicKey)
			networkProvider := local.ConnectWithKey(&memberNetworkKey)

			defer waitGroup.Done()

			preParams := testData[0].LocalPreParams

			signer, err := tss.GenerateThresholdSigner(
				ctx,
				keep.ID().Hex(),
				memberID,
				groupMembers,
				uint(len(groupMembers)-1),
				networkProvider,
				pubKeyToAddressFn,
				params.NewBox(&preParams),
			)
			if err != nil {
				errChan <- err
				return
			}

			signedHexString, err := BuildBitcoinTransaction(
				ctx,
				networkProvider,
				localChain,
				tbtcHandle,
				keep,
				signer,
				btcAddresses,
				maxFeePerVByte,
			)
			if err != nil {
				errChan <- err
				return
			}

			mutex.Lock()
			result[memberID.String()] = signedHexString
			mutex.Unlock()
		}(memberID, i)
	}

	go func() {
		waitGroup.Wait()
		cancel()
	}()

	select {
	case <-ctx.Done():
		if len(result) != groupSize {
			t.Errorf(
				"invalid number of results\nexpected: [%d]\nactual:  [%d]",
				groupSize,
				len(result),
			)
		}

		expectedSignature := ""
		for _, memberID := range groupMembers {
			if memberResult, ok := result[memberID.String()]; ok {
				if memberResult != expectedSignature {
					t.Errorf(
						"unexpected signed hex string\nexpected: %s\nactual:   %s",
						expectedSignature,
						memberResult,
					)
				}
			} else {
				t.Errorf("missing result for member [%v]", memberID)
			}
		}
	case err := <-errChan:
		t.Fatal(err)
	}
}

func generateMemberKeys(groupSize int) ([]tss.MemberID, error) {
	memberIDs := []tss.MemberID{}

	for i := 0; i < groupSize; i++ {
		_, publicKey, err := operator.GenerateKeyPair()
		if err != nil {
			return nil, fmt.Errorf("failed to generate operator key: [%v]", err)
		}

		memberIDs = append(memberIDs, tss.MemberIDFromPublicKey(publicKey))
	}

	return memberIDs, nil
}
