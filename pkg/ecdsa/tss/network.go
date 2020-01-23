package tss

import (
	"context"
	"fmt"
	"sync"

	"github.com/binance-chain/tss-lib/tss"
	"github.com/keep-network/keep-tecdsa/pkg/net"
)

// NetworkID is an implementation of net.TransportIdentifier interface. It defines
// operator's network layer identifier used for unicast messages routing.
// TODO: We can consider replacing it with a type implemented in specific network
// implementation we will use.
type NetworkID string

func (n NetworkID) String() string {
	return string(n)
}

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
		unicastChannels: make(map[string]net.UnicastChannel),

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
	b.netInChan = make(chan *TSSProtocolMessage, len(b.groupInfo.groupMemberIDs))

	if err := b.initializeBroadcastChannel(ctx); err != nil {
		return fmt.Errorf("failed to initialize channels: [%v]", err)
	}

	go func() {
		for {
			select {
			case tssLibMsg := <-tssOutChan:
				go b.sendTSSMessage(ctx, tssLibMsg)
			case msg := <-b.netInChan:
				go b.handleTSSProtocolMessage(msg)
			case <-ctx.Done():
				if err := b.unregisterRecvs(); err != nil {
					logger.Errorf("failed to unregister receivers: [%v]", err)
				}

				return
			}
		}
	}()

	b.registerProtocolMessageHandler(party, sortedPartyIDs)

	return nil
}

func (b *networkBridge) initializeBroadcastChannel(ctx context.Context) error {
	_, err := b.getBroadcastChannel(ctx)
	if err != nil {
		return fmt.Errorf("failed to get broadcast channel: [%v]", err)
	}

	return nil
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
		if peerMemberID.Equal(b.groupInfo.memberID) {
			continue
		}

		unicastChannel, err := b.getUnicastChannelWith(
			b.groupInfo.membersNetworkIDs[peerMemberID.String()],
		)
		if err != nil {
			return fmt.Errorf("failed to get unicast channel: [%v]", err)
		}

		if err := unicastChannel.Recv(handleMessageFunc); err != nil {
			return fmt.Errorf("failed to register receive handler for unicast channel: [%v]", err)
		}
	}

func (b *networkBridge) initializeUnicastChannels(
	ctx context.Context,
	membersPublicKeys map[string][]byte,
) error {
	b.membersPublicKeys = membersPublicKeys

	// Initialize unicast channels.
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(len(b.groupInfo.groupMemberIDs) - 1)

	for _, peerMemberID := range b.groupInfo.groupMemberIDs {
		go func(peerMemberID MemberID) {
			if peerMemberID.Equal(b.groupInfo.memberID) {
				return
			}

			_, err := b.getUnicastChannelWith(ctx, peerMemberID)
			if err != nil {
				logger.Errorf("failed to get unicast channel with [%s]: [%v]", peerMemberID, err)
			}

			waitGroup.Done()
		}(peerMemberID)
	}

	waitGroup.Wait()

	return nil
}

func (b *networkBridge) getUnicastChannelWith(remotePeerID net.TransportIdentifier) (net.UnicastChannel, error) {
	b.channelsMutex.Lock()
	defer b.channelsMutex.Unlock()

	unicastChannel, exists := b.unicastChannels[remotePeerID.String()]
	if exists {
		return unicastChannel, nil
	}

	unicastChannel, err := b.networkProvider.UnicastChannelWith(remotePeerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unicast channel: [%v]", err)
	}

	if err := unicastChannel.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
		return &TSSProtocolMessage{}
	}); err != nil {
		return nil, fmt.Errorf("failed to register unmarshaler for unicast channel: [%v]", err)
	}

	b.unicastChannels[remotePeerID.String()] = unicastChannel

	return unicastChannel, nil
}

func (b *networkBridge) sendTSSMessage(tssLibMsg tss.Message) {
	bytes, routing, err := tssLibMsg.WireBytes()
	if err != nil {
		logger.Errorf("failed to encode message: [%v]", err)
		return
	}

	if routing.To == nil {
		b.broadcast(
			ctx,
			&TSSProtocolMessage{
				SenderID:    memberIDFromBytes(routing.From.GetKey()),
				Payload:     bytes,
				IsBroadcast: routing.IsBroadcast,
			},
		)
	} else {
		for _, destination := range routing.To {
			b.sendTo(
				b.groupInfo.membersNetworkIDs[destination.GetId()],
				&TSSProtocolMessage{
					SenderID:    memberIDFromBytes(routing.From.GetKey()),
					ReceiverID:  memberIDFromBytes(destination.GetKey()),
					Payload:     bytes,
					IsBroadcast: routing.IsBroadcast,
				},
			)
		}
	}
}

func (b *networkBridge) broadcast(ctx context.Context, msg *TSSProtocolMessage) error {
	broadcastChannel, err := b.getBroadcastChannel(ctx)
	if err != nil {
		return fmt.Errorf("failed to find broadcast channel: [%v]", err)

	}

	if broadcastChannel.Send(msg); err != nil {
		return fmt.Errorf("failed to send broadcast message: [%v]", err)
	}

	return nil
}

func (b *networkBridge) sendTo(ctx context.Context, msg *TSSProtocolMessage) error { // TODO: Rename to `send`
	receiverID := msg.ReceiverID
	if receiverID == nil {
		return fmt.Errorf("receiver id not provided")
	}

	if b.isSameOperator(receiverID) {
		// TODO: Not supported
		return fmt.Errorf("remote peer network ID same as current member network ID")
	}

	unicastChannel, err := b.getUnicastChannelWith(ctx, receiverID)
	if err != nil {
		return fmt.Errorf(
			"[m:%x]: failed to find unicast channel for [%v]: [%v]",
			b.groupInfo.memberID,
			receiverID,
			err,
		)

	}

	if err := unicastChannel.Send(msg); err != nil {
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

		// Ignore unicast messages addressed to other parties.
		if !protocolMessage.IsBroadcast {
			receiverPartyID := sortedPartyIDs.FindByKey(protocolMessage.ReceiverID.bigInt())

			if receiverPartyID != party.PartyID() {
				return nil
			}
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
