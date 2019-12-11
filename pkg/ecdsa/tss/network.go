package tss

import (
	"bytes"
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

	groupID string

	party   tss.Party
	params  *tss.Parameters
	errChan chan<- error

	channelsMutex    *sync.Mutex
	broadcastChannel net.BroadcastChannel
	unicastChannels  map[string]net.UnicastChannel
}

// newNetworkBridge initializes a new network bridge for the given network provider.
func newNetworkBridge(networkProvider net.Provider) *networkBridge {
	return &networkBridge{
		networkProvider: networkProvider,
		channelsMutex:   &sync.Mutex{},
		unicastChannels: make(map[string]net.UnicastChannel),
	}
}

func (b *networkBridge) connect(
	groupID string,
	party tss.Party,
	params *tss.Parameters,
	tssOutChan <-chan tss.Message,
	errChan chan<- error,
) error {
	b.groupID = groupID
	b.party = party
	b.params = params
	b.errChan = errChan

	netInChan := make(chan *TSSMessage, params.PartyCount())

	if err := b.initializeChannels(netInChan); err != nil {
		return fmt.Errorf("failed to initialize channels: [%v]", err)
	}

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
		Type: b.groupID,
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
	for _, peerPartyID := range b.params.Parties().IDs() {
		if bytes.Equal(peerPartyID.GetKey(), b.party.PartyID().GetKey()) {
			continue
		}

		unicastChannel, err := b.getUnicastChannelWith(peerPartyID)
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

	broadcastChannel, err := b.networkProvider.BroadcastChannelFor(string(b.groupID))
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

func (b *networkBridge) getUnicastChannelWith(partyID *tss.PartyID) (net.UnicastChannel, error) {
	b.channelsMutex.Lock()
	defer b.channelsMutex.Unlock()

	remotePeerID := string(partyID.GetKey())

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
		b.errChan <- fmt.Errorf("failed to encode message: [%v]", b.party.WrapError(err))
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
			unicastChannel, err := b.getUnicastChannelWith(destination)
			if err != nil {
				b.errChan <- fmt.Errorf(
					"[party-%s]: failed to find unicast channel for [%v]: [%v]",
					b.party.PartyID().String(),
					destination,
					err,
				)
				continue
			}

			if err := unicastChannel.Send(msg); err != nil {
				b.errChan <- fmt.Errorf(
					"[party-%s]: failed to send unicast message: [%v]",
					b.party.PartyID().String(),
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
	if err := b.broadcastChannel.UnregisterRecv(b.groupID); err != nil {
		return fmt.Errorf(
			"failed to unregister receive handler for broadcast channel: [%v]",
			err,
		)
	}

	for _, unicastChannel := range b.unicastChannels {
		if err := unicastChannel.UnregisterRecv(b.groupID); err != nil {
			return fmt.Errorf(
				"failed to unregister receive handler for unicast channel: [%v]",
				err,
			)
		}
	}

	return nil
}
