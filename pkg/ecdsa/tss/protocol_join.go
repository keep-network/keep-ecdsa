package tss

import (
	"context"
	"fmt"
	"time"

	"github.com/keep-network/keep-tecdsa/pkg/net"
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
func joinProtocol(group *groupInfo, networkProvider net.Provider) error {
	ctx, cancel := context.WithTimeout(context.Background(), protocolJoinTimeout)
	defer cancel()

	broadcastChannel, err := networkProvider.BroadcastChannelFor(group.groupID)
	if err != nil {
		return fmt.Errorf("failed to initialize broadcast channel: [%v]", err)
	}
	// TODO: We ignore the error for the case when the unmarshaler is already
	// registered. We should rework the `RegisterUnmarshaler` to not return
	// an error in such case.
	broadcastChannel.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
		return &JoinMessage{}
	})

	handleType := fmt.Sprintf("%s-%s", group.groupID, time.Now())
	joinInChan := make(chan *JoinMessage, len(group.groupMemberIDs))
	handleJoinMessage := net.HandleMessageFunc{
		Type: handleType,
		Handler: func(netMsg net.Message) error {
			switch msg := netMsg.Payload().(type) {
			case *JoinMessage:
				if msg.SenderID != group.memberID {
					joinInChan <- msg
				}
			}

			return nil
		},
	}
	broadcastChannel.Recv(handleJoinMessage)
	defer broadcastChannel.UnregisterRecv(handleType)

	go func() {
		readyMembers := make(map[MemberID]bool)

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-joinInChan:
				for _, memberID := range group.groupMemberIDs {
					if msg.SenderID == memberID {
						readyMembers[msg.SenderID] = true
						break
					}
				}

				if len(readyMembers) == (len(group.groupMemberIDs) - 1) { // don't wait for self (minus 1)
					cancel()
				}
			}
		}
	}()

	go func() {
		sendMessage := func() {
			if err := broadcastChannel.Send(
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
