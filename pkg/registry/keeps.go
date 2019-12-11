package registry

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss"
)

// Keeps represents a collection of keeps in which the given client is a member.
type Keeps struct {
	myKeepsMutex *sync.RWMutex
	myKeeps      map[eth.KeepAddress][]*tss.ThresholdSigner

	storage storage
}

// NewKeepsRegistry returns an empty keeps registry.
func NewKeepsRegistry(persistence persistence.Handle) *Keeps {
	return &Keeps{
		myKeepsMutex: &sync.RWMutex{},
		myKeeps:      make(map[eth.KeepAddress][]*tss.ThresholdSigner),
		storage:      newStorage(persistence),
	}
}

// RegisterSigner registers that a signer was successfully created for the given
// keep.
func (g *Keeps) RegisterSigner(
	keepAddress common.Address,
	signer *tss.ThresholdSigner,
) error {
	err := g.storage.save(keepAddress, signer)
	if err != nil {
		return fmt.Errorf("could not persist signer to the storage: [%v]", err)
	}

	g.storeSigner(keepAddress, signer)

	return nil
}

// GetSigners gets signers by a keep address.
func (g *Keeps) GetSigners(keepAddress common.Address) ([]*tss.ThresholdSigner, error) {
	g.myKeepsMutex.RLock()
	defer g.myKeepsMutex.RUnlock()

	signers, ok := g.myKeeps[keepAddress]
	if !ok {
		return nil, fmt.Errorf("could not find signers for keep: [%s]", keepAddress.String())
	}
	return signers, nil
}

// GetKeepsAddresses returns addresses of all registered keeps.
func (g *Keeps) GetKeepsAddresses() []common.Address {
	g.myKeepsMutex.RLock()
	defer g.myKeepsMutex.RUnlock()

	keepsAddresses := make([]common.Address, 0)

	for keepAddress := range g.myKeeps {
		keepsAddresses = append(keepsAddresses, keepAddress)
	}

	return keepsAddresses
}

// LoadExistingKeeps iterates over all signers stored on disk and loads them
// into memory
func (g *Keeps) LoadExistingKeeps() {
	keepSignersChannel, errorsChannel := g.storage.readAll()

	// Two goroutines read from signers and errors channels and either adds
	// signers to the keeps registry or outputs an error to stderr.
	// The reason for using two goroutines at the same time - one for signers
	// and one for errors is because channels do not have to be
	// buffered and we do not know in what order information is written to
	// channels.
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for keepSigner := range keepSignersChannel {
			g.storeSigner(keepSigner.keepAddress, keepSigner.signer)
		}

		wg.Done()
	}()

	go func() {
		for err := range errorsChannel {
			logger.Errorf("could not load signer from disk: [%v]", err)
		}

		wg.Done()
	}()

	wg.Wait()

	g.printSigners()
}

// ForEachKeep executes callback function for every entry in keeps' registry.
func (g *Keeps) ForEachKeep(
	callback func(keepAddress common.Address, signer []*tss.ThresholdSigner),
) {
	g.myKeepsMutex.RLock()
	defer g.myKeepsMutex.RUnlock()

	for keepAddress, signers := range g.myKeeps {
		callback(keepAddress, signers)
	}
}

func (g *Keeps) printSigners() {
	g.myKeepsMutex.RLock()
	defer g.myKeepsMutex.RUnlock()

	logger.Infof(
		"loaded [%s] keeps from the local storage",
		len(g.myKeeps),
	)

	for keepAddress, signers := range g.myKeeps {
		logger.Debugf(
			"loaded [%d] signers for keep [%s]",
			len(signers),
			keepAddress,
		)
	}
}

func (g *Keeps) storeSigner(
	keepAddress common.Address,
	signer *tss.ThresholdSigner,
) {
	g.myKeepsMutex.Lock()
	defer g.myKeepsMutex.Unlock()

	signers, exists := g.myKeeps[keepAddress]
	if exists {
		g.myKeeps[keepAddress] = append(signers, signer)
	} else {
		g.myKeeps[keepAddress] = []*tss.ThresholdSigner{signer}
	}
}
