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
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss/params"
	"github.com/keep-network/keep-ecdsa/pkg/extensions/tbtc"
	"github.com/keep-network/keep-ecdsa/pkg/extensions/tbtc/recovery"
	"github.com/keep-network/keep-ecdsa/pkg/node"
	"github.com/keep-network/keep-ecdsa/pkg/registry"
)

func TestHandleLiquidationRecovery(t *testing.T) {
	if err := log.SetLogLevel("*", "DEBUG"); err != nil {
		t.Fatal(err)
	}

	keepAddress := common.HexToAddress("0x4e09cadc7037afa36603138d1c0b76fe2aa5039c")
	depositAddress := common.HexToAddress("0x39122253af729AA39FE886A105B6a580C0d54F80")

	groupSize := 3

	localChain := chainLocal.Connect(context.Background())

	keepID, err := localChain.UnmarshalID(keepAddress.String())
	if err != nil {
		panic(err)
	}

	groupMemberIDs, keepMembersAddresses, signers, networkProviders, err := initializeSigners(groupSize, keepAddress)
	if err != nil {
		panic(err)
	}

	keep := localChain.OpenKeep(keepAddress, depositAddress, keepMembersAddresses)

	tbtcHandle, err := localChain.TBTCApplicationHandle()
	if err != nil {
		panic(err)
	}
	tbtcHandle.(*chainLocal.TBTCLocalChain).CreateDeposit(depositAddress.String(), keepMembersAddresses)
	tbtcHandle.(*chainLocal.TBTCLocalChain).FundDeposit(depositAddress.String())

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
		configureBitcoinHandle func(memberIndex int) *localBitcoinConnection
		expectedErrors         []error
	}{
		// bitcoin connection working
		"bitcoin addresses and working bitcoin connection": {
			bitcoinAddressesOrKeys: bitcoinAddresses,
			configureBitcoinHandle: func(memberIndex int) *localBitcoinConnection {
				return newLocalBitcoinConnection()
			},
		},
		"bitcoin extended public keys and working bitcoin connection": {
			bitcoinAddressesOrKeys: bitcoinExtendedPublicKeys,
			configureBitcoinHandle: func(memberIndex int) *localBitcoinConnection {
				return newLocalBitcoinConnection()
			},
		},
		// bitcoin connection not working: failing IsAddressUnused
		"bitcoin addresses and failing bitcoin call to IsAddressUnused": {
			bitcoinAddressesOrKeys: bitcoinAddresses,
			configureBitcoinHandle: func(memberIndex int) *localBitcoinConnection {
				bitcoinHandle := newLocalBitcoinConnection()
				bitcoinHandle.isAddressUnusedError = fmt.Errorf("mocked failure")

				return bitcoinHandle
			},
		},
		"bitcoin extended public keys and failing bitcoin call to IsAddressUnused": {
			bitcoinAddressesOrKeys: bitcoinExtendedPublicKeys,
			configureBitcoinHandle: func(memberIndex int) *localBitcoinConnection {
				bitcoinHandle := newLocalBitcoinConnection()
				bitcoinHandle.isAddressUnusedError = fmt.Errorf("mocked failure")

				return bitcoinHandle
			},
		},
		// bitcoin connection not working: failing VbyteFeeFor25Blocks
		"bitcoin addresses and failing bitcoin call to VbyteFeeFor25Blocks": {
			bitcoinAddressesOrKeys: bitcoinAddresses,
			configureBitcoinHandle: func(memberIndex int) *localBitcoinConnection {
				bitcoinHandle := newLocalBitcoinConnection()
				bitcoinHandle.vbyteFeeFor25BlocksError = fmt.Errorf("mocked failure")

				return bitcoinHandle
			},
		},
		"bitcoin extended public keys and failing bitcoin call to VbyteFeeFor25Blocks": {
			bitcoinAddressesOrKeys: bitcoinExtendedPublicKeys,
			configureBitcoinHandle: func(memberIndex int) *localBitcoinConnection {
				bitcoinHandle := newLocalBitcoinConnection()
				bitcoinHandle.vbyteFeeFor25BlocksError = fmt.Errorf("mocked failure")

				return bitcoinHandle
			},
		},
		// bitcoin connection not working: failing Broadcast
		"bitcoin addresses and failing bitcoin call to Broadcast": {
			bitcoinAddressesOrKeys: bitcoinAddresses,
			configureBitcoinHandle: func(memberIndex int) *localBitcoinConnection {
				bitcoinHandle := newLocalBitcoinConnection()
				bitcoinHandle.broadcastError = fmt.Errorf("mocked failure")

				return bitcoinHandle
			},
		},
		"bitcoin extended public keys and failing bitcoin call to Broadcast": {
			bitcoinAddressesOrKeys: bitcoinExtendedPublicKeys,
			configureBitcoinHandle: func(memberIndex int) *localBitcoinConnection {
				bitcoinHandle := newLocalBitcoinConnection()
				bitcoinHandle.broadcastError = fmt.Errorf("mocked failure")

				return bitcoinHandle
			},
		},
		// bitcoin connection not working for 1 member:
		"bitcoin addresses and failing bitcoin connection for one member": {
			bitcoinAddressesOrKeys: bitcoinAddresses,
			configureBitcoinHandle: func(memberIndex int) *localBitcoinConnection {
				bitcoinHandle := newLocalBitcoinConnection()
				bitcoinHandle.broadcastError = fmt.Errorf("mocked failure")

				if memberIndex == 2 {
					bitcoinHandle.isAddressUnusedError = fmt.Errorf("mocked failure")
					bitcoinHandle.vbyteFeeFor25BlocksError = fmt.Errorf("mocked failure")
					bitcoinHandle.broadcastError = fmt.Errorf("mocked failure")
				}

				return bitcoinHandle
			},
		},
		"bitcoin extended public keys and failing bitcoin connection for one member": {
			bitcoinAddressesOrKeys: bitcoinExtendedPublicKeys,
			configureBitcoinHandle: func(memberIndex int) *localBitcoinConnection {
				bitcoinHandle := newLocalBitcoinConnection()

				if memberIndex == 2 {
					bitcoinHandle.isAddressUnusedError = fmt.Errorf("mocked failure")
					bitcoinHandle.vbyteFeeFor25BlocksError = fmt.Errorf("mocked failure")
					bitcoinHandle.broadcastError = fmt.Errorf("mocked failure")
				}

				return bitcoinHandle
			},
		},
		// TODO: Add tests to verify logged output:
		// - logged 5x warn on broadcast failure
		// - cover more failures
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			bitcoinHandles := []*localBitcoinConnection{}
			for i := range groupMemberIDs {
				bitcoinHandles = append(bitcoinHandles, testData.configureBitcoinHandle(i))
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

						signer, ok := signers[memberID.String()]
						if !ok {
							errChan <- fmt.Errorf("failed to load signer for member [%s]", memberID)
						}

						persistenceMock, keepsRegistry := newTestKeepsRegistry(localChain)
						keepsRegistry.RegisterSigner(keepID, signer)
						persistenceMock.MockSigner(0, keepID.String(), signer)

						derivationIndexStorage := newTestDerivationIndexStorage(t)

						if err := handleLiquidationRecovery(
							ctx,
							localChain,
							tbtcHandle,
							bitcoinHandles[index],
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
				expectedTransaction := bitcoinHandles[0].transactions[0]

				for i, bitcoinHandle := range bitcoinHandles {
					if len(bitcoinHandle.transactions) != 1 {
						t.Errorf(
							"unexpected number of broadcasted transactions for member [%d]\n"+
								"expected: [%v]\n"+
								"actual:   [%v]",
							i,
							1,
							len(bitcoinHandle.transactions),
						)
					}

					if expectedTransaction != bitcoinHandle.transactions[0] {
						t.Errorf(
							"bitcoin transaction for member [%d] doesn't match first member's\n"+
								"expected: [%s]\n"+
								"actual:   [%s]",
							i,
							expectedTransaction,
							bitcoinHandle.transactions[0],
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

func TestResolveVbyteFee(t *testing.T) {
	previousOutputValue := int32(1000000) // 0.01 BTC

	// We want to check if transaction fee is greater than 5% of the previous
	// output value. We assume that transaction vByte size is 175, so to get
	// a corresponding vByte fee we need to calculate:
	// vByte fee = (output value * 5%) / 175
	fivePercentVbyteFee := int32(285)

	workingBitcoinConnection := func(fee int32) func() *localBitcoinConnection {
		return func() *localBitcoinConnection {
			bitcoinHandle := newLocalBitcoinConnection()
			bitcoinHandle.vbyteFeeFor25Blocks = fee

			return bitcoinHandle
		}
	}

	failingBitconConnection := func() *localBitcoinConnection {
		bitcoinHandle := newLocalBitcoinConnection()
		bitcoinHandle.vbyteFeeFor25BlocksError = fmt.Errorf("mocked failure")

		return bitcoinHandle
	}

	testCases := map[string]struct {
		configureBitcoinHandle func() *localBitcoinConnection
		tbtcConfig             *tbtc.Config
		previousOutputValue    int32
		expectedResult         int32
	}{
		// fee estimate greater than 5% of the previous transaction output value
		"fee returned from bitcoin handle greater than 5% of output value and greater than max fee defined in config": {
			configureBitcoinHandle: workingBitcoinConnection(fivePercentVbyteFee + 2),
			tbtcConfig: &tbtc.Config{
				Bitcoin: bitcoin.Config{
					MaxFeePerVByte: fivePercentVbyteFee + 1,
				},
			},
			previousOutputValue: previousOutputValue,
			expectedResult:      fivePercentVbyteFee + 1,
		},
		"fee returned from bitcoin handle greater than 5% of output value and less than max fee defined in config": {
			configureBitcoinHandle: workingBitcoinConnection(fivePercentVbyteFee + 1),
			tbtcConfig: &tbtc.Config{
				Bitcoin: bitcoin.Config{
					MaxFeePerVByte: fivePercentVbyteFee + 2,
				},
			},
			previousOutputValue: previousOutputValue,
			expectedResult:      fivePercentVbyteFee + 1,
		},
		"fee returned from bitcoin handle greater than 5% of output value and greater than default max fee; max fee not defined in config": {
			configureBitcoinHandle: workingBitcoinConnection(defaultVbyteFee + 1),
			tbtcConfig:             &tbtc.Config{},
			previousOutputValue:    (defaultVbyteFee + 1) * 20,
			expectedResult:         defaultVbyteFee + 1,
		},
		"fee returned from bitcoin handle greater than 5% of output value and less than default max fee; max fee not defined in config": {
			configureBitcoinHandle: workingBitcoinConnection(defaultVbyteFee - 1),
			tbtcConfig:             &tbtc.Config{},
			previousOutputValue:    (defaultVbyteFee - 2) * 20,
			expectedResult:         defaultVbyteFee - 1,
		},
		// fee estimate less than 5% of the previous transaction output value
		"fee returned from bitcoin handle less than 5% of output value and greater than max fee defined in config": {
			configureBitcoinHandle: workingBitcoinConnection(fivePercentVbyteFee - 1),
			tbtcConfig: &tbtc.Config{
				Bitcoin: bitcoin.Config{
					MaxFeePerVByte: fivePercentVbyteFee - 2,
				},
			},
			previousOutputValue: previousOutputValue,
			expectedResult:      fivePercentVbyteFee - 1,
		},
		"fee returned from bitcoin handle less than 5% of output value and less than max fee defined in config": {
			configureBitcoinHandle: workingBitcoinConnection(fivePercentVbyteFee - 2),
			tbtcConfig: &tbtc.Config{
				Bitcoin: bitcoin.Config{
					MaxFeePerVByte: fivePercentVbyteFee - 1,
				},
			},
			previousOutputValue: previousOutputValue,
			expectedResult:      fivePercentVbyteFee - 2,
		},
		"fee returned from bitcoin handle less than 5% of output value and max fee not defined in config": {
			configureBitcoinHandle: workingBitcoinConnection(fivePercentVbyteFee - 1),
			tbtcConfig:             &tbtc.Config{},
			previousOutputValue:    previousOutputValue,
			expectedResult:         fivePercentVbyteFee - 1,
		},
		// non-working bitcoin connection
		"non-working bitcoin handle connection and max fee defined in config": {
			configureBitcoinHandle: failingBitconConnection,
			tbtcConfig: &tbtc.Config{
				Bitcoin: bitcoin.Config{
					MaxFeePerVByte: 25,
				},
			},
			previousOutputValue: previousOutputValue,
			expectedResult:      25,
		},
		"non-working bitcoin handle connection and max fee not defined in config": {
			configureBitcoinHandle: failingBitconConnection,
			tbtcConfig:             &tbtc.Config{},
			previousOutputValue:    previousOutputValue,
			expectedResult:         75,
		},
		// previous output value less than 20 to test if it doesn't blow up on
		// 5% calculation
		"previous output value is less than 20": {
			configureBitcoinHandle: workingBitcoinConnection(50),
			tbtcConfig: &tbtc.Config{
				Bitcoin: bitcoin.Config{
					MaxFeePerVByte: 30,
				},
			},
			previousOutputValue: 19,
			expectedResult:      30,
		},
		// very low previous output value (0.0001 BTC)
		"low previous output value": {
			configureBitcoinHandle: workingBitcoinConnection(20),
			tbtcConfig: &tbtc.Config{
				Bitcoin: bitcoin.Config{
					MaxFeePerVByte: 30,
				},
			},
			previousOutputValue: 10000,
			expectedResult:      20,
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			actual := resolveVbyteFee(testData.configureBitcoinHandle(), testData.tbtcConfig, testData.previousOutputValue)

			if testData.expectedResult != actual {
				t.Errorf(
					"unexpected result\n"+
						"expected: [%v]\n"+
						"actual:   [%v]",
					testData.expectedResult,
					actual,
				)
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

func initializeSigners(
	groupSize int,
	keepAddress common.Address,
) (
	[]tss.MemberID,
	[]common.Address,
	map[string]*tss.ThresholdSigner,
	map[string]net.Provider, error,
) {
	if err := log.SetLogLevel("*", "INFO"); err != nil {
		return nil, nil, nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	groupMemberIDs, keepMembersAddresses, err := generateMemberKeys()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	doneChan := make(chan interface{})
	errChan := make(chan error)

	testData, err := testdata.LoadKeygenTestFixtures(groupSize)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to load key gen test fixtures: [%v]", err)
	}

	pubKeyToAddressFn := func(publicKey cecdsa.PublicKey) []byte {
		return elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
	}

	var testWait sync.WaitGroup
	testWait.Add(groupSize)

	var providersInitializedWg sync.WaitGroup
	providersInitializedWg.Add(groupSize)

	signers := make(map[string]*tss.ThresholdSigner)
	signersMutex := &sync.Mutex{}

	networkProviders := make(map[string]net.Provider)
	networkProvidersMutex := &sync.Mutex{}

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

				networkProvidersMutex.Lock()
				networkProviders[memberID.String()] = networkProvider
				networkProvidersMutex.Unlock()

				providersInitializedWg.Done()
				providersInitializedWg.Wait()

				// FIXME: Load signers from local test data storage instead of
				// generating them with tss.GenerateThresholdSigner.
				signer, err := tss.GenerateThresholdSigner(
					ctx,
					keepAddress.String(),
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
		return groupMemberIDs, keepMembersAddresses, signers, networkProviders, nil
	case err := <-errChan:
		return nil, nil, nil, nil, err
	case <-ctx.Done():
		return nil, nil, nil, nil, ctx.Err()
	}
}

func newTestKeepsRegistry(localChain chain.Handle) (*testhelper.PersistenceHandleMock, *registry.Keeps) {
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
	if l.vbyteFeeFor25BlocksError != nil {
		return 0, l.vbyteFeeFor25BlocksError
	}

	return l.vbyteFeeFor25Blocks, nil
}

func (l *localBitcoinConnection) IsAddressUnused(btcAddress string) (bool, error) {
	if l.isAddressUnusedError != nil {
		return true, nil
	}

	return l.isAddressUnused, l.isAddressUnusedError
}
