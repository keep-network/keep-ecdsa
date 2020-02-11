package tss

import (
	"context"
	"fmt"
	"github.com/binance-chain/tss-lib/tss"
	"github.com/keep-network/keep-core/pkg/net"
	"sync"
	"time"
)

const (
	unicastChannelRetryCount    = 2
	unicastChannelRetryWaitTime = 30 * time.Second
)

// networkBridge translates TSS library network interface to unicast and
// broadcast channels provided by our net abstraction.
type networkBridge struct {
	networkProvider net.Provider

	groupInfo *groupInfo

	channelsMutex    *sync.Mutex
	broadcastChannel net.BroadcastChannel
	unicastChannels  map[net.TransportIdentifier]net.UnicastChannel

	tssMessageHandlersMutex *sync.Mutex
	tssMessageHandlers      []tssMessageHandler
}

type tssMessageHandler func(netMsg *TSSProtocolMessage) error

// newNetworkBridge initializes a new network bridge for the given network provider.
func newNetworkBridge(
	groupInfo *groupInfo,
	networkProvider net.Provider,
) (*networkBridge, error) {
	networkBridge := &networkBridge{
		networkProvider: networkProvider,
		groupInfo:       groupInfo,

		channelsMutex:   &sync.Mutex{},
		unicastChannels: make(map[net.TransportIdentifier]net.UnicastChannel),

		tssMessageHandlersMutex: &sync.Mutex{},
		tssMessageHandlers:      []tssMessageHandler{},
	}

	return networkBridge, nil
}

func (b *networkBridge) connect(
	ctx context.Context,
	tssOutChan <-chan tss.Message,
	party tss.Party,
	sortedPartyIDs tss.SortedPartyIDs,
) error {
	netInChan := make(chan *TSSProtocolMessage, len(b.groupInfo.groupMemberIDs))

	if err := b.initializeChannels(ctx, netInChan); err != nil {
		return fmt.Errorf("failed to initialize channels: [%v]", err)
	}

	go func() {
		for {
			select {
			case tssLibMsg := <-tssOutChan:
				go b.sendTSSMessage(tssLibMsg)
			case msg := <-netInChan:
				go b.handleTSSProtocolMessage(msg)
			case <-ctx.Done():
				return
			}
		}
	}()

	b.registerProtocolMessageHandler(party, sortedPartyIDs)

	return nil
}

func (b *networkBridge) initializeChannels(
	ctx context.Context,
	netInChan chan *TSSProtocolMessage,
) error {
	handleFn := func(msg net.Message) {
		switch protocolMessage := msg.Payload().(type) {
		case *TSSProtocolMessage:
			netInChan <- protocolMessage
		}
	}

	// Initialize broadcast channel.
	broadcastChannel, err := b.getBroadcastChannel()
	if err != nil {
		return fmt.Errorf("failed to get broadcast channel: [%v]", err)
	}

	broadcastChannel.Recv(ctx, handleFn)

	// Initialize unicast channels.
	for _, peerMemberID := range b.groupInfo.groupMemberIDs {
		if peerMemberID.Equal(b.groupInfo.memberID) {
			continue
		}

		peerTransportID, err := b.getTransportIdentifier(peerMemberID)
		if err != nil {
			return fmt.Errorf("failed to get transport identifier: [%v]", err)
		}

		unicastChannel, err := b.getUnicastChannel(
			peerTransportID,
			unicastChannelRetryCount,
			unicastChannelRetryWaitTime,
		)
		if err != nil {
			return fmt.Errorf("failed to get unicast channel: [%v]", err)
		}

		unicastChannel.Recv(ctx, handleFn)
	}

	return nil
}

func (b *networkBridge) getUnicastChannel(
	peerTransportID net.TransportIdentifier,
	retryCount int,
	retryWaitTime time.Duration,
) (net.UnicastChannel, error) {
	var (
		unicastChannel net.UnicastChannel
		err            error
	)

	for i := 0; i < retryCount+1; i++ {
		unicastChannel, err = b.getUnicastChannelWith(peerTransportID)
		if unicastChannel != nil && err == nil {
			return unicastChannel, nil
		}

		logger.Warningf(
			"failed to get unicast channel with peer [%v] "+
				"because of: [%v]; will retry after wait time",
			peerTransportID.String(),
			err,
		)

		time.Sleep(retryWaitTime)
	}

	if err == nil {
		err = fmt.Errorf("unknown error")
	}

	return nil, err
}

func (b *networkBridge) getTransportIdentifier(member MemberID) (net.TransportIdentifier, error) {
	publicKey := b.groupInfo.groupMemberPublicKeys[member.String()]
	return b.networkProvider.CreateTransportIdentifier(publicKey)
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
		return &TSSProtocolMessage{}
	}); err != nil {
		return nil, fmt.Errorf("failed to register unmarshaler for broadcast channel: [%v]", err)
	}

	b.broadcastChannel = broadcastChannel

	return broadcastChannel, nil
}

func (b *networkBridge) getUnicastChannelWith(
	peerTransportID net.TransportIdentifier,
) (net.UnicastChannel, error) {
	b.channelsMutex.Lock()
	defer b.channelsMutex.Unlock()

	unicastChannel, exists := b.unicastChannels[peerTransportID]
	if exists {
		return unicastChannel, nil
	}

	unicastChannel, err := b.networkProvider.UnicastChannelWith(peerTransportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unicast channel: [%v]", err)
	}

	unicastChannel.SetUnmarshaler(func() net.TaggedUnmarshaler {
		return &TSSProtocolMessage{}
	})

	b.unicastChannels[peerTransportID] = unicastChannel

	return unicastChannel, nil
}

func (b *networkBridge) sendTSSMessage(tssLibMsg tss.Message) {
	bytes, routing, err := tssLibMsg.WireBytes()
	if err != nil {
		logger.Errorf("failed to encode message: [%v]", err)
		return
	}

	protocolMessage := &TSSProtocolMessage{
		SenderID:    memberIDFromBytes(routing.From.GetKey()),
		Payload:     bytes,
		IsBroadcast: routing.IsBroadcast,
	}

	if routing.To == nil {
		b.broadcast(protocolMessage)
	} else {
		for _, destination := range routing.To {
			destinationMemberID, err := MemberIDFromHex(destination.GetId())
			if err != nil {
				logger.Errorf("failed to get destination member id: [%v]", err)
				return
			}
			destinationTransportID, err := b.getTransportIdentifier(destinationMemberID)
			if err != nil {
				logger.Errorf("failed to get transport identifier: [%v]", err)
			}
			b.sendTo(destinationTransportID, protocolMessage)
		}
	}
}

func (b *networkBridge) broadcast(msg *TSSProtocolMessage) error {
	broadcastChannel, err := b.getBroadcastChannel()
	if err != nil {
		return fmt.Errorf("failed to find broadcast channel: [%v]", err)

	}

	if broadcastChannel.Send(context.Background(), msg); err != nil {
		return fmt.Errorf("failed to send broadcast message: [%v]", err)
	}

	return nil
}

func (b *networkBridge) sendTo(
	receiverTransportID net.TransportIdentifier,
	message *TSSProtocolMessage,
) error {
	unicastChannel, err := b.getUnicastChannelWith(receiverTransportID)
	if err != nil {
		return fmt.Errorf(
			"[m:%x]: failed to find unicast channel for [%v]: [%v]",
			b.groupInfo.memberID,
			receiverTransportID,
			err,
		)

	}

	if err := unicastChannel.Send(message); err != nil {
		return fmt.Errorf(
			"[m:%x]: failed to send unicast message: [%v]",
			b.groupInfo.memberID,
			err,
		)

	}

	return nil
}

func (b *networkBridge) registerProtocolMessageHandler(
	party tss.Party,
	sortedPartyIDs tss.SortedPartyIDs,
) {
	handler := func(protocolMessage *TSSProtocolMessage) error {
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

func (b *networkBridge) handleTSSProtocolMessage(protocolMessage *TSSProtocolMessage) {
	b.tssMessageHandlersMutex.Lock()
	defer b.tssMessageHandlersMutex.Unlock()

	for _, handler := range b.tssMessageHandlers {
		if err := handler(protocolMessage); err != nil {
			logger.Errorf("failed to handle protocol message: [%v]", err)
		}
	}
}
