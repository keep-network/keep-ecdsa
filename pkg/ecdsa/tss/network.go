package tss

import (
	"fmt"
	"sync"

	"github.com/binance-chain/tss-lib/tss"
	"github.com/keep-network/keep-tecdsa/pkg/net"
)

// networkBridge translates TSS library network interface to unicast and
// broadcast channels provided by our net abstraction.
type networkBridge struct {
	networkProvider net.Provider

	groupInfo *groupInfo

	channelsMutex    *sync.Mutex
	broadcastChannel net.BroadcastChannel
	unicastChannels  map[string]net.UnicastChannel

	tssMessageHandlersMutex *sync.Mutex
	tssMessageHandlers      []tssMessageHandler
}

type tssMessageHandler func(netMsg *ProtocolMessage) error

// newNetworkBridge initializes a new network bridge for the given network provider.
func newNetworkBridge(
	groupInfo *groupInfo,
	networkProvider net.Provider,
) (*networkBridge, error) {
	networkBridge := &networkBridge{
		networkProvider: networkProvider,
		groupInfo:       groupInfo,

		channelsMutex:   &sync.Mutex{},
		unicastChannels: make(map[string]net.UnicastChannel),

		tssMessageHandlersMutex: &sync.Mutex{},
		tssMessageHandlers:      []tssMessageHandler{},
	}

	return networkBridge, nil
}

func (b *networkBridge) connect(
	tssOutChan <-chan tss.Message,
) error {
	netInChan := make(chan interface{}, len(b.groupInfo.groupMemberIDs))

	if err := b.initializeChannels(netInChan); err != nil {
		return fmt.Errorf("failed to initialize channels: [%v]", err)
	}

	go func() {
		for {
			select {
			case tssLibMsg := <-tssOutChan:
				go b.sendTSSMessage(tssLibMsg)
			case inMessage := <-netInChan:
				switch msg := inMessage.(type) {
				case *ProtocolMessage:
					go b.handleTSSMessage(msg)
				}
			}
		}
	}()

	return nil
}

func (b *networkBridge) initializeChannels(netInChan chan interface{}) error {
	handleMessageFunc := net.HandleMessageFunc{
		// TODO: This will be set to group ID now, but we may want to add some
		// session ID for concurrent execution.
		Type: b.groupInfo.groupID,
		Handler: func(msg net.Message) error {
			switch tssMessage := msg.Payload().(type) {
			case *ProtocolMessage:
				netInChan <- tssMessage
			}

			return nil
		},
	}

	// Initialize broadcast channel.
	broadcastChannel, err := b.getBroadcastChannel()
	if err != nil {
		return fmt.Errorf("failed to get broadcast channel: [%v]", err)
	}

	if err := broadcastChannel.Recv(handleMessageFunc); err != nil {
		return fmt.Errorf("failed to register receive handler for broadcast channel: [%v]", err)
	}

	// Initialize unicast channels.
	for _, peerMemberID := range b.groupInfo.groupMemberIDs {
		if peerMemberID == b.groupInfo.memberID {
			continue
		}

		unicastChannel, err := b.getUnicastChannelWith(peerMemberID.string())
		if err != nil {
			return fmt.Errorf("failed to get unicast channel: [%v]", err)
		}

		if err := unicastChannel.Recv(handleMessageFunc); err != nil {
			return fmt.Errorf("failed to register receive handler for unicast channel: [%v]", err)
		}
	}

	return nil
}

func (b *networkBridge) getBroadcastChannel() (net.BroadcastChannel, error) {
	b.channelsMutex.Lock()
	defer b.channelsMutex.Unlock()

	if b.broadcastChannel != nil {
		return b.broadcastChannel, nil
	}

	broadcastChannel, err := b.networkProvider.BroadcastChannelFor(b.groupInfo.groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcast channel: [%v]", err)
	}

	if err := broadcastChannel.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
		return &ProtocolMessage{}
	}); err != nil {
		return nil, fmt.Errorf("failed to register unmarshaler for broadcast channel: [%v]", err)
	}

	b.broadcastChannel = broadcastChannel

	return broadcastChannel, nil
}

func (b *networkBridge) getUnicastChannelWith(remotePeerID string) (net.UnicastChannel, error) {
	b.channelsMutex.Lock()
	defer b.channelsMutex.Unlock()

	unicastChannel, exists := b.unicastChannels[remotePeerID]
	if exists {
		return unicastChannel, nil
	}

	unicastChannel, err := b.networkProvider.UnicastChannelWith(remotePeerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unicast channel: [%v]", err)
	}

	if err := unicastChannel.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
		return &ProtocolMessage{}
	}); err != nil {
		return nil, fmt.Errorf("failed to register unmarshaler for unicast channel: [%v]", err)
	}

	b.unicastChannels[remotePeerID] = unicastChannel

	return unicastChannel, nil
}

func (b *networkBridge) sendTSSMessage(tssLibMsg tss.Message) {
	bytes, routing, err := tssLibMsg.WireBytes()
	if err != nil {
		logger.Errorf("failed to encode message: [%v]", err)
		return
	}

	protocolMessage := &ProtocolMessage{
		SenderID:    memberIDFromBytes(routing.From.GetKey()),
		Payload:     bytes,
		IsBroadcast: routing.IsBroadcast,
	}

	if routing.To == nil {
		b.sendMessage(&MessageRouting{
			ReceiverID: "",
			Message:    protocolMessage,
		})
	} else {
		for _, destination := range routing.To {
			b.sendMessage(&MessageRouting{
				ReceiverID: destination.GetId(),
				Message:    protocolMessage,
			})
		}
	}
}

func (b *networkBridge) sendMessage(msg *MessageRouting) error {
	if msg.ReceiverID == "" {
		broadcastChannel, err := b.getBroadcastChannel()
		if err != nil {
			return fmt.Errorf("failed to find broadcast channel: [%v]", err)

		}

		if broadcastChannel.Send(msg.Message); err != nil {
			return fmt.Errorf("failed to send broadcast message: [%v]", err)
		}
	} else {
		unicastChannel, err := b.getUnicastChannelWith(msg.ReceiverID)
		if err != nil {
			return fmt.Errorf(
				"[m:%x]: failed to find unicast channel for [%v]: [%v]",
				b.groupInfo.memberID,
				msg.ReceiverID,
				err,
			)

		}

		if err := unicastChannel.Send(msg.Message); err != nil {
			return fmt.Errorf(
				"[m:%x]: failed to send unicast message: [%v]",
				b.groupInfo.memberID,
				err,
			)

		}
	}
	return nil
}

func (b *networkBridge) registerTSSMessageHandler(
	party tss.Party,
	sortedPartyIDs tss.SortedPartyIDs,
) {
	handler := func(protocolMessage *ProtocolMessage) error {
		senderPartyID := sortedPartyIDs.FindByKey(protocolMessage.SenderID.bigInt())

		if senderPartyID == party.PartyID() {
			return nil
		}

		_, err := party.UpdateFromBytes(
			protocolMessage.Payload,
			senderPartyID,
			protocolMessage.IsBroadcast,
		)
		if err != nil {
			return fmt.Errorf("failed to update party: [%v]", party.WrapError(err))
		}

		return nil
	}

	b.tssMessageHandlersMutex.Lock()
	defer b.tssMessageHandlersMutex.Unlock()

	b.tssMessageHandlers = append(b.tssMessageHandlers, handler)
}

func (b *networkBridge) handleTSSMessage(protocolMessage *ProtocolMessage) {
	b.tssMessageHandlersMutex.Lock()
	defer b.tssMessageHandlersMutex.Unlock()

	for _, handler := range b.tssMessageHandlers {
		if err := handler(protocolMessage); err != nil {
			logger.Errorf("failed to handle protocol message: [%v]", err)
		}
	}
}

func (b *networkBridge) close() error {
	if err := b.unregisterRecvs(); err != nil {
		return fmt.Errorf("failed to unregister receivers: [%v]", err)
	}

	return nil
}

func (b *networkBridge) unregisterRecvs() error {
	if err := b.broadcastChannel.UnregisterRecv(b.groupInfo.groupID); err != nil {
		return fmt.Errorf(
			"failed to unregister receive handler for broadcast channel: [%v]",
			err,
		)
	}

	for _, unicastChannel := range b.unicastChannels {
		if err := unicastChannel.UnregisterRecv(b.groupInfo.groupID); err != nil {
			return fmt.Errorf(
				"failed to unregister receive handler for unicast channel: [%v]",
				err,
			)
		}
	}

	return nil
}
