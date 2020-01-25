package local2

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-tecdsa/pkg/net"
	"github.com/keep-network/keep-tecdsa/pkg/net/internal"
)

func TestReceiveUnicastMessage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	peer1ID := localIdentifier("peer-1")
	_, peer1PubKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	peer2ID := localIdentifier("peer-2")
	_, peer2PubKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	unicastChannel := newUnicastChannel(peer1ID, peer1PubKey, peer2ID)

	received := make(chan net.Message)
	unicastChannel.Recv(ctx, func(msg net.Message) {
		received <- msg
	})

	received2 := make(chan net.Message)
	unicastChannel.Recv(ctx2, func(msg net.Message) {
		received2 <- msg
	})

	message := internal.BasicMessage(peer2ID, "payload", "type", peer2PubKey.X.Bytes())
	unicastChannel.receiveMessage(message)

	select {
	case <-ctx.Done():
		t.Fatal("expected message not received")
	case actual := <-received:
		if !reflect.DeepEqual(actual, message) {
			t.Errorf(
				"unexpected message\nactual:   [%v]\nexpected: [%v]",
				actual,
				message,
			)
		}
	}

	select {
	case <-ctx.Done():
		t.Fatal("expected message not received")
	case actual := <-received2:
		if !reflect.DeepEqual(actual, message) {
			t.Errorf(
				"unexpected message\nactual:   [%v]\nexpected: [%v]",
				actual,
				message,
			)
		}
	}
}

func TestTimedOutHandlerNotReceiveUnicastMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx2, cancel2 := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel2()

	peer1ID := localIdentifier("peer-1")
	_, peer1PubKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	peer2ID := localIdentifier("peer-2")
	_, peer2PubKey, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	unicastChannel := newUnicastChannel(peer1ID, peer1PubKey, peer2ID)

	received := make(chan net.Message)
	unicastChannel.Recv(ctx, func(msg net.Message) {
		received <- msg
	})

	received2 := make(chan net.Message)
	unicastChannel.Recv(ctx2, func(msg net.Message) {
		received2 <- msg
	})

	cancel() // cancel the first context

	message := internal.BasicMessage(peer2ID, "payload", "type", peer2PubKey.X.Bytes())
	unicastChannel.receiveMessage(message)

	select {
	case <-received:
		t.Fatal("receiver should not be called")
	default:
		// ok, should not receive
	}

	select {
	case <-ctx2.Done():
		t.Fatal("expected message not received")
	case <-received2:
		// ok, should receive
	}
}
