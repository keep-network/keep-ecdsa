package local

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-core/pkg/net/key"
	brdcLocal "github.com/keep-network/keep-core/pkg/net/local"
	"github.com/keep-network/keep-tecdsa/pkg/net"
)

var logger = log.Logger("keep-net")

type localProvider struct {
	transportID       net.TransportIdentifier
	broadcastProvider net.BroadcastProvider
	unicastProvider   *unicastProvider
}

// LocalProvider returns local implementation of net.Provider which can be used
// for testing.
func LocalProvider(
	publicKey *key.NetworkPublic, // node's public key
) net.Provider {
	return &localProvider{
		transportID:       localIdentifierFromNetworkKey(publicKey),
		broadcastProvider: brdcLocal.ConnectWithKey(publicKey),
		unicastProvider:   unicastConnectWithKey(publicKey),
	}
}

func (p *localProvider) ID() net.TransportIdentifier {
	return p.transportID
}

func (p *localProvider) TransportIDForPublicKey(publicKey []byte) (net.TransportIdentifier, error) {
	ecdsaPublicKey, err := crypto.UnmarshalPubkey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal public key")
	}
	ethereumAddress := crypto.PubkeyToAddress(*ecdsaPublicKey)

	networkID := hex.EncodeToString(ethereumAddress.Bytes())
	return localIdentifier(networkID), nil
}

func (p *localProvider) BroadcastChannelFor(name string) (net.BroadcastChannel, error) {
	return p.broadcastProvider.ChannelFor(name)
}

func (p *localProvider) UnicastChannelWith(peer net.TransportIdentifier) (net.UnicastChannel, error) {
	return p.unicastProvider.ChannelFor(peer)
}

type localIdentifier string

func (li localIdentifier) String() string {
	return string(li)
}

func localIdentifierFromNetworkKey(publicKey *key.NetworkPublic) localIdentifier {
	ethereumAddress := key.NetworkPubKeyToEthAddress(publicKey)
	return localIdentifier(hex.EncodeToString(common.FromHex(ethereumAddress)))
}
