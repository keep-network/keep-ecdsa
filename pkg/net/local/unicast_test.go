package local

import (
	"context"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-tecdsa/pkg/net"
	"github.com/keep-network/keep-tecdsa/pkg/net/internal"
)

func TestRegisterAndFireHandler(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, publicKey1, _ := key.GenerateStaticNetworkKey()
	_, publicKey2, _ := key.GenerateStaticNetworkKey()

	transportID1 := localIdentifierFromNetworkKey(publicKey1)
	transportID2 := localIdentifierFromNetworkKey(publicKey2)

	netProvider1 := LocalProvider(publicKey1)
	netProvider2 := LocalProvider(publicKey2)

	var localChannel1 net.UnicastChannel
	var localChannel2 net.UnicastChannel

	waitInitialized := &sync.WaitGroup{}
	waitInitialized.Add(2)

	go func() {
		var err error
		localChannel1, err = netProvider1.UnicastChannelWith(transportID2)
		if err != nil {
			t.Fatal(err)
		}

		waitInitialized.Done()
	}()

	go func() {
		var err error
		localChannel2, err = netProvider2.UnicastChannelWith(transportID1)
		if err != nil {
			t.Fatal(err)
		}

		waitInitialized.Done()
	}()

	waitInitialized.Wait()

	if err := localChannel1.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
		return &mockMessage{}
	}); err != nil {
		t.Fatalf("failed to register unmarshaler: [%v]", err)
	}
	if err := localChannel2.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
		return &mockMessage{}
	}); err != nil {
		t.Fatalf("failed to register unmarshaler: [%v]", err)
	}

	msgToSend := &mockMessage{}

	deliveredMsgChan := make(chan net.Message)
	handler := func(msg net.Message) {
		deliveredMsgChan <- msg
	}

	localChannel2.Recv(ctx, handler)

	expectedDeliveredMessage := internal.BasicMessage(
		localIdentifier(transportID1),
		msgToSend,
		msgToSend.Type(),
		key.Marshal(publicKey1),
	)

	if err := localChannel1.Send(msgToSend); err != nil {
		t.Fatalf("failed to send message: [%v]", err)
	}

	select {
	case deliveredMsg := <-deliveredMsgChan:
		if !reflect.DeepEqual(deliveredMsg, expectedDeliveredMessage) {
			t.Errorf("invalid delivered message\nexpected: %+v\nactual:   %+v\n", expectedDeliveredMessage, deliveredMsg)
		}
	case <-ctx.Done():
		t.Errorf("expected handler not called")
	}
}

func TestChannelFor(t *testing.T) {
	_, publicKey1, _ := key.GenerateStaticNetworkKey()
	_, publicKey2, _ := key.GenerateStaticNetworkKey()
	_, publicKey3, _ := key.GenerateStaticNetworkKey()

	transportID1 := localIdentifierFromNetworkKey(publicKey1)
	transportID2 := localIdentifierFromNetworkKey(publicKey2)
	transportID3 := localIdentifierFromNetworkKey(publicKey3)

	netProvider1 := LocalProvider(publicKey1)
	netProvider2 := LocalProvider(publicKey2)
	netProvider3 := LocalProvider(publicKey3)

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(4)

	go func() {
		netProvider1.UnicastChannelWith(transportID2)
		waitGroup.Done()
	}()

	go func() {
		netProvider1.UnicastChannelWith(transportID3)
		waitGroup.Done()
	}()

	go func() {
		netProvider2.UnicastChannelWith(transportID1)
		waitGroup.Done()
	}()

	go func() {
		netProvider3.UnicastChannelWith(transportID1)
		waitGroup.Done()
	}()

	waitGroup.Wait()
}

func TestSendMessageBeforeHandlerRegistration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, publicKey1, _ := key.GenerateStaticNetworkKey()
	_, publicKey2, _ := key.GenerateStaticNetworkKey()

	transportID1 := localIdentifierFromNetworkKey(publicKey1)
	transportID2 := localIdentifierFromNetworkKey(publicKey2)

	netProvider1 := LocalProvider(publicKey1)
	netProvider2 := LocalProvider(publicKey2)

	var localChannel1 net.UnicastChannel
	var localChannel2 net.UnicastChannel

	waitInitialized := &sync.WaitGroup{}
	waitInitialized.Add(2)

	go func() {
		var err error
		localChannel1, err = netProvider1.UnicastChannelWith(transportID2)
		if err != nil {
			t.Fatal(err)
		}

		waitInitialized.Done()
	}()

	go func() {
		var err error
		localChannel2, err = netProvider2.UnicastChannelWith(transportID1)
		if err != nil {
			t.Fatal(err)
		}

		waitInitialized.Done()
	}()

	waitInitialized.Wait()

	if err := localChannel1.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
		return &mockMessage{}
	}); err != nil {
		t.Fatalf("failed to register unmarshaler: [%v]", err)
	}
	if err := localChannel2.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
		return &mockMessage{}
	}); err != nil {
		t.Fatalf("failed to register unmarshaler: [%v]", err)
	}

	msgToSend := &mockMessage{}

	expectedDeliveredMessage := internal.BasicMessage(
		localIdentifier(transportID1),
		msgToSend,
		msgToSend.Type(),
		key.Marshal(publicKey1),
	)

	go func() {
		if err := localChannel1.Send(msgToSend); err != nil {
			t.Fatalf("failed to send message: [%v]", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	deliveredMsgChan := make(chan net.Message)
	handler := func(msg net.Message) {
		deliveredMsgChan <- msg
	}

	localChannel2.Recv(ctx, handler)

	select {
	case deliveredMsg := <-deliveredMsgChan:
		if !reflect.DeepEqual(deliveredMsg, expectedDeliveredMessage) {
			t.Errorf("invalid delivered message\nexpected: %+v\nactual:   %+v\n", expectedDeliveredMessage, deliveredMsg)
		}
	case <-ctx.Done():
		t.Errorf("expected handler not called")
	}
}

func TestUnregisterHandler(t *testing.T) {
	tests := map[string]struct {
		handlersRegistered   []string
		handlersUnregistered []string
		handlersFired        []string
	}{
		"unregister the first registered handler": {
			handlersRegistered:   []string{"a", "b", "c"},
			handlersUnregistered: []string{"a"},
			handlersFired:        []string{"b", "c"},
		},
		"unregister the last registered handler": {
			handlersRegistered:   []string{"a", "b", "c"},
			handlersUnregistered: []string{"c"},
			handlersFired:        []string{"a", "b"},
		},
		"unregister handler registered in the middle": {
			handlersRegistered:   []string{"a", "b", "c"},
			handlersUnregistered: []string{"b"},
			handlersFired:        []string{"a", "c"},
		},
		"unregister various handlers": {
			handlersRegistered:   []string{"a", "b", "c", "d", "e", "f", "g"},
			handlersUnregistered: []string{"a", "c", "f", "g"},
			handlersFired:        []string{"b", "d", "e"},
		},
		"unregister all handlers": {
			handlersRegistered:   []string{"a", "b", "c"},
			handlersUnregistered: []string{"a", "b", "c"},
			handlersFired:        []string{},
		},
	}

	for testName, test := range tests {
		test := test
		t.Run(testName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			_, publicKey1, _ := key.GenerateStaticNetworkKey()
			_, publicKey2, _ := key.GenerateStaticNetworkKey()

			transportID1 := localIdentifierFromNetworkKey(publicKey1)
			transportID2 := localIdentifierFromNetworkKey(publicKey2)

			netProvider1 := LocalProvider(publicKey1)
			netProvider2 := LocalProvider(publicKey2)

			var localChannel1 net.UnicastChannel
			var localChannel2 net.UnicastChannel

			waitInitialized := &sync.WaitGroup{}
			waitInitialized.Add(2)

			go func() {
				var err error
				localChannel1, err = netProvider1.UnicastChannelWith(transportID2)
				if err != nil {
					t.Fatal(err)
				}

				if err := localChannel1.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
					return &mockMessage{}
				}); err != nil {
					t.Fatalf("failed to register unmarshaler: [%v]", err)
				}

				waitInitialized.Done()
			}()

			go func() {
				var err error
				localChannel2, err = netProvider2.UnicastChannelWith(transportID1)
				if err != nil {
					t.Fatal(err)
				}

				if err := localChannel2.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
					return &mockMessage{}
				}); err != nil {
					t.Fatalf("failed to register unmarshaler: [%v]", err)
				}

				waitInitialized.Done()
			}()

			waitInitialized.Wait()

			handlersFiredMutex := &sync.Mutex{}
			handlersFired := []string{}

			handlerCancellations := map[string]context.CancelFunc{}

			// Register all handlers. If the handler is called, append its
			// type to `handlersFired` slice.
			for _, handlerName := range test.handlersRegistered {
				handlerName := handlerName

				handlerCtx, cancel := context.WithCancel(ctx)
				defer cancel()

				handlerCancellations[handlerName] = cancel

				localChannel2.Recv(handlerCtx, func(msg net.Message) {
					handlersFiredMutex.Lock()
					handlersFired = append(handlersFired, handlerName)
					handlersFiredMutex.Unlock()
				})
			}

			// Unregister specified handlers.
			for _, handlerName := range test.handlersUnregistered {
				handlerCancellations[handlerName]()
			}

			// Send a message, all handlers should be called.
			if err := localChannel1.Send(&mockMessage{}); err != nil {
				t.Fatalf("failed to send message: [%v]", err)
			}

			// Handlers are fired asynchronously; wait for them.
			<-ctx.Done()

			sort.Strings(handlersFired)
			if !reflect.DeepEqual(test.handlersFired, handlersFired) {
				t.Errorf(
					"Unexpected handlers fired\nExpected: %v\nActual:   %v\n",
					test.handlersFired,
					handlersFired,
				)
			}
		})
	}
}

type mockMessage struct{}

func (mm *mockMessage) Type() string {
	return "mock_message"
}

func (mm *mockMessage) Marshal() ([]byte, error) {
	return []byte("some mocked bytes"), nil
}

func (mm *mockMessage) Unmarshal(bytes []byte) error {
	return nil
}
