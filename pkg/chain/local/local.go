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

	"github.com/keep-network/keep-core/pkg/chain/local"

	"github.com/ethereum/go-ethereum/common"
	corechain "github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain"

	commonLocal "github.com/keep-network/keep-common/pkg/chain/local"
)

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

	OpenKeep(keepAddress common.Address, members []common.Address) TestingBondedECDSAKeepHandle
	CloseKeep(keepAddress common.Address) error
	TerminateKeep(keepAddress common.Address) error
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

	keepAddresses []common.Address
	keeps         map[common.Address]*localKeep

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
		keeps:               make(map[common.Address]*localKeep),
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

func (lc *localChain) CloseKeep(keepAddress common.Address) error {
	return lc.closeKeep(keepAddress)
}

func (lc *localChain) TerminateKeep(keepAddress common.Address) error {
	return lc.terminateKeep(keepAddress)
}

func (lc *localChain) AuthorizeOperator(operator common.Address) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	lc.authorizations[operator] = true
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

func generateAddress() common.Address {
	var address [20]byte
	// #nosec G404 G104 (insecure random number source (rand) | error unhandled)
	// Local chain implementation doesn't require secure randomness.
	// Error can be ignored because according to the `rand.Read` docs it's
	// always `nil`.
	rand.Read(address[:])
	return address
}
