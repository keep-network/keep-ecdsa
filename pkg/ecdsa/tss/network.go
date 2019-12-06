package tss

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/tss"
	"github.com/keep-network/keep-tecdsa/pkg/net"
)

func broadcastChannelName(parties tss.SortedPartyIDs) string {
	sum := big.NewInt(0)
	for _, id := range parties.Keys() {
		sum = sum.Add(sum, id)
	}
	digest := sha256.Sum256(sum.Bytes())

	return hex.EncodeToString(digest[:])
}

func bridgeNetwork(
	networkProvider net.Provider,
	outChan <-chan tss.Message,
	endChan chan<- keygen.LocalPartySaveData,
	errChan chan error,
	party tss.Party,
	params *tss.Parameters,
) {
	recvMessage := make(chan *TSSMessage, params.PartyCount())

	handleMessageFunc := func(channel chan *TSSMessage) net.HandleMessageFunc {
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
	broadcastChannel, err := networkProvider.BroadcastChannelFor(
		broadcastChannelName(params.Parties().IDs()),
	)
	if err != nil {
		errChan <- fmt.Errorf("failed to get broadcast channel: [%v]", err)
	}
	broadcastChannel.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
		return &TSSMessage{}
	})
	broadcastChannel.Recv(handleMessageFunc(recvMessage))

	// Initialize unicast channels.
	unicastChannels := make(map[string]net.UnicastChannel)
	for _, peerPartyID := range params.Parties().IDs() {
		if peerPartyID.GetId() == party.PartyID().GetId() {
			continue
		}

		unicastChannel, err := networkProvider.UnicastChannelWith(peerPartyID.KeyInt().String())
		if err != nil {
			errChan <- fmt.Errorf("failed to get broadcast channel: [%v]", err)
		}
		unicastChannel.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
			return &TSSMessage{}
		})

		unicastChannel.Recv(handleMessageFunc(recvMessage))

		unicastChannels[peerPartyID.KeyInt().String()] = unicastChannel
	}

	go func() {
		for {
			select {
			case tssLibMsg := <-outChan:
				bytes, routing, err := tssLibMsg.WireBytes()
				if err != nil {
					errChan <- fmt.Errorf("failed to encode message: [%v]", party.WrapError(err))
					break
				}

				msg := &TSSMessage{
					SenderID:    routing.From.GetKey(),
					Payload:     bytes,
					IsBroadcast: routing.IsBroadcast,
				}

				if routing.To == nil {
					if err := broadcastChannel.Send(msg); err != nil {
						errChan <- fmt.Errorf("failed to send broadcast message: [%v]", err)
					}
				} else {
					for _, destination := range routing.To {
						peerID := destination.KeyInt().String()

						unicastChannel, ok := unicastChannels[peerID]
						if !ok {
							errChan <- fmt.Errorf("failed to find unicast channel for: [%v]", peerID)
							continue
						}

						unicastChannel.Send(msg)
					}
				}
			case msg := <-recvMessage:
				go func() {
					senderKey := new(big.Int).SetBytes(msg.SenderID)
					senderPartyID := params.Parties().IDs().FindByKey(senderKey)

					bytes := msg.Payload

					if _, err := party.UpdateFromBytes(
						bytes,
						senderPartyID,
						msg.IsBroadcast,
					); err != nil {
						errChan <- party.WrapError(err)
						return
					}
				}()
			}
		}
	}()
}

func unregisterRecv(
	networkProvider net.Provider,
	party tss.Party,
	params *tss.Parameters,
	errChan chan error,
) {
	if broadcastChannel, err := networkProvider.BroadcastChannelFor(
		broadcastChannelName(params.Parties().IDs()),
	); err == nil {
		if err := broadcastChannel.UnregisterRecv(TSSmessageType); err != nil {
			errChan <- fmt.Errorf("failed to unregister receiver: [%v]", err)
		}
	}

	for _, peerPartyID := range params.Parties().IDs() {
		if peerPartyID.GetId() == party.PartyID().GetId() {
			continue
		}

		if unicastChannel, err := networkProvider.UnicastChannelWith(
			peerPartyID.KeyInt().String(),
		); err != nil {
			if err := unicastChannel.UnregisterRecv(TSSmessageType); err != nil {
				errChan <- fmt.Errorf("failed to unregister receiver: [%v]", err)
			}
		}
	}
}
