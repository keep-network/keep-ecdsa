package local

import (
	"github.com/keep-network/keep-core/pkg/net/key"
	brdcLocal "github.com/keep-network/keep-core/pkg/net/local"
	"github.com/keep-network/keep-tecdsa/pkg/net"
)

type localProvider struct {
	errChan           chan error
	transportID       localIdentifier
	broadcastProvider net.BroadcastProvider
	unicastProvider   *unicastProvider
}

// LocalProvider returns local implementation of net.Provider which can be used
// for testing.
func LocalProvider(
	publicKey *key.NetworkPublic, // node's public key
	errChan chan error,
) net.Provider {
	return &localProvider{
		errChan:           errChan,
		broadcastProvider: brdcLocal.ConnectWithKey(publicKey),
		unicastProvider:   unicastConnectWithKey(publicKey, errChan),
	}
}

func (p *localProvider) BroadcastChannelFor(name string) (net.BroadcastChannel, error) {
	return p.broadcastProvider.ChannelFor(name)
}

func (p *localProvider) UnicastChannelWith(name string) (net.UnicastChannel, error) {
	return p.unicastProvider.ChannelFor(name)
}

type localIdentifier string

func (li localIdentifier) String() string {
	return string(li)
}
