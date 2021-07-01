package client

import (
	"context"
	cecdsa "crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-core/pkg/net"
	"github.com/keep-network/keep-core/pkg/net/key"
	localNet "github.com/keep-network/keep-core/pkg/net/local"
	configtime "github.com/keep-network/keep-ecdsa/config/time"
	"github.com/keep-network/keep-ecdsa/internal/testdata"
	"github.com/keep-network/keep-ecdsa/internal/testhelper"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/bitcoin"
	chainLocal "github.com/keep-network/keep-ecdsa/pkg/chain/local"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss/gen/pb"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss/params"
	"github.com/keep-network/keep-ecdsa/pkg/extensions/tbtc"
	"github.com/keep-network/keep-ecdsa/pkg/extensions/tbtc/recovery"
	"github.com/keep-network/keep-ecdsa/pkg/node"
	"github.com/keep-network/keep-ecdsa/pkg/registry"
)

var (
	keepIDString = "0x4e09cadc7037afa36603138d1c0b76fe2aa5039c"
	keepAddress  = common.HexToAddress(keepIDString)

	depositAddress = common.HexToAddress("0x39122253af729AA39FE886A105B6a580C0d54F80")

	groupSize = 3

	localChain chainLocal.Chain
	keepID     chain.ID
	keep       chain.BondedECDSAKeepHandle

	tbtcHandle chain.TBTCHandle

	networkProviders map[string]net.Provider

	groupMemberIDs       []tss.MemberID
	keepMembersAddresses []common.Address
	signers              map[string]*tss.ThresholdSigner
	signersMutex         *sync.RWMutex
)

func init() {
	localChain = chainLocal.Connect(context.Background())

	signersMutex = &sync.RWMutex{}

	networkProviders = make(map[string]net.Provider)

	var err error
	keepID, err = localChain.UnmarshalID(keepIDString)
	if err != nil {
		panic(err)
	}

	groupMemberIDs, keepMembersAddresses, signers, err = initializeSigners()
	if err != nil {
		panic(err)
	}

	tbtcHandle, err = localChain.TBTCApplicationHandle()
	if err != nil {
		panic(err)
	}
	tbtcHandle.(*chainLocal.TBTCLocalChain).CreateDeposit(depositAddress.String(), keepMembersAddresses)
}

func TestHandleLiquidationRecovery(t *testing.T) {
	if err := log.SetLogLevel("*", "DEBUG"); err != nil {
		t.Fatal(err)
	}

	bitcoinAddresses := []string{
		"1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx",
		"bc1q46uejlhm9vkswfcqs9plvujzzmqjvtfda3mra6",
		"398r9poPaoKJ7vHkaVzNVsXBGRB3mFMXEK",
	}
	bitcoinExtendedPublicKeys := []string{
		"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1",
		"ypub6Xxan668aiJqvh4SVfd7EzqjWvf36gWufTkhWHv3gaxnBh44HpkTi2TTkm1u136qjUxk7F3jGzoyfrGpHvALMgJgbF4WNXpoPu3QYrqogMK",
		"zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9",
	}

	testCases := map[string]struct {
		bitcoinAddressesOrKeys []string
		configureBitcoinHandle func() *localBitcoinConnection
		expectedErrors         []error
	}{
		// bitcoin connection working
		"bitcoin addresses and working bitcoin connection": {
			bitcoinAddressesOrKeys: bitcoinAddresses,
			configureBitcoinHandle: func() *localBitcoinConnection {
				return newLocalBitcoinConnection()
			},
		},
		"bitcoin extended public keys and working bitcoin connection": {
			bitcoinAddressesOrKeys: bitcoinExtendedPublicKeys,
			configureBitcoinHandle: func() *localBitcoinConnection {
				return newLocalBitcoinConnection()
			},
		},
		// bitcoin connection not working: failing IsAddressUnused
		"bitcoin addresses and failing bitcoin call to IsAddressUnused": {
			bitcoinAddressesOrKeys: bitcoinAddresses,
			configureBitcoinHandle: func() *localBitcoinConnection {
				bitcoinHandle := newLocalBitcoinConnection()
				bitcoinHandle.isAddressUnusedError = fmt.Errorf("mocked failure")

				return bitcoinHandle
			},
		},
		"bitcoin extended public keys and failing bitcoin call to IsAddressUnused": {
			bitcoinAddressesOrKeys: bitcoinExtendedPublicKeys,
			configureBitcoinHandle: func() *localBitcoinConnection {
				bitcoinHandle := newLocalBitcoinConnection()
				bitcoinHandle.isAddressUnusedError = fmt.Errorf("mocked failure")

				return bitcoinHandle
			},
		},
		// bitcoin connection not working: failing VbyteFeeFor25Blocks
		"bitcoin addresses and failing bitcoin call to VbyteFeeFor25Blocks": {
			bitcoinAddressesOrKeys: bitcoinAddresses,
			configureBitcoinHandle: func() *localBitcoinConnection {
				bitcoinHandle := newLocalBitcoinConnection()
				bitcoinHandle.vbyteFeeFor25BlocksError = fmt.Errorf("mocked failure")

				return bitcoinHandle
			},
		},
		"bitcoin extended public keys and failing bitcoin call to VbyteFeeFor25Blocks": {
			bitcoinAddressesOrKeys: bitcoinExtendedPublicKeys,
			configureBitcoinHandle: func() *localBitcoinConnection {
				bitcoinHandle := newLocalBitcoinConnection()
				bitcoinHandle.vbyteFeeFor25BlocksError = fmt.Errorf("mocked failure")

				return bitcoinHandle
			},
		},
		// bitcoin connection not working: failing Broadcast
		"bitcoin addresses and failing bitcoin call to Broadcast": {
			bitcoinAddressesOrKeys: bitcoinAddresses,
			configureBitcoinHandle: func() *localBitcoinConnection {
				bitcoinHandle := newLocalBitcoinConnection()
				bitcoinHandle.broadcastError = fmt.Errorf("mocked failure")

				return bitcoinHandle
			},
		},
		"bitcoin extended public keys and failing bitcoin call to Broadcast": {
			bitcoinAddressesOrKeys: bitcoinExtendedPublicKeys,
			configureBitcoinHandle: func() *localBitcoinConnection {
				bitcoinHandle := newLocalBitcoinConnection()
				bitcoinHandle.broadcastError = fmt.Errorf("mocked failure")

				return bitcoinHandle
			},
		},
		// TODO: Add tests to verify logged output:
		// - resolved maxFeePerVByte
		// - logged 5x warn on broadcast failure
		// - cover more failures
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			bitcoinHandle := testData.configureBitcoinHandle()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			doneChan := make(chan interface{})
			errChan := make(chan error)

			var testWait sync.WaitGroup
			testWait.Add(groupSize)

			go func() {
				for i, member := range groupMemberIDs {
					go func(memberID tss.MemberID, index int) {
						operatorPublicKey, err := memberID.PublicKey()
						if err != nil {
							errChan <- err
							return
						}

						electrsURL := "http://fake.electrs.address"
						tbtcConfig := &tbtc.Config{
							Bitcoin: bitcoin.Config{
								BeneficiaryAddress: testData.bitcoinAddressesOrKeys[index],
								ElectrsURL:         &electrsURL,
							},
							LiquidationRecoveryTimeout: configtime.Duration{10 * time.Second},
						}

						networkProvider := networkProviders[memberID.String()]

						tssNode := node.NewNode(localChain, networkProvider, &tss.Config{})

						signersMutex.RLock()
						signer, ok := signers[memberID.String()]
						if !ok {
							t.Fatalf("failed to load signer for member [%s]", memberID)
						}
						signersMutex.RUnlock()

						persistenceMock, keepsRegistry := newTestKeepsRegistry()
						keepsRegistry.RegisterSigner(keepID, signer)
						persistenceMock.MockSigner(0, keepID.String(), signer)

						derivationIndexStorage := newTestDerivationIndexStorage(t)

						if err := handleLiquidationRecovery(
							ctx,
							localChain,
							tbtcHandle,
							bitcoinHandle,
							networkProvider,
							tbtcConfig,
							tssNode,
							operatorPublicKey,
							keep,
							keepsRegistry,
							derivationIndexStorage,
						); err != nil {
							errChan <- fmt.Errorf("handle liquidation recovery failed for member index [%d]: %w", index, err)
						}

						testWait.Done()
					}(member, i)
				}

				testWait.Wait()
				close(doneChan)
			}()

			select {
			case <-doneChan:
				if len(bitcoinHandle.transactions) != groupSize {
					t.Errorf(
						"unexpected number of broadcasted transactions\n"+
							"expected: [%v]\n"+
							"actual:   [%v]",
						groupSize,
						len(bitcoinHandle.transactions),
					)
				}

				for i, transaction := range bitcoinHandle.transactions {
					if transaction != bitcoinHandle.transactions[0] {
						t.Errorf(
							"bitcoin transaction for member [%d] doesn't match first member's\n"+
								"expected: [%s]\n"+
								"actual:   [%s]",
							i,
							bitcoinHandle.transactions[0],
							transaction,
						)
					}

				}
			case err := <-errChan:
				t.Fatalf("unexpected error: %v", err)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
		})
	}
}

func generateMemberKeys() ([]tss.MemberID, []common.Address, error) {
	memberIDs := []tss.MemberID{}
	memberAddresses := []common.Address{}

	for _, memberIDString := range []string{
		"04754b25e1b91dc4006acf17d2c28788be8398a8ed591ba2cbbff5c424d23d91971a8881edd3fc64772d90a181665b4b2ffdbbf05776b8fa8bd08893c26c1cad44",
		"045300560c6c1619d8e2fd4bacc5566c330a89b6402c8c8ceb748d4232b5157dce812ab86645fc66e534a7a3238299eb258245e48a3885d3eea7b885e6c94ddfed",
		"047279cff18c9bdfad9f6f23407070b9ace75acb5570d687de3416a306ecae16a7b40e6f1721f30bcee9b910e8a3d9bb298e9a6540826cf3ae5fbe1163a60d86ec",
	} {
		memberID, err := tss.MemberIDFromString(memberIDString)
		if err != nil {
			return nil, nil, err
		}

		memberIDs = append(memberIDs, memberID)
		memberAddresses = append(memberAddresses, common.HexToAddress(memberID.String()))
	}

	return memberIDs, memberAddresses, nil
}

func initializeSigners() ([]tss.MemberID, []common.Address, map[string]*tss.ThresholdSigner, error) {
	if err := log.SetLogLevel("*", "INFO"); err != nil {
		return nil, nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := log.SetLogLevel("*", "DEBUG"); err != nil {
		return nil, nil, nil, err
	}

	groupMemberIDs, keepMembersAddresses, err := generateMemberKeys()
	if err != nil {
		return nil, nil, nil, err
	}

	keep = localChain.OpenKeep(keepAddress, depositAddress, keepMembersAddresses)

	doneChan := make(chan interface{})
	errChan := make(chan error)

	testData, err := testdata.LoadKeygenTestFixtures(groupSize)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load key gen test fixtures: [%v]", err)
	}

	pubKeyToAddressFn := func(publicKey cecdsa.PublicKey) []byte {
		return elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
	}

	var testWait sync.WaitGroup
	testWait.Add(groupSize)

	var providersInitializedWg sync.WaitGroup
	providersInitializedWg.Add(groupSize)

	signers := make(map[string]*tss.ThresholdSigner)

	go func() {
		for i, memberID := range groupMemberIDs {
			go func(memberID tss.MemberID, index int) {
				operatorPublicKey, err := memberID.PublicKey()
				if err != nil {
					errChan <- err
					return
				}
				networkPublicKey := key.NetworkPublic(*operatorPublicKey)
				networkProvider := localNet.ConnectWithKey(&networkPublicKey)

				networkProviders[memberID.String()] = networkProvider

				providersInitializedWg.Done()
				providersInitializedWg.Wait()

				// FIXME: Load signers from local test data storage instead of
				// generating them with tss.GenerateThresholdSigner.
				//
				// signer, err := newTestSigner(groupMemberIDs, index)
				// if err != nil {
				// 	errChan <- err
				// 	return
				// }
				signer, err := tss.GenerateThresholdSigner(
					ctx,
					keep.ID().String(),
					memberID,
					groupMemberIDs,
					uint(len(groupMemberIDs)-1),
					networkProvider,
					pubKeyToAddressFn,
					params.NewBox(&testData[index].LocalPreParams),
				)
				if err != nil {
					errChan <- err
					return
				}

				signersMutex.Lock()
				signers[memberID.String()] = signer
				signersMutex.Unlock()

				testWait.Done()
			}(memberID, i)
		}

		testWait.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		return groupMemberIDs, keepMembersAddresses, signers, nil
	case err := <-errChan:
		return nil, nil, nil, err
	case <-ctx.Done():
		return nil, nil, nil, ctx.Err()
	}
}

func newTestSigner(groupMemberIDs []tss.MemberID, memberIndex int) (*tss.ThresholdSigner, error) {
	testData, err := testdata.LoadKeygenTestFixtures(memberIndex + 1)
	if err != nil {
		return nil, fmt.Errorf("failed to load key gen test fixtures: [%v]", err)
	}

	thresholdKey := tss.ThresholdKey(testData[memberIndex])
	threshdolKeyBytes, err := thresholdKey.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal threshold key: [%v]", err)
	}

	signer := &tss.ThresholdSigner{}

	groupMemberIDsBytes := [][]byte{
		groupMemberIDs[0],
		groupMemberIDs[1],
		groupMemberIDs[2],
	}

	pbGroup := &pb.ThresholdSigner_GroupInfo{
		GroupID:            "test-group-1",
		MemberID:           groupMemberIDsBytes[memberIndex],
		GroupMemberIDs:     groupMemberIDsBytes,
		DishonestThreshold: 2,
	}
	pbSigner := &pb.ThresholdSigner{
		GroupInfo:    pbGroup,
		ThresholdKey: threshdolKeyBytes,
	}

	bytes, err := proto.Marshal(pbSigner)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal signer: [%v]", err)
	}

	err = signer.Unmarshal(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal signer: [%v]", err)
	}

	return signer, nil
}

func newTestKeepsRegistry() (*testhelper.PersistenceHandleMock, *registry.Keeps) {
	persistenceMock := testhelper.NewPersistenceHandleMock(1)

	return persistenceMock, registry.NewKeepsRegistry(persistenceMock, localChain.UnmarshalID)
}

func newTestDerivationIndexStorage(t *testing.T) *recovery.DerivationIndexStorage {
	dir, err := ioutil.TempDir(t.TempDir(), "test-storage")
	if err != nil {
		t.Fatal(err)
	}

	dis, err := recovery.NewDerivationIndexStorage(dir)
	if err != nil {
		t.Fatal(err)
	}

	return dis
}

// Mock bitcoin connection for testing.
type localBitcoinConnection struct {
	transactions        []string
	vbyteFeeFor25Blocks int32
	isAddressUnused     bool

	broadcastError           error
	vbyteFeeFor25BlocksError error
	isAddressUnusedError     error

	mutex *sync.RWMutex
}

func newLocalBitcoinConnection() *localBitcoinConnection {
	return &localBitcoinConnection{
		transactions:        []string{},
		vbyteFeeFor25Blocks: 34,
		isAddressUnused:     true,
		mutex:               &sync.RWMutex{},
	}
}

func (l *localBitcoinConnection) Broadcast(transaction string) error {
	l.mutex.Lock()
	l.transactions = append(l.transactions, transaction)
	l.mutex.Unlock()

	return l.broadcastError
}

func (l *localBitcoinConnection) VbyteFeeFor25Blocks() (int32, error) {
	return l.vbyteFeeFor25Blocks, l.vbyteFeeFor25BlocksError
}

func (l *localBitcoinConnection) IsAddressUnused(btcAddress string) (bool, error) {
	return l.isAddressUnused, l.isAddressUnusedError
}
