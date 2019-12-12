package tss

import (
	"fmt"
	"math/big"
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

	errChan chan error
}

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
	}

	return networkBridge, nil
}

func (b *networkBridge) connect(
	tssOutChan <-chan tss.Message,
) error {
	netInChan := make(chan *TSSMessage, len(b.groupInfo.groupMemberIDs))

	if err := b.initializeChannels(netInChan); err != nil {
		return fmt.Errorf("failed to initialize channels: [%v]", err)
	}

	b.errChan = make(chan error, len(b.groupInfo.groupMemberIDs))
	go func() {
		for {
			select {
			case err := <-b.errChan:
				logger.Errorf("network error ocurred: [%v]", err)
			}
		}
	}()

	go func() {
		for {
			select {
			case tssLibMsg := <-tssOutChan:
				go b.sendTSSMessage(tssLibMsg)
			case netMsg := <-netInChan:
				go b.receiveMessage(netMsg)
			}
		}
	}()

	return nil
}

func (b *networkBridge) initializeChannels(netInChan chan *TSSMessage) error {
	handleMessageFunc := net.HandleMessageFunc{
		// TODO: This will be set to group ID now, but we may want to add some
		// session ID for concurrent execution.
		Type: b.groupInfo.groupID,
		Handler: func(msg net.Message) error {
			switch tssMessage := msg.Payload().(type) {
			case *TSSMessage:
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

		unicastChannel, err := b.getUnicastChannelWith(string(peerMemberID))
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

	broadcastChannel, err := b.networkProvider.BroadcastChannelFor(string(b.groupInfo.groupID))
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcast channel: [%v]", err)
	}

	if err := broadcastChannel.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
		return &TSSMessage{}
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
		return &TSSMessage{}
	}); err != nil {
		return nil, fmt.Errorf("failed to register unmarshaler for unicast channel: [%v]", err)
	}

	b.unicastChannels[remotePeerID] = unicastChannel

	return unicastChannel, nil
}

func (b *networkBridge) sendTSSMessage(tssLibMsg tss.Message) {
	bytes, routing, err := tssLibMsg.WireBytes()
	if err != nil {
		b.errChan <- fmt.Errorf("failed to encode message: [%v]", err)
		return
	}

	msg := &TSSMessage{
		SenderID:    routing.From.GetKey(),
		Payload:     bytes,
		IsBroadcast: routing.IsBroadcast,
	}

	if routing.To == nil {
		broadcastChannel, err := b.getBroadcastChannel()
		if err != nil {
			b.errChan <- fmt.Errorf("failed to find broadcast channel: [%v]", err)
			return
		}

		if broadcastChannel.Send(msg); err != nil {
			b.errChan <- fmt.Errorf("failed to send broadcast message: [%v]", err)
			return
		}
	} else {
		for _, destination := range routing.To {
			unicastChannel, err := b.getUnicastChannelWith(string(destination.GetKey()))
			if err != nil {
				b.errChan <- fmt.Errorf(
					"[m:%x]: failed to find unicast channel for [%v]: [%v]",
					b.groupInfo.memberID,
					destination,
					err,
				)
				continue
			}

			if err := unicastChannel.Send(msg); err != nil {
				b.errChan <- fmt.Errorf(
					"[m:%x]: failed to send unicast message: [%v]",
					b.groupInfo.memberID,
					err,
				)
				continue
			}
		}
	}
}

func (b *networkBridge) receiveMessage(netMsg *TSSMessage) {
	senderKey := new(big.Int).SetBytes(netMsg.SenderID)
	senderPartyID := b.params.Parties().IDs().FindByKey(senderKey)

	if senderPartyID == b.party.PartyID() {
		return
	}

	_, err := b.party.UpdateFromBytes(
		netMsg.Payload,
		senderPartyID,
		netMsg.IsBroadcast,
	)
	if err != nil {
		b.errChan <- fmt.Errorf("failed to update party: [%v]", b.party.WrapError(err))
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
