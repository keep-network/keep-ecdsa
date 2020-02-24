package tss

import (
	"context"
	"fmt"
	"time"

	"github.com/keep-network/keep-core/pkg/net"
)

// protocolJoinTimeout defines a period within which the member sends and receives
// notifications from peer members about their readiness to begin the protocol
// execution. If the time limit is reached the join protocol stage fails.
const protocolJoinTimeout = 120 * time.Second

// joinProtocol exchanges messages with peer members about readiness to start
// the protocol execution. The member keeps sending the message in intervals
// until they receive messages from all peer members. Function exits without an
// error if messages were received from all peer members. If the timeout is
// reached before receiving messages from all peer members the function returns
// an error.
//
// TODO: consider renaming of `joinProtocol` to something related with readiness signaling.
func joinProtocol(
	parentCtx context.Context,
	group *groupInfo,
	broadcastChannel net.BroadcastChannel,
) error {
	ctx, cancel := context.WithTimeout(parentCtx, protocolJoinTimeout)
	defer cancel()

	joinInChan := make(chan *JoinMessage, len(group.groupMemberIDs))
	handleJoinMessage := func(netMsg net.Message) {
		switch msg := netMsg.Payload().(type) {
		case *JoinMessage:
			joinInChan <- msg
		}
	}
	broadcastChannel.Recv(ctx, handleJoinMessage)

	go func() {
		readyMembers := make(map[string]bool)

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-joinInChan:
				for _, memberID := range group.groupMemberIDs {
					if msg.SenderID.Equal(memberID) {
						readyMembers[msg.SenderID.String()] = true
						break
					}
				}

				if len(readyMembers) == len(group.groupMemberIDs) {
					cancel()
				}
			}
		}
	}()

	go func() {
		sendMessage := func() {
			if err := broadcastChannel.Send(ctx,
				&JoinMessage{SenderID: group.memberID},
			); err != nil {
				logger.Errorf("failed to send readiness notification: [%v]", err)
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
		return fmt.Errorf(
			"waiting for notifications timed out after: [%v]", protocolJoinTimeout,
		)
	case context.Canceled:
		return nil
	default:
		return fmt.Errorf("unexpected context error: [%v]", ctx.Err())
	}
}
