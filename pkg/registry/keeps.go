package registry

import (
	"fmt"
	"sync"

	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss"
)

var logger = log.Logger("keep-registry")

// Keeps represents a collection of keeps in which the given client is a member.
type Keeps struct {
	myKeepsMutex *sync.RWMutex
	// 🤔🤔🤔
	myKeepsByChains map[chain.Handle][]chain.BondedECDSAKeepHandle
	myKeeps         map[chain.KeepID]*tss.ThresholdSigner

	storage storage
}

// NewKeepsRegistry returns an empty keeps registry.
func NewKeepsRegistry(persistence persistence.Handle) *Keeps {
	return &Keeps{
		myKeepsMutex:    &sync.RWMutex{},
		myKeepsByChains: make(map[chain.Handle][]chain.BondedECDSAKeepHandle),
		myKeeps:         make(map[chain.KeepID]*tss.ThresholdSigner),
		storage:         newStorage(persistence),
	}
}

// RegisterSigner registers that a signer was successfully created for the given
// keep.
func (k *Keeps) RegisterSigner(
	keepID chain.KeepID,
	signer *tss.ThresholdSigner,
) error {
	k.myKeepsMutex.Lock()
	defer k.myKeepsMutex.Unlock()

	if _, exists := k.myKeeps[keepID]; exists {
		return fmt.Errorf(
			"signer for keep [%s] already registered",
			keepID.String(),
		)
	}

	err := k.storage.save(keepID, signer)
	if err != nil {
		return fmt.Errorf(
			"could not persist signer for keep [%s] in the storage: [%v]",
			keepID.String(),
			err,
		)
	}

	k.myKeeps[keepID] = signer

	return nil
}

func (k *Keeps) SnapshotSigner(
	keepID chain.KeepID,
	signer *tss.ThresholdSigner,
) error {
	return k.storage.snapshot(keepID, signer)
}

// UnregisterKeep archives threeshold signer info for the given keep address.
func (k *Keeps) UnregisterKeep(keepID chain.KeepID) {
	k.myKeepsMutex.Lock()
	defer k.myKeepsMutex.Unlock()

	err := k.storage.archive(keepID.String())
	if err != nil {
		logger.Errorf("could not archive keep to the storage: [%v]", err)
	}

	delete(k.myKeeps, keepID)
}

// GetSigner gets signer for a keep address.
func (k *Keeps) GetSigner(keepID chain.KeepID) (*tss.ThresholdSigner, error) {
	k.myKeepsMutex.RLock()
	defer k.myKeepsMutex.RUnlock()

	signer, ok := k.myKeeps[keepID]
	if !ok {
		return nil, fmt.Errorf(
			"could not find signer for keep: [%s]",
			keepID.String(),
		)
	}

	return signer, nil
}

// HasSigner returns true if at least one signer exists in the registry
// for the keep with the given addres.
func (k *Keeps) HasSigner(keepID chain.KeepID) bool {
	k.myKeepsMutex.RLock()
	defer k.myKeepsMutex.RUnlock()

	_, has := k.myKeeps[keepID]
	return has
}

// GetKeepsAddresses returns addresses of all registered keeps.
func (k *Keeps) GetKeepsAddresses() []chain.KeepID {
	k.myKeepsMutex.RLock()
	defer k.myKeepsMutex.RUnlock()

	keepIDs := make([]chain.KeepID, 0)

	for keepID := range k.myKeeps {
		keepIDs = append(keepIDs, keepID)
	}

	return keepIDs
}

// LoadExistingKeeps iterates over all signers stored on disk and loads them
// into memory
func (k *Keeps) LoadExistingKeeps() {
	k.myKeepsMutex.Lock()
	defer k.myKeepsMutex.Unlock()

	keepSignersChannel, errorsChannel := k.storage.readAll()

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
			if _, exists := k.myKeeps[keepSigner.keepID]; exists {
				logger.Errorf(
					"signer for keep [%s] already loaded; "+
						"possible duplicate in the storage layer",
					keepSigner.keepID.String(),
				)
				continue
			}

			k.myKeeps[keepSigner.keepID] = keepSigner.signer
		}

		wg.Done()
	}()

	go func() {
		for err := range errorsChannel {
			logger.Errorf("could not load signer from storage: [%v]", err)
		}

		wg.Done()
	}()

	wg.Wait()

	logger.Infof(
		"loaded [%d] keeps from the local storage",
		len(k.myKeeps),
	)

	for keepAddress := range k.myKeeps {
		logger.Debugf(
			"loaded signer for keep [%s]",
			keepAddress.String(),
		)
	}
}
