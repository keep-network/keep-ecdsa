package registry

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss"
)

// Keeps represents a collection of keeps in which the given client is a member.
type Keeps struct {
	myKeepsMutex *sync.RWMutex
	myKeeps      map[common.Address][]*tss.ThresholdSigner

	storage storage
}

// NewKeepsRegistry returns an empty keeps registry.
func NewKeepsRegistry(persistence persistence.Handle) *Keeps {
	return &Keeps{
		myKeepsMutex: &sync.RWMutex{},
		myKeeps:      make(map[common.Address][]*tss.ThresholdSigner),
		storage:      newStorage(persistence),
	}
}

// RegisterSigner registers that a signer was successfully created for the given
// keep.
func (k *Keeps) RegisterSigner(
	keepAddress common.Address,
	signer *tss.ThresholdSigner,
) error {
	err := k.storage.save(keepAddress, signer)
	if err != nil {
		return fmt.Errorf("could not persist signer to the storage: [%v]", err)
	}

	k.storeSigner(keepAddress, signer)

	return nil
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

// GetSigners gets signers by a keep address.
func (k *Keeps) GetSigners(keepAddress common.Address) ([]*tss.ThresholdSigner, error) {
	k.myKeepsMutex.RLock()
	defer k.myKeepsMutex.RUnlock()

	signers, ok := k.myKeeps[keepAddress]
	if !ok {
		return nil, fmt.Errorf("could not find signers for keep: [%s]", keepAddress.String())
	}
	return signers, nil
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
			k.storeSigner(keepSigner.keepAddress, keepSigner.signer)
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

	k.printSigners()
}

// ForEachKeep executes callback function for every entry in keeps' registry.
func (k *Keeps) ForEachKeep(
	callback func(keepAddress common.Address, signer []*tss.ThresholdSigner),
) {
	k.myKeepsMutex.RLock()
	defer k.myKeepsMutex.RUnlock()

	for keepAddress, signers := range k.myKeeps {
		callback(keepAddress, signers)
	}
}

func (k *Keeps) printSigners() {
	k.myKeepsMutex.RLock()
	defer k.myKeepsMutex.RUnlock()

	logger.Infof(
		"loaded [%d] keeps from the local storage",
		len(k.myKeeps),
	)

	for keepAddress, signers := range k.myKeeps {
		logger.Debugf(
			"loaded [%d] signers for keep [%s]",
			len(signers),
			keepAddress.String(),
		)
	}
}

func (k *Keeps) storeSigner(
	keepAddress common.Address,
	signer *tss.ThresholdSigner,
) {
	k.myKeepsMutex.Lock()
	defer k.myKeepsMutex.Unlock()

	signers, exists := k.myKeeps[keepAddress]
	if exists {
		k.myKeeps[keepAddress] = append(signers, signer)
	} else {
		k.myKeeps[keepAddress] = []*tss.ThresholdSigner{signer}
	}
}
