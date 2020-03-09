package tss

import (
	"context"
	"fmt"
	"time"

	"github.com/keep-network/keep-core/pkg/net"
)

// protocolReadyTimeout defines a period within which the member sends and receives
// notifications from peer members about their readiness to begin the protocol
// execution. If the time limit is reached the ready protocol stage fails.
const protocolReadyTimeout = 2 * time.Minute

// readyProtocol exchanges messages with peer members about readiness to start
// the protocol execution. The member keeps sending the message in intervals
// until they receive messages from all peer members. Function exits without an
// error if messages were received from all peer members. If the timeout is
// reached before receiving messages from all peer members the function returns
// an error.
func readyProtocol(
	parentCtx context.Context,
	group *groupInfo,
	broadcastChannel net.BroadcastChannel,
) error {
	logger.Infof("signalling readiness")

	ctx, cancel := context.WithTimeout(parentCtx, protocolReadyTimeout)
	defer cancel()

	readyInChan := make(chan *ReadyMessage, len(group.groupMemberIDs))
	handleReadyMessage := func(netMsg net.Message) {
		switch msg := netMsg.Payload().(type) {
		case *ReadyMessage:
			readyInChan <- msg
		}
	}
	broadcastChannel.Recv(ctx, handleReadyMessage)

	go func() {
		readyMembers := make(map[string]bool)

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-readyInChan:
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
				&ReadyMessage{SenderID: group.memberID},
			); err != nil {
				logger.Errorf("failed to send readiness notification: [%v]", err)
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
		return fmt.Errorf(
			"waiting for readiness timed out after: [%v]", protocolReadyTimeout,
		)
	case context.Canceled:
		logger.Infof("successfully signalled readiness")

		return nil
	default:
		return fmt.Errorf("unexpected context error: [%v]", ctx.Err())
	}
}
