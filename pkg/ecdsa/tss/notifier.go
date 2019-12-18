package tss

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/keep-network/keep-tecdsa/pkg/net"
)

const notificationWaitTimeout = 120 * time.Second

func joinProtocol(group *groupInfo, networkProvider net.Provider) error {
	ctx, cancel := context.WithTimeout(context.Background(), notificationWaitTimeout)
	defer cancel()

	broadcastChannel, err := networkProvider.BroadcastChannelFor(group.groupID)
	if err != nil {
		return fmt.Errorf("failed to initialize broadcast channel: [%v]", err)
	}
	if err := broadcastChannel.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
		return &JoinMessage{}
	}); err != nil {
		return fmt.Errorf("failed to register unmarshaler for broadcast channel: [%v]", err)
	}

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

	joinWait := &sync.WaitGroup{}
	joinWait.Add(len(group.groupMemberIDs) - 1) // don't wait for self (minus 1)

	go func() {
		waitingForMember := []MemberID{}
		for _, memberID := range group.groupMemberIDs {
			waitingForMember = append(waitingForMember, memberID)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-joinInChan:

				for i, memberID := range waitingForMember {
					if msg.SenderID == memberID {
						waitingForMember[i] = waitingForMember[len(waitingForMember)-1]
						waitingForMember = waitingForMember[:len(waitingForMember)-1]

						joinWait.Done()
						continue
					}
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
				sendMessage()
				return
			default:
				sendMessage()
				time.Sleep(1 * time.Second)
			}
		}
	}()

	go func() {
		joinWait.Wait()
		cancel()
	}()

	<-ctx.Done()
	switch ctx.Err() {
	case context.DeadlineExceeded:
		return fmt.Errorf(
			"waiting for notifications timed out after: [%v]", notificationWaitTimeout,
		)
	case context.Canceled:
		return nil
	default:
		return fmt.Errorf("unexpected context error: [%v]", ctx.Err())
	}
}
