package local

import (
	"fmt"
	"sync"

	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-tecdsa/pkg/net"
	"github.com/keep-network/keep-tecdsa/pkg/net/internal"
)

var providersMutex = &sync.Mutex{}
var providers = &sync.Map{} // < transportID, unicastProvider >

type unicastProvider struct {
	publicKey     *key.NetworkPublic
	transportID   localIdentifier
	channelsMutex *sync.Mutex
	channels      *sync.Map // < transportID, net.UnicastChannel>
	errChan       chan error
}

func unicastConnectWithKey(
	publicKey *key.NetworkPublic,
	errChan chan error,
) *unicastProvider {
	providersMutex.Lock()
	defer providersMutex.Unlock()

	transportID := key.NetworkPubKeyToEthAddress(publicKey)

	existingProvider, ok := providers.Load(transportID)
	if ok {
		return existingProvider.(*unicastProvider)
	}

	provider := &unicastProvider{
		publicKey:     publicKey,
		transportID:   localIdentifier(transportID),
		channels:      &sync.Map{},
		channelsMutex: &sync.Mutex{},
		errChan:       errChan,
	}

	providers.Store(transportID, provider)

	return provider
}

func (up *unicastProvider) ChannelFor(peer string) (net.UnicastChannel, error) {
	return up.channel(peer)
}

func (up *unicastProvider) channel(peerID string) (net.UnicastChannel, error) {
	up.channelsMutex.Lock()
	defer up.channelsMutex.Unlock()

	existingChannel, ok := up.channels.Load(peerID)
	if ok {
		return existingChannel.(*unicastChannel), nil
	}

	channel := &unicastChannel{
		transportID:          up.transportID,
		peerID:               localIdentifier(peerID),
		senderPublicKey:      up.publicKey,
		messageHandlersMutex: sync.Mutex{},
		errChan:              up.errChan,
	}

	up.channels.Store(peerID, channel)

	return channel, nil
}

type unicastChannel struct {
	transportID localIdentifier
	peerID      localIdentifier

	senderPublicKey *key.NetworkPublic

	messageHandlersMutex sync.Mutex
	messageHandlers      sync.Map // <string, []net.HandleMessageFunc>
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

	if value, exists := uc.messageHandlers.Load(handler.Type); exists {
		handlers = value.([]*net.HandleMessageFunc)
	}

	handlers = append(handlers, &handler)

	uc.messageHandlers.Store(handler.Type, handlers)

	return
}

func (uc *unicastChannel) UnregisterRecv(handlerType string) error {
	uc.messageHandlersMutex.Lock()
	defer uc.messageHandlersMutex.Unlock()

	uc.messageHandlers.Delete(handlerType)

	return nil
}

func (uc *unicastChannel) Send(message net.TaggedMarshaler) error {
	return uc.doSend(message)
}

func (uc *unicastChannel) doSend(payload net.TaggedMarshaler) error {
	value, found := providers.Load(uc.peerID.String())
	if !found {
		return fmt.Errorf("failed to find provider for: [%v]", uc.peerID)
	}
	provider := value.(*unicastProvider)

	bytes, err := payload.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal payload: [%v]", err)
	}

	provider.channels.Range(func(key, value interface{}) bool {
		targetChannel := value.(*unicastChannel)

		if targetChannel.transportID.String() == key.(string) {
			return true // don't send to self
		}

		unmarshaler, found := targetChannel.unmarshalersByType.Load(payload.Type())
		if !found {
			provider.errChan <- fmt.Errorf(
				"couldn't find unmarshaler for type [%s] in unicast channel",
				payload.Type(),
			)

			return true
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

	uc.messageHandlers.Range(func(key, value interface{}) bool {
		handlers := value.([]*net.HandleMessageFunc)

		for _, handler := range handlers {
			go func(handler *net.HandleMessageFunc) {
				err := handler.Handler(message)
				if err != nil {
					errChan <- fmt.Errorf("failed to handle message: [%v]", err)
				}
			}(handler)
		}

		return true
	})

	return nil
}
