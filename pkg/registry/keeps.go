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
	myKeeps sync.Map // <keepAddress, membership>

	storage storage
}

// Membership represents a member of a group.
type Membership struct {
	KeepAddress common.Address
	Signer      *ecdsa.Signer
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
	membership := &Membership{
		KeepAddress: keepAddress,
		Signer:      signer,
	}

	err := g.storage.save(membership)
	if err != nil {
		return fmt.Errorf("could not persist membership to the storage: [%v]", err)
	}

	g.myKeeps.Store(keepAddress.String(), membership)

	return nil
}

// GetMembership gets a membership by a keep address.
func (g *Keeps) GetMembership(keepAddress common.Address) (*Membership, error) {
	membership, ok := g.myKeeps.Load(keepAddress.String())
	if !ok {
		return nil, fmt.Errorf("failed to find signer for keep: [%s]", keepAddress.String())
	}
	return membership.(*Membership), nil
}

// ForEachKeep calls function sequentially for each keep address and its'
// membership present in the keeps map. If the function returns false, range
// stops the iteration.
func (g *Keeps) ForEachKeep(
	function func(keepAddress common.Address, membership *Membership) bool,
) {
	g.myKeeps.Range(func(key, value interface{}) bool {
		keepAddress := common.HexToAddress(key.(string))
		return function(keepAddress, value.(*Membership))
	})
}

// LoadExistingKeeps iterates over all stored memberships on disk and loads them
// into memory
func (g *Keeps) LoadExistingKeeps() {
	membershipsChannel, errorsChannel := g.storage.readAll()

	// Two goroutines read from memberships and errors channels and either
	// adds memberships to the keeps registry or outputs an error to stderr.
	// The reason for using two goroutines at the same time - one for
	// memberships and one for errors is because channels do not have to be
	// buffered and we do not know in what order information is written to
	// channels.
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for membership := range membershipsChannel {
			g.myKeeps.Store(membership.KeepAddress.String(), membership)
		}

		wg.Done()
	}()

	go func() {
		for err := range errorsChannel {
			logger.Errorf(
				"could not load membership from disk: [%v]",
				err,
			)
		}

		wg.Done()
	}()

	wg.Wait()

	g.printMemberships()
}

func (g *Keeps) printMemberships() {
	g.ForEachKeep(func(keepAddress common.Address, membership *Membership) bool {
		logger.Infof(
			"membership for keep [%s] was loaded with member public key: [x: [%x], y: [%x]]",
			keepAddress.String(),
			membership.Signer.PublicKey().X,
			membership.Signer.PublicKey().Y,
		)
		return true
	})
}
