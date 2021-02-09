package local

import (
	chain "github.com/keep-network/keep-ecdsa/pkg/chain"
)

type tbtcApplicationHandle struct {
	handle *localChain
}

func (lc *localChain) TBTCApplicationHandle() (chain.BondedECDSAKeepApplicationHandle, error) {
	return &tbtcApplicationHandle{handle: lc}, nil
}

func (*tbtcApplicationHandle) IsRegisteredForApplication() (bool, error) {
	panic("implement")
}

func (*tbtcApplicationHandle) IsEligibleForApplication() (bool, error) {
	panic("implement")
}

func (*tbtcApplicationHandle) IsStatusUpToDateForApplication() (bool, error) {
	panic("implement")
}

func (*tbtcApplicationHandle) UpdateStatusForApplication() error {
	panic("implement")
}

// RegisterAsMemberCandidate registers client as a candidate to be selected
// to a keep.
func (tah *tbtcApplicationHandle) RegisterAsMemberCandidate() error {
	return nil
}

func (tah *tbtcApplicationHandle) IsOperatorAuthorized(operator chain.OperatorID) (bool, error) {
	tah.handle.localChainMutex.Lock()
	defer tah.handle.localChainMutex.Unlock()

	return tah.handle.authorizations[operator], nil
}
