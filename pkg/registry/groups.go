package registry

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

// Groups represents a collection of keep groups in which the given client is a
// member.
type Groups struct {
	myGroups sync.Map // <keepAddress, membership>

	storage storage
}

// Membership represents a member of a group.
type Membership struct {
	KeepAddress common.Address
	Signer      *ecdsa.Signer
}

// NewGroupRegistry returns an empty GroupRegistry.
func NewGroupRegistry(persistence persistence.Handle) *Groups {
	return &Groups{
		storage: newStorage(persistence),
	}
}

// RegisterGroup registers that a group was successfully created for the given
// keep.
func (g *Groups) RegisterGroup(
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

	g.myGroups.Store(keepAddress.String(), membership)

	return nil
}

// GetGroup gets a group by a keep address.
func (g *Groups) GetGroup(keepAddress common.Address) (*Membership, error) {
	membership, ok := g.myGroups.Load(keepAddress.String())
	if !ok {
		return nil, fmt.Errorf("failed to find signer for keep: [%s]", keepAddress.String())
	}
	return membership.(*Membership), nil
}

// ForEachGroup calls function sequentially for each keep address and its'
// membership present in the groups map. If the function returns false, range
// stops the iteration.
func (g *Groups) ForEachGroup(
	function func(keepAddress common.Address, membership *Membership) bool,
) {
	g.myGroups.Range(func(key, value interface{}) bool {
		keepAddress := common.HexToAddress(key.(string))
		return function(keepAddress, value.(*Membership))
	})
}

// LoadExistingGroups iterates over all stored memberships on disk and loads them
// into memory
func (g *Groups) LoadExistingGroups() {
	membershipsChannel, errorsChannel := g.storage.readAll()

	// Two goroutines read from memberships and errors channels and either
	// adds memberships to the group registry or outputs an error to stderr.
	// The reason for using two goroutines at the same time - one for
	// memberships and one for errors is because channels do not have to be
	// buffered and we do not know in what order information is written to
	// channels.
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for membership := range membershipsChannel {
			g.myGroups.Store(membership.KeepAddress.String(), membership)
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

func (g *Groups) printMemberships() {
	g.ForEachGroup(func(keepAddress common.Address, membership *Membership) bool {
		logger.Infof(
			"membership for keep [%s] was loaded with member public key: [x: [%x], y: [%x]]",
			keepAddress.String(),
			membership.Signer.PublicKey().X,
			membership.Signer.PublicKey().Y,
		)
		return true
	})
}
