package local

import (
	"context"
	"encoding/hex"
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
	networkPublicKey *key.NetworkPublic, // node's public key
) net.Provider {
	publicKey, _ := hex.DecodeString(key.NetworkPubKeyToEthAddress(networkPublicKey)[2:])

	return &localProvider{
		broadcastProvider: brdcLocal.ConnectWithKey(networkPublicKey),
		unicastProvider:   unicastConnectWithKey(publicKey),
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

func (p *localProvider) CreateTransportIdentifier(publicKey []byte) net.TransportIdentifier {
	return createTransportIdentifier(publicKey)
}

type localIdentifier string

func createTransportIdentifier(publicKey []byte) net.TransportIdentifier {
	return localIdentifier(hex.EncodeToString(publicKey))
}

func (li localIdentifier) String() string {
	return string(li)
}
