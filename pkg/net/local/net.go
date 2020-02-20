package local

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-core/pkg/net/key"
	brdcLocal "github.com/keep-network/keep-core/pkg/net/local"
	"github.com/keep-network/keep-tecdsa/pkg/net"
)

var logger = log.Logger("keep-net")

type localProvider struct {
	transportID       localIdentifier
	broadcastProvider net.BroadcastProvider
	unicastProvider   *unicastProvider
}

// LocalProvider returns local implementation of net.Provider which can be used
// for testing.
func LocalProvider(
	staticKey *key.NetworkPublic, // node's public key
) net.Provider {
	return &localProvider{
		broadcastProvider: brdcLocal.ConnectWithKey(staticKey),
		unicastProvider:   unicastConnectWithKey(staticKey),
	}
}

func (p *localProvider) BroadcastChannelFor(name string) (net.BroadcastChannel, error) {
	return p.broadcastProvider.ChannelFor(name)
}

func (p *localProvider) UnicastChannelWith(peerID net.TransportIdentifier) (
	net.UnicastChannel,
	error,
) {
	return p.unicastProvider.UnicastChannelWith(peerID)
}

func (p *localProvider) OnUnicastChannelOpened(
	handler func(channel net.UnicastChannel),
) {
	p.unicastProvider.OnUnicastChannelOpened(context.Background(), handler)
}

func (p *localProvider) SeekTransportIdentifier(address common.Address) (net.TransportIdentifier, error) {
	providersMutex.RLock()
	defer providersMutex.RUnlock()

	addressHex := address.Hex()

	for transportIdentifier, provider := range providers {
		if key.NetworkPubKeyToEthAddress(provider.staticKey) == addressHex {
			return localIdentifier(transportIdentifier), nil
		}
	}

	return nil, fmt.Errorf("transport identifier not found for address: [%v]", addressHex)
}

type localIdentifier string

func createLocalIdentifier(staticKey *key.NetworkPublic) localIdentifier {
	return localIdentifier(hex.EncodeToString(key.Marshal(staticKey)))
}

func (li localIdentifier) String() string {
	return string(li)
}
