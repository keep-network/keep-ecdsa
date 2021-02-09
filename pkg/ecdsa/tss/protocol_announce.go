package tss

import (
	"context"
	"fmt"
	"time"

	"github.com/keep-network/keep-core/pkg/net"
	"github.com/keep-network/keep-core/pkg/operator"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

const protocolAnnounceTimeout = 2 * time.Minute

func AnnounceProtocol(
	parentCtx context.Context,
	publicKey *operator.PublicKey,
	keepAddress chain.KeepID,
	keepMemberAddresses []chain.KeepMemberID,
	broadcastChannel net.BroadcastChannel,
) (
	[]MemberID,
	error,
) {
	logger.Infof("announcing presence")

	ctx, cancel := context.WithTimeout(parentCtx, protocolAnnounceTimeout)
	defer cancel()

	announceInChan := make(chan *AnnounceMessage, len(keepMemberAddresses))
	handleAnnounceMessage := func(netMsg net.Message) {
		switch msg := netMsg.Payload().(type) {
		case *AnnounceMessage:
			announceInChan <- msg
		}
	}
	broadcastChannel.Recv(ctx, handleAnnounceMessage)

	receivedMemberIDs := make(map[string]chain.KeepMemberID)

	markAnnounced := func(memberID MemberID, keepMemberID chain.KeepMemberID) {
		receivedMemberIDs[memberID.String()] = keepMemberID
	}
	hasAnnounced := func(memberID MemberID) bool {
		_, ok := receivedMemberIDs[memberID.String()]
		return ok
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-announceInChan:
				logger.Infof(
					"member [%s] from keep [%s] announced its presence",
					msg.SenderID,
					keepAddress,
				)

				markAnnounced(msg.SenderID)
				if len(receivedMemberIDs) == len(keepMemberAddresses) {
					cancel()
				}
			}
		}
	}()

	go func() {
		sendMessage := func() {
			if err := broadcastChannel.Send(ctx,
				&AnnounceMessage{
					SenderID: MemberIDFromPublicKey(publicKey),
				},
			); err != nil {
				logger.Errorf("failed to send announcement: [%v]", err)
			}
		}

		// Send the message first time. It will be periodically retransmitted
		// by the broadcast channel for the entire lifetime of the context.
		sendMessage()

		<-ctx.Done()
		// Send the message once again as the member received messages
		// from all peer members but not all peer members could receive
		// the message from the member as some peer member could join
		// the protocol after the member sent the last message.
		sendMessage()
		return
	}()

	<-ctx.Done()

	switch ctx.Err() {
	case context.DeadlineExceeded:
		for _, member := range keepMemberAddresses {
			if !hasAnnounced(member) {
				logger.Errorf(
					"member [%s] has not announced its presence for keep [%s]; "+
						"check if keep client for that operator is active and "+
						"connected",
					member,
					keepAddress,
				)
			}
		}
		return nil, fmt.Errorf(
			"waiting for announcements timed out after: [%v]",
			protocolAnnounceTimeout,
		)
	case context.Canceled:
		logger.Infof("announce protocol completed successfully")

		memberIDs := make([]MemberID, 0)
		for _, memberID := range receivedMemberIDs {
			memberIDs = append(memberIDs, memberID)
		}

		return memberIDs, nil
	default:
		return nil, fmt.Errorf("unexpected context error: [%v]", ctx.Err())
	}
}
