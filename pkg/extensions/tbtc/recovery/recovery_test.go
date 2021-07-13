package recovery

import (
	"bytes"
	"context"
	cecdsa "crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-core/pkg/net/local"
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
	expectedScriptCode, _ := hex.DecodeString("76a9141d0f172a0ecb48aee1be1f2687d2963ae33f71a188ac")

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

func TestBuildBitcoinTransaction(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	localChain := lc.Connect(ctx)
	tbtcHandle := lc.NewTBTCLocalChain(ctx)

	err := log.SetLogLevel("*", "INFO")
	if err != nil {
		t.Fatalf("logger initialization failed: [%v]", err)
	}

	groupSize := 3

	groupMembers := make([]tss.MemberID, groupSize)
	for i, memberIDString := range []string{
		"04754b25e1b91dc4006acf17d2c28788be8398a8ed591ba2cbbff5c424d23d91971a8881edd3fc64772d90a181665b4b2ffdbbf05776b8fa8bd08893c26c1cad44",
		"045300560c6c1619d8e2fd4bacc5566c330a89b6402c8c8ceb748d4232b5157dce812ab86645fc66e534a7a3238299eb258245e48a3885d3eea7b885e6c94ddfed",
		"047279cff18c9bdfad9f6f23407070b9ace75acb5570d687de3416a306ecae16a7b40e6f1721f30bcee9b910e8a3d9bb298e9a6540826cf3ae5fbe1163a60d86ec",
	} {
		memberID, err := tss.MemberIDFromString(memberIDString)
		if err != nil {
			t.Fatal(err)
		}
		groupMembers[i] = memberID
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

	fundingInfo, err := tbtcHandle.FundingInfo(depositAddress.String())
	if err != nil {
		t.Fatal(err)
	}

	btcAddresses := []string{
		"1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx",
		"3EktnHQD7RiAE6uzMj2ZifT9YgRrkSgzQX",
		"bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
	}
	maxFeePerVByte := int32(73)

	mutex := &sync.RWMutex{}

	btcTransactions := make(map[string]string)

	var providersInitializedWg sync.WaitGroup
	providersInitializedWg.Add(groupSize)

	for i, memberID := range groupMembers {
		go func(memberID tss.MemberID, index int) {
			memberPublicKey, err := memberID.PublicKey()
			if err != nil {
				errChan <- err
				return
			}

			memberNetworkKey := key.NetworkPublic(*memberPublicKey)
			networkProvider := local.ConnectWithKey(&memberNetworkKey)

			providersInitializedWg.Done()
			providersInitializedWg.Wait()

			defer waitGroup.Done()

			preParams := testData[index].LocalPreParams

			signer, err := tss.GenerateThresholdSigner(
				ctx,
				keep.ID().String(),
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

			signedBtcTransaction, err := BuildBitcoinTransaction(
				ctx,
				networkProvider,
				localChain,
				fundingInfo,
				signer,
				&chaincfg.MainNetParams,
				btcAddresses,
				maxFeePerVByte,
			)
			if err != nil {
				errChan <- err
				return
			}

			mutex.Lock()
			btcTransactions[memberID.String()] = signedBtcTransaction
			mutex.Unlock()
		}(memberID, i)
	}

	go func() {
		waitGroup.Wait()
		cancel()
	}()

	select {
	case <-ctx.Done():
		if len(btcTransactions) != groupSize {
			t.Errorf(
				"invalid number of results\nexpected: [%d]\nactual:  [%d]",
				groupSize,
				len(btcTransactions),
			)
		}

		firstBtcTransaction := btcTransactions[groupMembers[0].String()]
		decodedTransaction := decodeTransaction(t, firstBtcTransaction)
		if len(decodedTransaction.TxOut) != 3 {
			t.Errorf("wrong number of outputs\nexpected: 1\nactual:   %d", len(decodedTransaction.TxOut))
		}

		expectedOutputValue := int64(3329050) // (original deposit of 10000000 - fee) / 3
		for _, outputTransaction := range decodedTransaction.TxOut {
			if outputTransaction.Value != expectedOutputValue {
				t.Errorf(
					"incorrect output value\nexpected: %d\nactual:   %d",
					expectedOutputValue,
					outputTransaction.Value,
				)
			}
		}

		validateTransaction(t, decodedTransaction, int64(10000000)) // original deposit amount

		for _, memberID := range groupMembers {
			if memberResult, ok := btcTransactions[memberID.String()]; ok {
				if memberResult != firstBtcTransaction {
					t.Errorf(
						"hex strings must all be identical\nexpected: %s\nactual:   %s",
						firstBtcTransaction,
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

func validateTransaction(t *testing.T, transaction *wire.MsgTx, previousOutputValue int64) {
	subscript, err := txscript.ComputePkScript(transaction.TxIn[0].SignatureScript, transaction.TxIn[0].Witness)
	if err != nil {
		t.Fatalf("failed to decode pk script: [%v]", err)
	}

	if len(transaction.TxIn) != 1 {
		t.Error("only transactions with one input are supported")
	}
	inputIndex := 0

	validationEngine, err := txscript.NewEngine(
		subscript.Script(),
		transaction,
		inputIndex,
		txscript.StandardVerifyFlags,
		nil,
		nil,
		previousOutputValue,
	)
	if err != nil {
		t.Fatalf(
			"failed to create validation engine: [%v]",
			err,
		)
	}

	if err := validationEngine.Execute(); err != nil {
		t.Errorf(
			"failed to validate transaction: [%v]",
			err,
		)
	}
}
