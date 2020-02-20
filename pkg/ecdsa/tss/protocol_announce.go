package tss

import (
	"context"
	cecdsa "crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"time"

	"github.com/keep-network/keep-core/pkg/net"
)

const protocolAnnounceTimeout = 120 * time.Second

func announceProtocol(
	parentCtx context.Context,
	group *groupInfo,
	networkProvider net.Provider,
) (
	map[string]cecdsa.PublicKey,
	error,
) {
	logger.Infof("starting announce protocol")

	ctx, cancel := context.WithTimeout(parentCtx, protocolAnnounceTimeout)
	defer cancel()

	broadcastChannel, err := networkProvider.BroadcastChannelFor(group.groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize broadcast channel: [%v]", err)
	}

	broadcastChannel.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
		return &AnnounceMessage{}
	})

	// TODO: register group member filter

	announceInChan := make(chan *AnnounceMessage, len(group.groupMemberIDs))
	handleAnnounceMessage := func(netMsg net.Message) {
		switch msg := netMsg.Payload().(type) {
		case *AnnounceMessage:
			announceInChan <- msg
		}
	}
	broadcastChannel.Recv(ctx, handleAnnounceMessage)

	groupMemberPublicKeys := make(map[string]cecdsa.PublicKey)

	go func() {

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-announceInChan:
				for _, memberID := range group.groupMemberIDs {
					if msg.SenderID.Equal(memberID) && isValidAnnouncement(msg) {
						groupMemberPublicKeys[msg.SenderID.String()] = *msg.SenderPublicKey
						break
					}
				}

				if len(groupMemberPublicKeys) == len(group.groupMemberIDs) {
					cancel()
				}
			}
		}
	}()

	go func() {
		sendMessage := func() {
			if err := broadcastChannel.Send(ctx,
				&AnnounceMessage{
					SenderID:        group.memberID,
					SenderPublicKey: &group.memberPublicKey,
				},
			); err != nil {
				logger.Errorf("failed to send announcement: [%v]", err)
			}
		}

		for {
			select {
			case <-ctx.Done():
				// Send the message once again as the member received messages
				// from all peer members but not all peer members could receive
				// the message from the member as some peer member could join
				// the protocol after the member sent the last message.
				sendMessage()
				return
			default:
				sendMessage()
				time.Sleep(1 * time.Second)
			}
		}
	}()

	<-ctx.Done()
	switch ctx.Err() {
	case context.DeadlineExceeded:
		return nil, fmt.Errorf(
			"waiting for announcements timed out after: [%v]",
			protocolAnnounceTimeout,
		)
	case context.Canceled:
		logger.Infof("announce protocol completed successfully")
		return groupMemberPublicKeys, nil
	default:
		return nil, fmt.Errorf("unexpected context error: [%v]", ctx.Err())
	}
}

func isValidAnnouncement(message *AnnounceMessage) bool {
	resolvedAddress := crypto.PubkeyToAddress(*message.SenderPublicKey).Bytes()
	return memberIDFromBytes(resolvedAddress).Equal(message.SenderID)
}
