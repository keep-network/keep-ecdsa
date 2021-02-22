//+build celo

package celo

import (
	celocommon "github.com/celo-org/celo-blockchain/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

// TODO: Tools from this file are temporary. Celo chain implementation uses
//  them to fit into the chain interface which still depends on `go-ethereum`.
//  Once that problem is addressed, tools from this file will be no longer
//  needed. To make future refactoring easier, this is the only file within
//  the whole `celo` package which can import the `go-ethereum` module.

type ExternalAddress = ethcommon.Address
type InternalAddress = celocommon.Address

func toExternalAddress(address InternalAddress) ExternalAddress {
	return ethcommon.BytesToAddress(address.Bytes())
}

func toExternalAddresses(addresses []InternalAddress) []ExternalAddress {
	result := make([]ethcommon.Address, len(addresses))

	for i, address := range addresses {
		result[i] = toExternalAddress(address)
	}

	return result
}

func fromExternalAddress(address ExternalAddress) InternalAddress {
	return celocommon.BytesToAddress(address.Bytes())
}
