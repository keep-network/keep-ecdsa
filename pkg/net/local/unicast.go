package local

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-tecdsa/pkg/net"
	"github.com/keep-network/keep-tecdsa/pkg/net/internal"
)

var providersMutex = &sync.RWMutex{}
var providers = make(map[localIdentifier]*unicastProvider)

var connectionsMutex = &sync.RWMutex{}
var connections = make(map[connectionID]*localConnection)

type connectionID string

type unicastProvider struct {
	publicKey   *key.NetworkPublic
	transportID localIdentifier
}

type localConnection struct {
	connectionID    connectionID
	channels        map[localIdentifier]*unicastChannel
	waitInitialized *sync.WaitGroup
}

type unicastChannel struct {
	transportID localIdentifier
	peerID      localIdentifier

	senderPublicKey *key.NetworkPublic

	msgInChannel chan net.Message

	messageHandlersMutex *sync.RWMutex
	messageHandlers      []*messageHandler
	unmarshalersByType   *sync.Map // <string, net.TaggedUnmarshaler>
}

func unicastConnectWithKey(
	publicKey *key.NetworkPublic,
) *unicastProvider {
	providersMutex.Lock()
	defer providersMutex.Unlock()

	transportID := localIdentifierFromNetworkKey(publicKey)

	existingProvider, ok := providers[transportID]
	if ok {
		return existingProvider
	}

	provider := &unicastProvider{
		publicKey:   publicKey,
		transportID: transportID,
	}

	providers[transportID] = provider

	return provider
}

func (up *unicastProvider) ChannelFor(ctx context.Context, peer net.TransportIdentifier) (net.UnicastChannel, error) {
	peerID := localIdentifier(peer.String())

	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Infof("context is done, closing connection to %s", peerID)
				connectionName := calculateConnectionName(up.transportID, peerID)

				connectionsMutex.Lock()
				delete(connections, connectionName)
				connectionsMutex.Unlock()
				return
			}
		}
	}()
	return up.channel(peerID)
}

func calculateConnectionName(peer1, peer2 localIdentifier) connectionID {
	peers := []string{peer1.String(), peer2.String()}
	sort.Strings(peers)

	return connectionID(strings.Join(peers, "-"))
}

func (up *unicastProvider) channel(peerID localIdentifier) (net.UnicastChannel, error) {
	connectionName := calculateConnectionName(up.transportID, peerID)

	var resultChannel *unicastChannel

	connectionsMutex.Lock()

	connection, connectionExists := connections[connectionName]
	if connectionExists {
		channel, channelExists := connection.channels[up.transportID]
		if channelExists {
			return channel, nil
		}

		newChannel := &unicastChannel{
			transportID:          up.transportID,
			peerID:               peerID,
			senderPublicKey:      up.publicKey,
			msgInChannel:         make(chan net.Message),
			messageHandlersMutex: &sync.RWMutex{},
			messageHandlers:      []*messageHandler{},
			unmarshalersByType:   &sync.Map{},
		}

		connection.channels[up.transportID] = newChannel

		resultChannel = newChannel
		connection.waitInitialized.Done()
	} else {
		waitGroup := &sync.WaitGroup{}
		waitGroup.Add(1)

		resultChannel = &unicastChannel{
			transportID:          up.transportID,
			peerID:               peerID,
			senderPublicKey:      up.publicKey,
			msgInChannel:         make(chan net.Message),
			messageHandlersMutex: &sync.RWMutex{},
			messageHandlers:      []*messageHandler{},
			unmarshalersByType:   &sync.Map{},
		}

		connection = &localConnection{
			connectionID: connectionName,
			channels: map[localIdentifier]*unicastChannel{
				up.transportID: resultChannel,
			},
			waitInitialized: waitGroup,
		}

		connections[connectionName] = connection
	}

	connectionsMutex.Unlock()

	connection.waitInitialized.Wait()

	return resultChannel, nil
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

type messageHandler struct {
	ctx     context.Context
	channel chan net.Message
}

func (uc *unicastChannel) Recv(ctx context.Context, handler func(m net.Message)) {
	messageHandler := &messageHandler{
		ctx:     ctx,
		channel: make(chan net.Message),
	}

	uc.messageHandlersMutex.Lock()
	uc.messageHandlers = append(uc.messageHandlers, messageHandler)
	uc.messageHandlersMutex.Unlock()

	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Debugf("context is done, removing handler")
				uc.removeHandler(messageHandler)
				return
			case msg := <-messageHandler.channel:
				handler(msg)
			case msg := <-uc.msgInChannel:
				logger.Infof("pulling message from buffer")
				handler(msg)
			}
		}
	}()

	return
}

func (uc *unicastChannel) removeHandler(handler *messageHandler) {
	uc.messageHandlersMutex.Lock()
	defer uc.messageHandlersMutex.Unlock()

	for i, h := range uc.messageHandlers {
		if h.channel == handler.channel {
			uc.messageHandlers[i] = uc.messageHandlers[len(uc.messageHandlers)-1]
			uc.messageHandlers = uc.messageHandlers[:len(uc.messageHandlers)-1]
		}
	}
}

func (uc *unicastChannel) Send(message net.TaggedMarshaler) error {
	return uc.doSend(message)
}

func (uc *unicastChannel) doSend(payload net.TaggedMarshaler) error {
	bytes, err := payload.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal payload: [%v]", err)
	}

	connectionsMutex.RLock()
	defer connectionsMutex.RUnlock()

	connection, ok := connections[calculateConnectionName(uc.transportID, uc.peerID)]
	if !ok {
		return fmt.Errorf("failed to find connection to [%s]", uc.peerID)
	}

	for _, channel := range connection.channels {
		if channel.transportID == uc.peerID {
			go func(targetChannel *unicastChannel) {
				unmarshaler, found := targetChannel.unmarshalersByType.Load(payload.Type())
				if !found {
					logger.Errorf(
						"couldn't find unmarshaler for type [%s] in unicast channel",
						payload.Type(),
					)
				}

				unmarshaled := unmarshaler.(func() net.TaggedUnmarshaler)()
				err = unmarshaled.Unmarshal(bytes)
				if err != nil {
					logger.Errorf("failed to unmarshal message: [%v]", err)
				}

				targetChannel.deliver(
					uc.transportID,
					uc.senderPublicKey,
					unmarshaled,
					payload.Type(),
				)
			}(channel)
		}
	}

	return nil
}

func (uc *unicastChannel) deliver(
	senderTransportID localIdentifier,
	senderPublicKey *key.NetworkPublic,
	payload interface{},
	messageType string,
) {
	message := internal.BasicMessage(
		senderTransportID,
		payload,
		messageType,
		key.Marshal(senderPublicKey),
	)

	uc.messageHandlersMutex.RLock()
	defer uc.messageHandlersMutex.RUnlock()

	if len(uc.messageHandlers) == 0 {
		logger.Debugf("no handlers registered put message in buffer")
		uc.msgInChannel <- message
		return
	}

	for _, handler := range uc.messageHandlers {
		go func(handler *messageHandler) {
			select {
			case handler.channel <- message:
			// Nothing to do here; we block until the message is handled
			// or until the context gets closed.
			// This way we don't lose any message but also don't stay
			// with any dangling goroutines if there is no longer anyone
			// to receive messages.
			case <-handler.ctx.Done():
				return
			}
		}(handler)
	}
}
