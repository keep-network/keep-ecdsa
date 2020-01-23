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
func joinProtocol(parentCtx context.Context, group *groupInfo, networkProvider net.Provider) (map[string][]byte, error) {
	ctx, cancel := context.WithTimeout(parentCtx, protocolJoinTimeout)
	defer cancel()

	broadcastChannel, err := networkProvider.BroadcastChannelFor(group.groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize broadcast channel: [%v]", err)
	}
	// TODO: We ignore the error for the case when the unmarshaler is already
	// registered. We should rework the `RegisterUnmarshaler` to not return
	// an error in such case.
	broadcastChannel.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
		return &JoinMessage{}
	})

	joinInChan := make(chan net.Message, len(group.groupMemberIDs))
	handleJoinMessage := func(netMsg net.Message) {
		switch netMsg.Payload().(type) {
		case *JoinMessage:
			joinInChan <- netMsg
		}
	}
	broadcastChannel.Recv(ctx, handleJoinMessage)

	membersPublicKeys := make(map[string][]byte) // TODO: Change key type to MemberID

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case netMsg := <-joinInChan:
				for _, memberID := range group.groupMemberIDs {
					senderID := netMsg.Payload().(*JoinMessage).SenderID

					if senderID.Equal(memberID) {
						// TODO: Validate sender public key against SenderID

						membersPublicKeys[senderID.String()] = netMsg.SenderPublicKey()
						break
					}
				}

				if len(membersPublicKeys) == len(group.groupMemberIDs) {
					cancel()
				}
			}
		}
	}()

	go func() {
		sendMessage := func() {
			if err := broadcastChannel.Send(
				ctx,
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
		return nil, fmt.Errorf(
			"waiting for notifications timed out after: [%v]", protocolJoinTimeout,
		)
	case context.Canceled:
		return membersPublicKeys, nil
	default:
		return nil, fmt.Errorf("unexpected context error: [%v]", ctx.Err())
	}
}
