package local

import (
	"fmt"
	"sync"

	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-tecdsa/pkg/net"
	"github.com/keep-network/keep-tecdsa/pkg/net/internal"
)

var providersMutex = &sync.RWMutex{}
var providers = make(map[localIdentifier]*unicastProvider)

type unicastProvider struct {
	publicKey     *key.NetworkPublic
	transportID   localIdentifier
	channelsMutex *sync.RWMutex
	channels      map[localIdentifier]*unicastChannel
	errChan       chan error
}

func unicastConnectWithKey(
	publicKey *key.NetworkPublic,
	errChan chan error,
) *unicastProvider {
	providersMutex.Lock()
	defer providersMutex.Unlock()

	transportID := key.NetworkPubKeyToEthAddress(publicKey)

	existingProvider, ok := providers[localIdentifier(transportID)]
	if ok {
		return existingProvider
	}

	provider := &unicastProvider{
		publicKey:     publicKey,
		transportID:   localIdentifier(transportID),
		channels:      make(map[localIdentifier]*unicastChannel),
		channelsMutex: &sync.RWMutex{},
		errChan:       errChan,
	}

	providers[localIdentifier(transportID)] = provider

	return provider
}

func (up *unicastProvider) ChannelFor(peer string) (net.UnicastChannel, error) {
	return up.channel(peer)
}

func (up *unicastProvider) channel(peerID string) (net.UnicastChannel, error) {
	up.channelsMutex.Lock()
	defer up.channelsMutex.Unlock()

	existingChannel, ok := up.channels[localIdentifier(peerID)]
	if ok {
		return existingChannel, nil
	}

	channel := &unicastChannel{
		transportID:          up.transportID,
		peerID:               localIdentifier(peerID),
		senderPublicKey:      up.publicKey,
		messageHandlersMutex: &sync.RWMutex{},
		messageHandlers:      make(map[string][]*net.HandleMessageFunc),
		errChan:              up.errChan,
	}

	up.channels[localIdentifier(peerID)] = channel

	return channel, nil
}

type unicastChannel struct {
	transportID localIdentifier
	peerID      localIdentifier

	senderPublicKey *key.NetworkPublic

	messageHandlersMutex *sync.RWMutex
	messageHandlers      map[string][]*net.HandleMessageFunc
	unmarshalersByType   sync.Map // <string, net.TaggedUnmarshaler>

	errChan chan error
}

func (uc *unicastChannel) RegisterUnmarshaler(
	unmarshaler func() net.TaggedUnmarshaler,
) (err error) {
	_, exists := uc.unmarshalersByType.Load(unmarshaler().Type())
	if exists {
		// TODO: Disabled temporarily because it causes failures on protocol run.
		// We need to confirm how the local unicast channel should be implemented.
		// err = fmt.Errorf("unmarshaler already registered for type: [%v]", unmarshaler().Type())
	} else {
		uc.unmarshalersByType.Store(unmarshaler().Type(), unmarshaler)
	}

	return
}

func (uc *unicastChannel) Recv(handler net.HandleMessageFunc) (err error) {
	uc.messageHandlersMutex.Lock()
	defer uc.messageHandlersMutex.Unlock()

	var handlers []*net.HandleMessageFunc

	if value, exists := uc.messageHandlers[handler.Type]; exists {
		handlers = value
	}

	handlers = append(handlers, &handler)

	uc.messageHandlers[handler.Type] = handlers

	return
}

func (uc *unicastChannel) UnregisterRecv(handlerType string) error {
	uc.messageHandlersMutex.Lock()
	defer uc.messageHandlersMutex.Unlock()

	uc.messageHandlers[handlerType] = nil

	return nil
}

func (uc *unicastChannel) Send(message net.TaggedMarshaler) error {
	return uc.doSend(message)
}

func (uc *unicastChannel) doSend(payload net.TaggedMarshaler) error {
	providersMutex.RLock()
	provider, found := providers[uc.peerID]
	if !found {
		return fmt.Errorf("failed to find provider for: [%v]", uc.peerID)
	}
	providersMutex.RUnlock()

	bytes, err := payload.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal payload: [%v]", err)
	}

	for peerID, targetChannel := range provider.channels {
		if targetChannel.transportID.String() == peerID.String() {
			continue // don't send to self
		}

		unmarshaler, found := targetChannel.unmarshalersByType.Load(payload.Type())
		if !found {
			provider.errChan <- fmt.Errorf(
				"couldn't find unmarshaler for type [%s] in unicast channel",
				payload.Type(),
			)
		}

		unmarshaled := unmarshaler.(func() net.TaggedUnmarshaler)()
		err = unmarshaled.Unmarshal(bytes)
		if err != nil {
			provider.errChan <- err
		}

		if err := targetChannel.deliver(
			uc.transportID,
			uc.senderPublicKey,
			unmarshaled,
			payload.Type(),
			provider.errChan,
		); err != nil {
			provider.errChan <- err
		}

		return true
	})

	return nil
}

func (uc *unicastChannel) deliver(
	senderTransportID localIdentifier,
	senderPublicKey *key.NetworkPublic,
	payload interface{},
	messageType string,
	errChan chan error,
) error {
	message := internal.BasicMessage(
		senderTransportID,
		payload,
		messageType,
		key.Marshal(senderPublicKey),
	)

	uc.messageHandlersMutex.RLock()
	for _, handlers := range uc.messageHandlers {
		for _, handler := range handlers {
			go func(handler *net.HandleMessageFunc) {
				err := handler.Handler(message)
				if err != nil {
					errChan <- fmt.Errorf("failed to handle message: [%v]", err)
				}
			}(handler)
		}
	}
	uc.messageHandlersMutex.RUnlock()

	return nil
}
