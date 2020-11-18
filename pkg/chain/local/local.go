package local

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/keep-network/keep-core/pkg/chain/local"
	"github.com/keep-network/keep-ecdsa/pkg/utils/byteutils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-core/pkg/chain"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
)

// Chain is an extention of eth.Handle interface which exposes
// additional functions useful for testing.
type Chain interface {
	eth.Handle

	OpenKeep(keepAddress common.Address, members []common.Address)
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

	blockCounter     chain.BlockCounter
	blocksTimestamps sync.Map

	keepAddresses []common.Address
	keeps         map[common.Address]*localKeep

	keepCreatedHandlers map[int]func(event *eth.BondedECDSAKeepCreatedEvent)

	clientAddress common.Address

	authorizations map[common.Address]bool
}

// Connect performs initialization for communication with Ethereum blockchain
// based on provided config.
func Connect(ctx context.Context) Chain {
	blockCounter, err := local.BlockCounter()
	if err != nil {
		panic(err) // should never happen
	}

	localChain := &localChain{
		blockCounter:        blockCounter,
		keeps:               make(map[common.Address]*localKeep),
		keepCreatedHandlers: make(map[int]func(event *eth.BondedECDSAKeepCreatedEvent)),
		clientAddress:       common.HexToAddress("6299496199d99941193Fdd2d717ef585F431eA05"),
		authorizations:      make(map[common.Address]bool),
	}

	// block 0 must be stored manually as it is not delivered by the block counter
	localChain.blocksTimestamps.Store(uint64(0), uint64(time.Now().Unix()))

	go localChain.observeBlocksTimestamps(ctx)

	return localChain
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

func (lc *localChain) OpenKeep(keepAddress common.Address, members []common.Address) {
	err := lc.createKeepWithMembers(keepAddress, members)
	if err != nil {
		panic(err)
	}
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

// Address returns client's ethereum address.
func (lc *localChain) Address() common.Address {
	return lc.clientAddress
}

func (lc *localChain) StakeMonitor() (chain.StakeMonitor, error) {
	return nil, nil // not implemented.
}

func (lc *localChain) BalanceMonitor() (chain.BalanceMonitor, error) {
	panic("not implemented")
}

// RegisterAsMemberCandidate registers client as a candidate to be selected
// to a keep.
func (lc *localChain) RegisterAsMemberCandidate(application common.Address) error {
	return nil
}

// OnBondedECDSAKeepCreated is a callback that is invoked when an on-chain
// notification of a new ECDSA keep creation is seen.
func (lc *localChain) OnBondedECDSAKeepCreated(
	handler func(event *eth.BondedECDSAKeepCreatedEvent),
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

// OnSignatureRequested is a callback that is invoked on-chain
// when a keep's signature is requested.
func (lc *localChain) OnSignatureRequested(
	keepAddress common.Address,
	handler func(event *eth.SignatureRequestedEvent),
) (subscription.EventSubscription, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	handlerID := generateHandlerID()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return nil, fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	keep.signatureRequestedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lc.localChainMutex.Lock()
		defer lc.localChainMutex.Unlock()

		delete(keep.signatureRequestedHandlers, handlerID)
	}), nil
}

// SubmitKeepPublicKey checks if public key has been already submitted for given
// keep address, if not it stores the key in a map.
func (lc *localChain) SubmitKeepPublicKey(
	keepAddress common.Address,
	publicKey [64]byte,
) error {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	if keep.publicKey != [64]byte{} {
		return fmt.Errorf(
			"public key already submitted for keep [%s]",
			keepAddress.String(),
		)
	}

	keep.publicKey = publicKey

	return nil
}

// SubmitSignature submits a signature to a keep contract deployed under a
// given address.
func (lc *localChain) SubmitSignature(
	keepAddress common.Address,
	signature *ecdsa.Signature,
) error {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	// force the right workflow sequence
	if keep.latestDigest == [32]byte{} {
		return fmt.Errorf(
			"keep [%s] is not awaiting for a signature",
			keepAddress.String(),
		)
	}

	rBytes, err := byteutils.BytesTo32Byte(signature.R.Bytes())
	if err != nil {
		return err
	}

	sBytes, err := byteutils.BytesTo32Byte(signature.S.Bytes())
	if err != nil {
		return err
	}

	keep.signatureSubmittedEvents = append(
		keep.signatureSubmittedEvents,
		&eth.SignatureSubmittedEvent{
			Digest:     keep.latestDigest,
			R:          rBytes,
			S:          sBytes,
			RecoveryID: uint8(signature.RecoveryID),
		},
	)

	return nil
}

// IsAwaitingSignature checks if the keep is waiting for a signature to be
// calculated for the given digest.
func (lc *localChain) IsAwaitingSignature(
	keepAddress common.Address,
	digest [32]byte,
) (bool, error) {
	panic("implement")
}

// IsActive checks for current state of a keep on-chain.
func (lc *localChain) IsActive(keepAddress common.Address) (bool, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return false, fmt.Errorf("no keep with address [%v]", keepAddress)
	}

	return keep.status == active, nil
}

func (lc *localChain) BlockCounter() chain.BlockCounter {
	return lc.blockCounter
}

func (lc *localChain) IsRegisteredForApplication(application common.Address) (bool, error) {
	panic("implement")
}

func (lc *localChain) IsEligibleForApplication(application common.Address) (bool, error) {
	panic("implement")
}

func (lc *localChain) IsStatusUpToDateForApplication(application common.Address) (bool, error) {
	panic("implement")
}

func (lc *localChain) UpdateStatusForApplication(application common.Address) error {
	panic("implement")
}

func (lc *localChain) IsOperatorAuthorized(operator common.Address) (bool, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	return lc.authorizations[operator], nil
}

func (lc *localChain) GetKeepCount() (*big.Int, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	return big.NewInt(int64(len(lc.keeps))), nil
}

func (lc *localChain) GetKeepAtIndex(
	keepIndex *big.Int,
) (common.Address, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	index := int(keepIndex.Uint64())

	if index > len(lc.keepAddresses) {
		return common.HexToAddress("0x0"), fmt.Errorf("out of bounds")
	}

	return lc.keepAddresses[index], nil
}

func (lc *localChain) OnKeepClosed(
	keepAddress common.Address,
	handler func(event *eth.KeepClosedEvent),
) (subscription.EventSubscription, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	handlerID := generateHandlerID()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return nil, fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	keep.keepClosedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lc.localChainMutex.Lock()
		defer lc.localChainMutex.Unlock()

		delete(keep.keepClosedHandlers, handlerID)
	}), nil
}

func (lc *localChain) OnKeepTerminated(
	keepAddress common.Address,
	handler func(event *eth.KeepTerminatedEvent),
) (subscription.EventSubscription, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	handlerID := generateHandlerID()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return nil, fmt.Errorf(
			"failed to find keep with address: [%s]",
			keepAddress.String(),
		)
	}

	keep.keepTerminatedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lc.localChainMutex.Lock()
		defer lc.localChainMutex.Unlock()

		delete(keep.keepTerminatedHandlers, handlerID)
	}), nil
}

func (lc *localChain) OnConflictingPublicKeySubmitted(
	keepAddress common.Address,
	handler func(event *eth.ConflictingPublicKeySubmittedEvent),
) (subscription.EventSubscription, error) {
	panic("implement")
}

func (lc *localChain) OnPublicKeyPublished(
	keepAddress common.Address,
	handler func(event *eth.PublicKeyPublishedEvent),
) (subscription.EventSubscription, error) {
	panic("implement")
}

func (lc *localChain) LatestDigest(keepAddress common.Address) ([32]byte, error) {
	panic("implement")
}

func (lc *localChain) SignatureRequestedBlock(
	keepAddress common.Address,
	digest [32]byte,
) (uint64, error) {
	panic("implement")
}

func (lc *localChain) GetPublicKey(keepAddress common.Address) ([]uint8, error) {
	panic("implement")
}

func (lc *localChain) GetMembers(
	keepAddress common.Address,
) ([]common.Address, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	keep, ok := lc.keeps[keepAddress]
	if !ok {
		return nil, fmt.Errorf("no keep with address [%v]", keepAddress)
	}
	return keep.members, nil
}

func (lc *localChain) GetHonestThreshold(
	keepAddress common.Address,
) (uint64, error) {
	panic("implement")
}

func (lc *localChain) GetOpenedTimestamp(keepAddress common.Address) (time.Time, error) {
	panic("implement")
}

func (lc *localChain) PastSignatureSubmittedEvents(
	keepAddress string,
	startBlock uint64,
) ([]*eth.SignatureSubmittedEvent, error) {
	lc.localChainMutex.Lock()
	defer lc.localChainMutex.Unlock()

	keep, ok := lc.keeps[common.HexToAddress(keepAddress)]
	if !ok {
		return nil, fmt.Errorf("no keep with address [%v]", keepAddress)
	}

	return keep.signatureSubmittedEvents, nil
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
