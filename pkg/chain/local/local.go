package local

import (
	"context"
	cecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	corechain "github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-core/pkg/chain/local"
	"github.com/keep-network/keep-ecdsa/pkg/chain"

	commonLocal "github.com/keep-network/keep-common/pkg/chain/local"
)

// TestingTBTCHandle is an extension of the chain.TBTCHandle interface which
// exposes additional functions useful for testing.
type TestingTBTCHandle interface {
	chain.TBTCHandle

	CreateDeposit(depositAddress string, signers []common.Address)
	DepositPubkey(depositAddress string) ([]byte, error)
	DepositRedemptionSignature(depositAddress string) (*Signature, error)
	DepositRedemptionProof(depositAddress string) (*TxProof, error)
	DepositRedemptionFee(depositAddress string) (*big.Int, error)
	RedeemDeposit(depositAddress string) error

	SetAlwaysFailingTransactions(transactionNames ...string)

	Logger() *localChainLogger
}

// TestingBondedECDSAKeepHandle is an extension of the
// chain.BondedECDSAKeepHandle interface which exposes additional functions
// useful for testing.
type TestingBondedECDSAKeepHandle interface {
	chain.BondedECDSAKeepHandle

	RequestSignature(digest [32]byte) error
}

// TestingChain is an extention of chain.Handle interface which exposes
// additional functions useful for testing.
type TestingChain interface {
	chain.Handle

	TestingTBTC() TestingTBTCHandle

	OpenKeep(keepAddress common.Address, members []common.Address) TestingBondedECDSAKeepHandle
	CloseKeep(keepAddress chain.KeepID) error
	TerminateKeep(keepID chain.KeepID) error
	AuthorizeOperator(operatorAddress common.Address)
}

// localChain is an implementation of ethereum blockchain interface.
//
// It mocks the behaviour of a real blockchain, without the complexity of deployments,
// accounts, async transactions and so on. For use in tests ONLY.
type localChain struct {
	localChainMutex sync.Mutex

	blockCounter     corechain.BlockCounter
	blocksTimestamps sync.Map

	keepAddresses []chain.KeepID
	keeps         map[chain.KeepID]*localKeep

	keepCreatedHandlers map[int]func(event *chain.BondedECDSAKeepCreatedEvent)

	operatorKey *cecdsa.PrivateKey
	signer      corechain.Signing

	authorizations map[chain.OperatorID]bool
}

// Connect performs initialization for communication with Ethereum blockchain
// based on provided config.
func Connect(ctx context.Context) TestingChain {
	blockCounter, err := local.BlockCounter()
	if err != nil {
		panic(err) // should never happen
	}

	operatorKey, err := cecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		panic(err)
	}

	signer := commonLocal.NewSigner(operatorKey)

	localChain := &localChain{
		blockCounter:        blockCounter,
		keeps:               make(map[chain.KeepID]*localKeep),
		keepCreatedHandlers: make(map[int]func(event *chain.BondedECDSAKeepCreatedEvent)),
		operatorKey:         operatorKey,
		signer:              signer,
		authorizations:      make(map[chain.OperatorID]bool),
	}

	// block 0 must be stored manually as it is not delivered by the block counter
	localChain.blocksTimestamps.Store(uint64(0), uint64(time.Now().Unix()))

	go localChain.observeBlocksTimestamps(ctx)

	return localChain
}

func (lc *localChain) Name() string {
	return "Local" // include operator address?
}

func (lc *localChain) OperatorID() chain.OperatorID {
	return lc.PublicKeyToOperatorID(&lc.operatorKey.PublicKey)
}

func (lc *localChain) PublicKeyToOperatorID(publicKey *cecdsa.PublicKey) chain.OperatorID {
	return combinedChainID(crypto.PubkeyToAddress(*publicKey))
}

func (lc *localChain) observeBlocksTimestamps(ctx context.Context) {
	blockCounter, _ := lc.BlockCounter() // always errorless in localChain
	blockChan := blockCounter.WatchBlocks(ctx)

	for {
		select {
		case blockNumber := <-blockChan:
			lc.blocksTimestamps.Store(blockNumber, uint64(time.Now().Unix()))
		case <-ctx.Done():
			return
		}
	}
}

func (lc *localChain) Signing() corechain.Signing {
	return commonLocal.NewSigner(lc.operatorKey)
}

func (lc *localChain) OpenKeep(
	keepAddress common.Address,
	members []common.Address,
) TestingBondedECDSAKeepHandle {
	keep, err := lc.createKeepWithMembers(keepAddress, members)
	if err != nil {
		panic(err)
	}

	return keep
}

func (lc *localChain) CloseKeep(keepID chain.KeepID) error {
	keepAddress := common.HexToAddress(keepID.String())
	return lc.closeKeep(keepAddress)
}

func (lc *localChain) TerminateKeep(keepID chain.KeepID) error {
	keepAddress := common.HexToAddress(keepID.String())
	return lc.terminateKeep(keepAddress)
}

func (lc *localChain) AuthorizeOperator(operator common.Address) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	lc.authorizations[combinedChainID(operator)] = true
}

func (lc *localChain) StakeMonitor() (corechain.StakeMonitor, error) {
	return nil, nil // not implemented.
}

func (lc *localChain) BalanceMonitor() (corechain.BalanceMonitor, error) {
	panic("not implemented")
}

func (lc *localChain) BlockCounter() (corechain.BlockCounter, error) {
	return lc.blockCounter, nil
}

func (lc *localChain) BlockTimestamp(blockNumber *big.Int) (uint64, error) {
	blockTimestamp, ok := lc.blocksTimestamps.Load(blockNumber.Uint64())
	if !ok {
		return 0, fmt.Errorf("no timestamp for block [%v]", blockNumber)
	}

	return blockTimestamp.(uint64), nil
}

func generateHandlerID() int {
	// #nosec G404 (insecure random number source (rand))
	// Local chain implementation doesn't require secure randomness.
	return rand.Int()
}

func RandomSigningGroup(size int) []common.Address {
	signers := make([]common.Address, size)

	for i := range signers {
		signers[i] = generateAddress()
	}

	return signers
}

func generateKeepID() chain.KeepID {
	return combinedChainID(generateAddress())
}

func generateAddress() common.Address {
	var address [20]byte
	// #nosec G404 G104 (insecure random number source (rand) | error unhandled)
	// Local chain implementation doesn't require secure randomness.
	// Error can be ignored because according to the `rand.Read` docs it's
	// always `nil`.
	rand.Read(address[:])
	return address
}

// combinedChainID represents all chain ids in one type: KeepMemberID,
// OperatorID, KeepApplicationID, and KeepID. For the Ethereum chain, this is an
// appropriate abstraction since all are underlied by the common.Address type.
type combinedChainID common.Address

func (cci combinedChainID) OperatorID() chain.OperatorID {
	return cci
}

func (cci combinedChainID) KeepMemberID(keepID chain.KeepID) chain.KeepMemberID {
	return cci
}

func (cci combinedChainID) String() string {
	return common.Address(cci).String()
}
