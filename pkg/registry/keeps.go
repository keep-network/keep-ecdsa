package registry

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss"
)

var logger = log.Logger("keep-registry")

// Keeps represents a collection of keeps in which the given client is a member.
type Keeps struct {
	myKeepsMutex *sync.RWMutex
	myKeeps      map[common.Address]*tss.ThresholdSigner

	storage storage
}

// NewKeepsRegistry returns an empty keeps registry.
func NewKeepsRegistry(persistence persistence.Handle) *Keeps {
	return &Keeps{
		myKeepsMutex: &sync.RWMutex{},
		myKeeps:      make(map[common.Address]*tss.ThresholdSigner),
		storage:      newStorage(persistence),
	}
}

// RegisterSigner registers that a signer was successfully created for the given
// keep.
func (k *Keeps) RegisterSigner(
	keepAddress common.Address,
	signer *tss.ThresholdSigner,
) error {
	k.myKeepsMutex.Lock()
	defer k.myKeepsMutex.Unlock()

	if _, exists := k.myKeeps[keepAddress]; exists {
		return fmt.Errorf(
			"signer for keep [%s] already registered",
			keepAddress.String(),
		)
	}

	err := k.storage.save(keepAddress, signer)
	if err != nil {
		return fmt.Errorf(
			"could not persist signer for keep [%s] in the storage: [%v]",
			keepAddress.String(),
			err,
		)
	}

	k.myKeeps[keepAddress] = signer

	return nil
}

func (k *Keeps) SnapshotSigner(
	keepAddress common.Address,
	signer *tss.ThresholdSigner,
) error {
	return k.storage.snapshot(keepAddress, signer)
}

// UnregisterKeep archives threeshold signer info for the given keep address.
func (k *Keeps) UnregisterKeep(keepAddress common.Address) {
	k.myKeepsMutex.Lock()
	defer k.myKeepsMutex.Unlock()

	err := k.storage.archive(keepAddress.String())
	if err != nil {
		logger.Errorf("could not archive keep to the storage: [%v]", err)
	}

	delete(k.myKeeps, keepAddress)
}

// GetSigner gets signer for a keep address.
func (k *Keeps) GetSigner(keepAddress common.Address) (*tss.ThresholdSigner, error) {
	k.myKeepsMutex.RLock()
	defer k.myKeepsMutex.RUnlock()

	signer, ok := k.myKeeps[keepAddress]
	if !ok {
		return nil, fmt.Errorf(
			"could not find signer for keep: [%s]",
			keepAddress.String(),
		)
	}

	return signer, nil
}

// HasSigner returns true if at least one signer exists in the registry
// for the keep with the given addres.
func (k *Keeps) HasSigner(keepAddress common.Address) bool {
	k.myKeepsMutex.RLock()
	defer k.myKeepsMutex.RUnlock()

	_, has := k.myKeeps[keepAddress]
	return has
}

// GetKeepsAddresses returns addresses of all registered keeps.
func (k *Keeps) GetKeepsAddresses() []common.Address {
	k.myKeepsMutex.RLock()
	defer k.myKeepsMutex.RUnlock()

	keepsAddresses := make([]common.Address, 0)

	for keepAddress := range k.myKeeps {
		keepsAddresses = append(keepsAddresses, keepAddress)
	}

	return keepsAddresses
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
			if _, exists := k.myKeeps[keepSigner.keepAddress]; exists {
				logger.Errorf(
					"signer for keep [%s] already loaded; "+
						"possible duplicate in the storage layer",
					keepSigner.keepAddress.String(),
				)
				continue
			}

			k.myKeeps[keepSigner.keepAddress] = keepSigner.signer
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
