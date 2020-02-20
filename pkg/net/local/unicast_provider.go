package local

import (
	"context"
	"fmt"
	"github.com/keep-network/keep-core/pkg/net/key"
	"sync"

	"github.com/keep-network/keep-tecdsa/pkg/net"
)

var providersMutex = &sync.RWMutex{}
var providers = make(map[string]*unicastProvider)

type unicastProvider struct {
	structMutex *sync.RWMutex

	transportID net.TransportIdentifier
	staticKey   *key.NetworkPublic

	channelManager          *unicastChannelManager
	onChannelOpenedHandlers []*onChannelOpenedHandler
}

type onChannelOpenedHandler struct {
	ctx      context.Context
	handleFn func(remote net.UnicastChannel)
}

func unicastConnectWithKey(
	staticKey *key.NetworkPublic,
) *unicastProvider {
	providersMutex.Lock()
	defer providersMutex.Unlock()

	transportID := createLocalIdentifier(staticKey)

	existingProvider, ok := providers[transportID.String()]
	if ok {
		return existingProvider
	}

	provider := &unicastProvider{
		structMutex:    &sync.RWMutex{},
		transportID:    transportID,
		staticKey:      staticKey,
		channelManager: newUnicastChannelManager(),
	}

	logger.Warningf("Registering as [%v]", transportID)
	providers[transportID.String()] = provider

	return provider
}

// UnicastChannelWith creates a unicast channel with the given remote peer.
// If peer is not known or connection could not be open, function returns error.
func (up *unicastProvider) UnicastChannelWith(
	peer net.TransportIdentifier,
) (net.UnicastChannel, error) {
	channel := up.createChannelWith(peer, false)

	providersMutex.RLock()
	remote, ok := providers[peer.String()]
	providersMutex.RUnlock()

	if !ok {
		return nil, fmt.Errorf("remote peer not known [%v]", peer)
	}

	remote.createChannelWith(up.transportID, true)

	return channel, nil
}

// OnUnicastChannelOpened registers UnicastChannelHandler that will be called
// for each incoming unicast channel opened by remote peers against this one.
// The handlers is active for the entire lifetime of the provided context.
// When the context is done, handler is never called again.
func (up *unicastProvider) OnUnicastChannelOpened(
	ctx context.Context,
	handler func(remote net.UnicastChannel),
) {
	up.structMutex.Lock()
	defer up.structMutex.Unlock()

	up.onChannelOpenedHandlers = append(
		up.onChannelOpenedHandlers,
		&onChannelOpenedHandler{ctx, handler},
	)
}

func (up *unicastProvider) createChannelWith(
	peer net.TransportIdentifier,
	notify bool,
) net.UnicastChannel {
	channel, ok := up.channelManager.getChannel(peer)
	if ok {
		return channel
	}

	channel = newUnicastChannel(up.transportID, up.staticKey, peer)
	up.channelManager.addChannel(channel)

	if notify {
		up.notifyNewChannel(channel)
	}

	return channel
}

func (up *unicastProvider) notifyNewChannel(channel net.UnicastChannel) {
	// first cleanup
	up.structMutex.Lock()
	defer up.structMutex.Unlock()

	i := 0
	for _, handler := range up.onChannelOpenedHandlers {
		if handler.ctx.Err() == nil {
			// still active, should remain in the slice
			up.onChannelOpenedHandlers[i] = handler
			i++

			// firing handler asynchronously to
			// do not block the loop
			go handler.handleFn(channel)
		}
	}

	// cleaning up those no longer active
	up.onChannelOpenedHandlers = up.onChannelOpenedHandlers[:i]
}

func deliverMessage(
	sender net.TransportIdentifier,
	receiver net.TransportIdentifier,
	messagePayload []byte,
	messageType string,
) error {
	providersMutex.RLock()
	receiverProviver, ok := providers[receiver.String()]
	providersMutex.RUnlock()

	if !ok {
		return fmt.Errorf("peer [%v] not known", receiver)
	}

	channel, ok := receiverProviver.channelManager.getChannel(sender)
	if !ok {
		return fmt.Errorf("peer [%v] could not find channel for [%v]", receiver, sender)
	}

	return channel.receiveMessage(messagePayload, messageType)
}
