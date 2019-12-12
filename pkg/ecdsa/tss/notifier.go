package tss

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/keep-network/keep-tecdsa/pkg/net"
)

const notificationWaitTimeout = 10 * time.Second

type joinNotifier struct {
	wait             *sync.WaitGroup
	broadcastChannel net.BroadcastChannel
}

func newJoinNotifier(group *groupInfo, networkProvider net.Provider) (*joinNotifier, error) {
	broadcastChannel, err := networkProvider.BroadcastChannelFor(group.groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize broadcast channel: [%v]", err)
	}
	if err := broadcastChannel.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
		return &JoinMessage{}
	}); err != nil {
		return nil, fmt.Errorf("failed to register unmarshaler for broadcast channel: [%v]", err)
	}

	joinInChan := make(chan net.Message)
	handleJoinMessage := net.HandleMessageFunc{
		// TODO: This will be set to group ID now, but we may want to add some
		// session ID for concurrent execution.
		Type: fmt.Sprintf(group.groupID),
		Handler: func(netMsg net.Message) error {
			switch netMsg.Payload().(type) {
			case *JoinMessage:
				joinInChan <- netMsg
			}

			return nil
		},
	}
	broadcastChannel.Recv(handleJoinMessage)

	joinWait := &sync.WaitGroup{}
	joinWait.Add(len(group.groupMemberIDs) - 1) // don't wait for self (minus 1)

	go func() {
		waitingForMember := []MemberID{}
		for _, memberID := range group.groupMemberIDs {
			waitingForMember = append(waitingForMember, memberID)
		}

		for {
			select {
			case msg := <-joinInChan:
				senderPublicKey := msg.SenderPublicKey()

				if bytes.Equal(senderPublicKey, []byte(group.memberID)) {
					continue
				}

				for i, memberID := range waitingForMember {
					if bytes.Equal(senderPublicKey, []byte(memberID)) {
						waitingForMember[i] = waitingForMember[len(waitingForMember)-1]
						waitingForMember = waitingForMember[:len(waitingForMember)-1]

						joinWait.Done()
						continue
					}
				}

			}
		}
	}()

	return &joinNotifier{
		wait:             joinWait,
		broadcastChannel: broadcastChannel,
	}, nil
}

func (jn *joinNotifier) notifyReady() error {
	ctx, cancel := context.WithTimeout(context.Background(), notificationWaitTimeout)

	go func() {
		for {
			if err := jn.broadcastChannel.Send(&JoinMessage{}); err != nil {
				logger.Errorf("failed to send readiness notification: [%v]", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		defer cancel()
		jn.wait.Wait()
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
