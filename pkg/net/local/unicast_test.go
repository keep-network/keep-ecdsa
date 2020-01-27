package local

import (
	"context"
	"testing"
	"time"

	"github.com/keep-network/keep-tecdsa/pkg/net"
)

func TestNewChannelNotification(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	peer1PubKey := []byte("0xAFF123")
	peer2PubKey := []byte("0xFF1231")

	peer1Provider := unicastConnectWithKey(peer1PubKey)
	peer2Provider := unicastConnectWithKey(peer2PubKey)

	peer1NewChannelNotificationCount := 0
	peer1Provider.OnUnicastChannelOpened(ctx, func(channel net.UnicastChannel) {
		peer1NewChannelNotificationCount++
	})

	peer2NewChannelNotificationCount := 0
	peer2Provider.OnUnicastChannelOpened(ctx, func(channel net.UnicastChannel) {
		peer2NewChannelNotificationCount++
	})

	remotePeerID := peer1Provider.TransportIDFromPublicKey(peer2PubKey)
	peer1Provider.UnicastChannelWith(remotePeerID)

	<-ctx.Done() // give some time for notifications...

	if peer1NewChannelNotificationCount != 0 {
		t.Errorf(
			"expected no notifications, has [%v]",
			peer1NewChannelNotificationCount,
		)
	}
	if peer2NewChannelNotificationCount != 1 {
		t.Errorf(
			"expected [1] notification, has [%v]",
			peer2NewChannelNotificationCount,
		)
	}
}

func TestExistingChannelNotification(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	peer1PubKey := []byte("0x717211")
	peer2PubKey := []byte("0xFA1231")

	peer1Provider := unicastConnectWithKey(peer1PubKey)
	peer2Provider := unicastConnectWithKey(peer2PubKey)

	newChannelNotificationCount := 0
	peer2Provider.OnUnicastChannelOpened(ctx, func(channel net.UnicastChannel) {
		newChannelNotificationCount++
	})

	remotePeerID := peer1Provider.TransportIDFromPublicKey(peer2PubKey)
	peer1Provider.UnicastChannelWith(remotePeerID)
	peer1Provider.UnicastChannelWith(remotePeerID)

	<-ctx.Done() // give some time for notifications...

	if newChannelNotificationCount != 1 {
		t.Errorf(
			"expected [1] notification, has [%v]",
			newChannelNotificationCount,
		)
	}
}

func TestSendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	//
	// Prepare communication channel between peer1 and peer2
	//
	peer1PubKey := []byte("0xFF12315")
	peer2PubKey := []byte("0xAEAA121")

	peer1Provider := unicastConnectWithKey(peer1PubKey)
	peer2Provider := unicastConnectWithKey(peer2PubKey)

	remotePeer1ID := peer2Provider.TransportIDFromPublicKey(peer1PubKey)
	remotePeer2ID := peer1Provider.TransportIDFromPublicKey(peer2PubKey)

	channel1, err := peer1Provider.UnicastChannelWith(remotePeer2ID)
	if err != nil {
		t.Fatal(err)
	}
	channel2, err := peer2Provider.UnicastChannelWith(remotePeer1ID)
	if err != nil {
		t.Fatal(err)
	}

	channel1.SetUnmarshaler(func() net.TaggedUnmarshaler {
		return &mockMessage{}
	})
	channel2.SetUnmarshaler(func() net.TaggedUnmarshaler {
		return &mockMessage{}
	})

	peer1Received := make(chan net.Message)
	peer2Received := make(chan net.Message)

	channel1.Recv(ctx, func(msg net.Message) {
		peer1Received <- msg
	})
	channel2.Recv(ctx, func(msg net.Message) {
		peer2Received <- msg
	})

	//
	// peer1 sends a message to peer2
	// make sure peer2 receives it
	//

	channel1Message := &mockMessage{"yolo1"}
	err = channel1.Send(channel1Message)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case msg := <-peer2Received:
		switch message := msg.Payload().(type) {
		case *mockMessage:
			if message.content != channel1Message.content {
				t.Fatalf(
					"unexpected message content\nactual:   [%v]\nexpected: [%v]",
					message.content,
					channel1Message.content,
				)
			}
		default:
			t.Fatal("unexpected message type")
		}

	case <-peer1Received:
		t.Fatal("peer 1 should not receive this message")
	case <-ctx.Done():
		t.Fatal("expected message not arrived to peer 2")
	}

	//
	// peer2 sends a message to peer1
	// make sure peer1 receives it
	//

	channel2Message := &mockMessage{"yolo2"}
	err = channel2.Send(channel2Message)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case msg := <-peer1Received:
		switch message := msg.Payload().(type) {
		case *mockMessage:
			if message.content != channel2Message.content {
				t.Fatalf(
					"unexpected message content\nactual:   [%v]\nexpected: [%v]",
					message.content,
					channel2Message.content,
				)
			}
		default:
			t.Fatal("unexpected message type")
		}
	case <-peer2Received:
		t.Fatal("peer 2 should not receive this message")
	case <-ctx.Done():
		t.Fatal("expected message not arrived")
	}
}

func TestTalkToSelf(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	//
	// Prepare self-communication channel (e.g. two goroutines)
	//
	peerPubKey := []byte("0x177171")
	peerProvider := unicastConnectWithKey(peerPubKey)

	peerTransportID := peerProvider.TransportIDFromPublicKey(peerPubKey)

	channel1, err := peerProvider.UnicastChannelWith(peerTransportID)
	if err != nil {
		t.Fatal(err)
	}
	channel2, err := peerProvider.UnicastChannelWith(peerTransportID)
	if err != nil {
		t.Fatal(err)
	}

	channel1.SetUnmarshaler(func() net.TaggedUnmarshaler {
		return &mockMessage{}
	})

	chan1Received := make(chan net.Message)
	chan2Received := make(chan net.Message)

	channel1.Recv(ctx, func(msg net.Message) {
		chan1Received <- msg
	})
	channel2.Recv(ctx, func(msg net.Message) {
		chan2Received <- msg
	})

	//
	// send message to self via the first channel
	// both handlers receive it
	//

	err = channel1.Send(&mockMessage{"yolo1"})
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-chan1Received: // ok
	case <-ctx.Done():
		t.Fatal("expected message not arrived")
	}

	select {
	case <-chan2Received: // ok
	case <-ctx.Done():
		t.Fatal("expected message not arrived")
	}

	//
	// send message to self via the second channel
	// again, both handlers should receive it
	//

	err = channel2.Send(&mockMessage{"yolo2"})
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-chan1Received: // ok
	case <-ctx.Done():
		t.Fatal("expected message not arrived")
	}

	select {
	case <-chan2Received: // ok
	case <-ctx.Done():
		t.Fatal("expected message not arrived")
	}
}

type mockMessage struct {
	content string
}

func (mm *mockMessage) Type() string {
	return "mock_message"
}

func (mm *mockMessage) Marshal() ([]byte, error) {
	return []byte(mm.content), nil
}

func (mm *mockMessage) Unmarshal(bytes []byte) error {
	mm.content = string(bytes)
	return nil
}
