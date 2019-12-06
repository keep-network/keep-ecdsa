package tss

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/binance-chain/tss-lib/tss"
	"github.com/keep-network/keep-tecdsa/pkg/net"
)

type networkBridge struct {
	networkProvider net.Provider
	party           tss.Party
	params          *tss.Parameters
	errChan         chan<- error

	broadcastChannel net.BroadcastChannel
	unicastChannels  map[string]net.UnicastChannel
}

func bridgeNetwork(
	networkProvider net.Provider,
	outChan <-chan tss.Message,
	errChan chan<- error,
	doneChan <-chan struct{},
	party tss.Party,
	params *tss.Parameters,
) error {
	recvMessage := make(chan *TSSMessage, params.PartyCount())

	bridge := &networkBridge{
		networkProvider: networkProvider,
		party:           party,
		params:          params,
		errChan:         errChan,
	}

	if err := bridge.initializeChannels(recvMessage); err != nil {
		return fmt.Errorf("failed to initialize channels: [%v]", err)
	}

	go func() {
		for {
			select {
			case tssLibMsg := <-outChan:
				go func(tssLibMsg tss.Message) {
					bridge.sendMessage(tssLibMsg)
				}(tssLibMsg)
			case msg := <-recvMessage:
				go func(msg *TSSMessage) {
					senderKey := new(big.Int).SetBytes(msg.SenderID)
					senderPartyID := params.Parties().IDs().FindByKey(senderKey)

					if senderPartyID == party.PartyID() {
						return
					}

					bytes := msg.Payload

					_, err := party.UpdateFromBytes(
						bytes,
						senderPartyID,
						msg.IsBroadcast,
					)
					if err != nil {
						errChan <- party.WrapError(err)
					}
				}(msg)
			case <-doneChan:
				bridge.unregisterRecv()
				return
			}
		}
	}()

	return nil
}

func broadcastChannelName(parties tss.SortedPartyIDs) string {
	ids := []string{}
	for _, id := range parties.Keys() {
		ids = append(ids, id.Text(16))
	}

	digest := sha256.Sum256([]byte(strings.Join(ids, "-")))

	return hex.EncodeToString(digest[:])
}

func (b *networkBridge) initializeChannels(recvMessageChan chan *TSSMessage) error {
	handleMessageFunc := func(channel chan<- *TSSMessage) net.HandleMessageFunc {
		return net.HandleMessageFunc{
			Type: TSSmessageType,
			Handler: func(msg net.Message) error {
				switch tssMessage := msg.Payload().(type) {
				case *TSSMessage:
					channel <- tssMessage
				default:
					return fmt.Errorf("unexpected message: [%v]", msg.Payload())
				}

				return nil
			},
		}
	}

	// Initialize broadcast channel.
	broadcastChannel, err := b.networkProvider.BroadcastChannelFor(
		broadcastChannelName(b.params.Parties().IDs()),
	)
	if err != nil {
		return fmt.Errorf("failed to get broadcast channel: [%v]", err)
	}

	if err := broadcastChannel.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
		return &TSSMessage{}
	}); err != nil {
		// TODO: For signing it fails because unmarshaler was already registered on keygen.
		// return fmt.Errorf("failed to register unmarshaler for broadcast channel: [%v]", err)
	}

	if err := broadcastChannel.Recv(handleMessageFunc(recvMessageChan)); err != nil {
		return fmt.Errorf("failed to register receive handler for broadcast channel: [%v]", err)
	}

	// Initialize unicast channels.
	unicastChannels := make(map[string]net.UnicastChannel)
	for _, peerPartyID := range b.params.Parties().IDs() {
		if bytes.Equal(peerPartyID.GetKey(), b.party.PartyID().GetKey()) {
			continue
		}

		unicastChannel, err := b.networkProvider.UnicastChannelWith(peerPartyID.KeyInt().Text(16))
		if err != nil {
			return fmt.Errorf("failed to get broadcast channel: [%v]", err)
		}

		if err := unicastChannel.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
			return &TSSMessage{}
		}); err != nil {
			// TODO: For signing it fails because unmarshaler was already registered on keygen.
			// return fmt.Errorf("failed to register unmarshaler: [%v]", err)
		}

		if err := unicastChannel.Recv(handleMessageFunc(recvMessageChan)); err != nil {
			return fmt.Errorf("failed to register receive handler for unicast channel: [%v]", err)
		}

		unicastChannels[peerPartyID.KeyInt().String()] = unicastChannel
	}

	b.broadcastChannel = broadcastChannel
	b.unicastChannels = unicastChannels

	return nil
}

func (b *networkBridge) sendMessage(tssLibMsg tss.Message) {
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
		if err := b.broadcastChannel.Send(msg); err != nil {
			b.errChan <- fmt.Errorf("failed to send broadcast message: [%v]", err)
			return
		}
	} else {
		for _, destination := range routing.To {
			// peerNetID := destination.KeyInt().Text(16)
			peerNetID := destination.KeyInt().String()

			unicastChannel, ok := b.unicastChannels[peerNetID]
			if !ok {
				b.errChan <- fmt.Errorf("failed to find unicast channel for: [%v]", peerNetID)
				continue
			}

			if err := unicastChannel.Send(msg); err != nil {
				b.errChan <- fmt.Errorf("failed to send unicast message: [%v]", err)
				continue
			}
		}
	}
}

func (b *networkBridge) receiveMessage(msg TSSMessage) {
	senderKey := new(big.Int).SetBytes(msg.SenderID)
	senderPartyID := b.params.Parties().IDs().FindByKey(senderKey)

	if senderPartyID == b.party.PartyID() {
		return
	}

	bytes := msg.Payload

	_, err := b.party.UpdateFromBytes(
		bytes,
		senderPartyID,
		msg.IsBroadcast,
	)
	if err != nil {
		b.errChan <- fmt.Errorf("failed to update party: [%v]", b.party.WrapError(err))
	}
}

func (b *networkBridge) unregisterRecv() {
	if err := b.broadcastChannel.UnregisterRecv(TSSmessageType); err != nil {
		b.errChan <- fmt.Errorf(
			"failed to unregister receive handler for broadcast channel: [%v]",
			err,
		)
	}

	for _, unicastChannel := range b.unicastChannels {
		if err := unicastChannel.UnregisterRecv(TSSmessageType); err != nil {
			b.errChan <- fmt.Errorf(
				"failed to unregister receive handler for unicast channel: [%v]",
				err,
			)
		}
	}
}
