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
	"github.com/keep-network/keep-common/pkg/subscription"
	corechain "github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain"

	commonLocal "github.com/keep-network/keep-common/pkg/chain/local"
)

// Chain is an extention of eth.Handle interface which exposes
// additional functions useful for testing.
type Chain interface {
	chain.Handle

	OperatorAddress() common.Address

	OpenKeep(keepAddress common.Address, members []common.Address) chain.BondedECDSAKeepHandle
	CloseKeep(keepAddress common.Address) error
	TerminateKeep(keepAddress common.Address) error
	RequestSignature(keepAddress common.Address, digest [32]byte) error
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

	authorizations map[common.Address]bool
}

// Connect performs initialization for the local chain, wrapped in the provided
// context.
func Connect(ctx context.Context) Chain {
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
		authorizations:      make(map[common.Address]bool),
	}

	// block 0 must be stored manually as it is not delivered by the block counter
	localChain.blocksTimestamps.Store(uint64(0), uint64(time.Now().Unix()))

	go localChain.observeBlocksTimestamps(ctx)

	return localChain
}

func (lc *localChain) Name() string {
	return "local"
}

func (lc *localChain) observeBlocksTimestamps(ctx context.Context) {
	blockChan := lc.BlockCounter().WatchBlocks(ctx)

	for {
		select {
		case blockNumber := <-blockChan:
			lc.blocksTimestamps.Store(blockNumber, uint64(time.Now().Unix()))
		case <-ctx.Done():
			return
		}
	}
}

func (lc *localChain) OperatorAddress() common.Address {
	return common.BytesToAddress(lc.signer.PublicKey())
}

func (lc *localChain) OperatorID() chain.ID {
	return localChainID(common.BytesToAddress(lc.signer.PublicKey()))
}

func (lc *localChain) Signing() corechain.Signing {
	return commonLocal.NewSigner(lc.operatorKey)
}

func (lc *localChain) OpenKeep(keepAddress common.Address, members []common.Address) chain.BondedECDSAKeepHandle {
	err := lc.createKeepWithMembers(keepAddress, members)
	if err != nil {
		panic(err)
	}

	// GetKeepWithID never errors in localChain.
	keep, _ := lc.GetKeepWithID(localChainID(keepAddress))
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

// OnBondedECDSAKeepCreated is a callback that is invoked when an on-chain
// notification of a new ECDSA keep creation is seen.
func (lc *localChain) OnBondedECDSAKeepCreated(
	handler func(event *chain.BondedECDSAKeepCreatedEvent),
) subscription.EventSubscription {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	handlerID := generateHandlerID()

	lc.keepCreatedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lc.localChainMutex.Lock()
		defer lc.localChainMutex.Unlock()

		delete(lc.keepCreatedHandlers, handlerID)
	})
}

func (lc *localChain) BlockCounter() corechain.BlockCounter {
	return lc.blockCounter
}

func (lc *localChain) IsOperatorAuthorized(operator chain.ID) (bool, error) {
	operatorAddress, err := fromChainID(operator)
	if err != nil {
		return false, err
	}

	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	return lc.authorizations[operatorAddress], nil
}

func (lc *localChain) GetKeepCount() (*big.Int, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	return big.NewInt(int64(len(lc.keeps))), nil
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

// RandomSigningGroup randmly chooses `size` signers to be a new signing group
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

func (lc *localChain) closeKeep(keepAddress common.Address) error {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	if keep.status != active {
		return fmt.Errorf("only active keeps can be closed")
	}

	keep.status = closed

	keepClosedEvent := &chain.KeepClosedEvent{}

	for _, handler := range keep.keepClosedHandlers {
		go func(
			handler func(event *chain.KeepClosedEvent),
			keepClosedEvent *chain.KeepClosedEvent,
		) {
			handler(keepClosedEvent)
		}(handler, keepClosedEvent)
	}

	return nil
}

func (lc *localChain) terminateKeep(keepAddress common.Address) error {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	if keep.status != active {
		return fmt.Errorf("only active keeps can be terminated")
	}

	keep.status = terminated

	keepTerminatedEvent := &chain.KeepTerminatedEvent{}

	for _, handler := range keep.keepTerminatedHandlers {
		go func(
			handler func(event *chain.KeepTerminatedEvent),
			keepTerminatedEvent *chain.KeepTerminatedEvent,
		) {
			handler(keepTerminatedEvent)
		}(handler, keepTerminatedEvent)
	}

	return nil
}
