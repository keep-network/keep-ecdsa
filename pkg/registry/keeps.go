package registry

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

// Keeps represents a collection of keeps in which the given client is a member.
type Keeps struct {
	myKeeps sync.Map // <keepAddress, signer>

	storage storage
}

// NewKeepsRegistry returns an empty keeps registry.
func NewKeepsRegistry(persistence persistence.Handle) *Keeps {
	return &Keeps{
		storage: newStorage(persistence),
	}
}

// RegisterSigner registers that a signer was successfully created for the given
// keep.
func (g *Keeps) RegisterSigner(
	keepAddress common.Address,
	signer *ecdsa.Signer,
) error {
	err := g.storage.save(keepAddress, signer)
	if err != nil {
		return fmt.Errorf("could not persist signer to the storage: [%v]", err)
	}

	g.myKeeps.Store(keepAddress.String(), signer)

	return nil
}

// GetSigner gets a signer by a keep address.
func (g *Keeps) GetSigner(keepAddress common.Address) (*ecdsa.Signer, error) {
	signer, ok := g.myKeeps.Load(keepAddress.String())
	if !ok {
		return nil, fmt.Errorf("could not find signer: [%s]", keepAddress.String())
	}
	return signer.(*ecdsa.Signer), nil
}

// GetKeepsAddresses returns addresses of all registered keeps.
func (g *Keeps) GetKeepsAddresses() []common.Address {
	keepsAddresses := make([]common.Address, 0)

	g.myKeeps.Range(func(key, value interface{}) bool {
		keepAddress := common.HexToAddress(key.(string))
		keepsAddresses = append(keepsAddresses, keepAddress)
		return true
	})

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
			g.myKeeps.Store(keepSigner.keepAddress.String(), keepSigner.signer)
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

func (g *Keeps) printSigners() {
	g.myKeeps.Range(func(key, value interface{}) bool {
		logger.Infof(
			"signer for keep [%s] was loaded with public key: [x: [%x], y: [%x]]",
			key,
			value.(*ecdsa.Signer).PublicKey().X,
			value.(*ecdsa.Signer).PublicKey().Y,
		)
		return true
	})
}
