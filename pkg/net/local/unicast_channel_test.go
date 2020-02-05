package local

import (
	"context"
	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-tecdsa/pkg/net"
	"reflect"
	"testing"
	"time"
)

func TestReceiveUnicastMessage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	peer1ID := localIdentifier("peer-0x1231AA")
	_, peer1StaticKey, _ := key.GenerateStaticNetworkKey()

	peer2ID := localIdentifier("peer-0xAEA712")

	unicastChannel := newUnicastChannel(peer1ID, peer1StaticKey, peer2ID)
	unicastChannel.SetUnmarshaler(func() net.TaggedUnmarshaler {
		return &mockMessage{}
	})

	received := make(chan net.Message)
	unicastChannel.Recv(ctx, func(msg net.Message) {
		received <- msg
	})

	received2 := make(chan net.Message)
	unicastChannel.Recv(ctx2, func(msg net.Message) {
		received2 <- msg
	})

	message := &mockMessage{"hello"}
	marshaled, err := message.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	unicastChannel.receiveMessage(marshaled, message.Type())

	select {
	case <-ctx.Done():
		t.Fatal("expected message not received")
	case actual := <-received:
		if !reflect.DeepEqual(actual.Payload(), message) {
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
		if !reflect.DeepEqual(actual.Payload(), message) {
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

	peer1ID := localIdentifier("peer-0xAAEF12")
	_, peer1StaticKey, _ := key.GenerateStaticNetworkKey()

	peer2ID := localIdentifier("peer-0x121211")

	unicastChannel := newUnicastChannel(peer1ID, peer1StaticKey, peer2ID)
	unicastChannel.SetUnmarshaler(func() net.TaggedUnmarshaler {
		return &mockMessage{}
	})

	received := make(chan net.Message)
	unicastChannel.Recv(ctx, func(msg net.Message) {
		received <- msg
	})

	received2 := make(chan net.Message)
	unicastChannel.Recv(ctx2, func(msg net.Message) {
		received2 <- msg
	})

	cancel() // cancel the first context

	message := &mockMessage{"hello"}
	marshaled, err := message.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	unicastChannel.receiveMessage(marshaled, message.Type())

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
