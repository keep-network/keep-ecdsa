package local

import (
	cecdsa "crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-core/pkg/operator"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

func (lc *localChain) UnmarshalID(idString string) (chain.ID, error) {
	if !common.IsHexAddress(idString) {
		return nil, fmt.Errorf(
			"[%v] is not a valid ethereum ID",
			idString,
		)
	}

	return localChainID(common.HexToAddress(idString)), nil
}

func (lc *localChain) PublicKeyToOperatorID(publicKey *cecdsa.PublicKey) chain.ID {
	return localChainID(operator.PubkeyToAddress(*publicKey))
}

// localChainID is the local chain.ID type; it is an alias for
// go-ethereum/common.Address.
type localChainID common.Address

func (ci localChainID) ChainName() string {
	return "local"
}

func (ci localChainID) String() string {
	return common.Address(ci).Hex()
}

func (ci localChainID) IsForChain(handle chain.Handle) bool {
	_, ok := handle.(*localChain)

	return ok
}

func toIDSlice(addresses []common.Address) []chain.ID {
	memberIDs := make([]chain.ID, 0, len(addresses))
	for _, address := range addresses {
		memberIDs = append(memberIDs, localChainID(address))
	}

	return memberIDs
}

func fromChainID(id chain.ID) (common.Address, error) {
	ci, ok := id.(localChainID)
	if !ok {
		return common.Address{}, fmt.Errorf("failed to convert to localChainID")
	}

	return common.Address(ci), nil
}
