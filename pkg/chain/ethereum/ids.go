//+build !celo

package ethereum

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

func (ec ethereumChain) UnmarshalID(idString string) (chain.ID, error) {
	if !common.IsHexAddress(idString) {
		return nil, fmt.Errorf(
			"[%v] is not a valid ethereum ID",
			idString,
		)
	}

	return ethereumChainID(common.HexToAddress(idString)), nil
}

// ethereumChainID is the Ethereum-speecific chain.ID type; it is an alias for
// go-ethereum/common.Address.
type ethereumChainID common.Address

func (eci ethereumChainID) ChainName() string {
	return "ethereum"
}

func (eci ethereumChainID) String() string {
	return common.Address(eci).Hex()
}

func (eci ethereumChainID) IsForChain(handle chain.Handle) bool {
	_, ok := handle.(*ethereumChain)

	return ok
}

func toIDSlice(addresses []common.Address) []chain.ID {
	memberIDs := make([]chain.ID, 0, len(addresses))
	for _, address := range addresses {
		memberIDs = append(memberIDs, ethereumChainID(address))
	}

	return memberIDs
}

func fromChainID(id chain.ID) (common.Address, error) {
	eci, ok := id.(ethereumChainID)
	if !ok {
		return common.Address{}, fmt.Errorf("failed to convert to ethereumChainID")
	}

	return common.Address(eci), nil
}
