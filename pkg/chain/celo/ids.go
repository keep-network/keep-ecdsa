//+build celo

package celo

import (
	cecdsa "crypto/ecdsa"
	"fmt"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/crypto"

	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

func (cc *celoChain) UnmarshalID(idString string) (chain.ID, error) {
	if !common.IsHexAddress(idString) {
		return nil, fmt.Errorf(
			"[%v] is not a valid celo ID",
			idString,
		)
	}

	return celoChainID(common.HexToAddress(idString)), nil
}

func (cc *celoChain) PublicKeyToOperatorID(publicKey *cecdsa.PublicKey) chain.ID {
	return celoChainID(crypto.PubkeyToAddress(*publicKey))
}

// celoChainID is the local chain.ID type; it is an alias for
// go-ethereum/common.Address.
type celoChainID common.Address

func (ci celoChainID) ChainName() string {
	return "celo"
}

func (ci celoChainID) String() string {
	return common.Address(ci).Hex()
}

func (ci celoChainID) IsForChain(handle chain.Handle) bool {
	_, ok := handle.(*celoChain)

	return ok
}

func toIDSlice(addresses []common.Address) []chain.ID {
	memberIDs := make([]chain.ID, 0, len(addresses))
	for _, address := range addresses {
		memberIDs = append(memberIDs, celoChainID(address))
	}

	return memberIDs
}

func fromChainID(id chain.ID) (common.Address, error) {
	ci, ok := id.(celoChainID)
	if !ok {
		return common.Address{}, fmt.Errorf("failed to convert to celoChainID")
	}

	return common.Address(ci), nil
}
