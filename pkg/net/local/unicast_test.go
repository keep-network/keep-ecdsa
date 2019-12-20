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

	localChannel1, err := netProvider1.UnicastChannelWith(transportID2.String())
	if err != nil {
		t.Fatal(err)
	}
	localChannel2, err := netProvider2.UnicastChannelWith(transportID1.String())
	if err != nil {
		t.Fatal(err)
	}

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

	handlerType := msgToSend.Type()

	deliveredMsgChan := make(chan net.Message)
	handler := net.HandleMessageFunc{
		Type: handlerType,
		Handler: func(msg net.Message) error {
			deliveredMsgChan <- msg
			return nil
		},
	}

	if err := localChannel2.Recv(handler); err != nil {
		t.Fatalf("failed to register receive handler: [%v]", err)
	}

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
		"unregister handler not previously registered": {
			handlersRegistered:   []string{"a", "b", "c"},
			handlersUnregistered: []string{"z"},
			handlersFired:        []string{"a", "b", "c"},
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

			localChannel1, err := netProvider1.UnicastChannelWith(transportID2.String())
			if err != nil {
				t.Fatal(err)
			}
			if err := localChannel1.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
				return &mockMessage{}
			}); err != nil {
				t.Fatalf("failed to register unmarshaler: [%v]", err)
			}

			localChannel2, err := netProvider2.UnicastChannelWith(transportID1.String())
			if err != nil {
				t.Fatal(err)
			}
			if err := localChannel2.RegisterUnmarshaler(func() net.TaggedUnmarshaler {
				return &mockMessage{}
			}); err != nil {
				t.Fatalf("failed to register unmarshaler: [%v]", err)
			}

			handlersFiredMutex := &sync.Mutex{}
			handlersFired := []string{}

			// Register all handlers. If the handler is called, append its
			// type to `handlersFired` slice.
			for _, handlerType := range test.handlersRegistered {
				handlerType := handlerType
				handler := net.HandleMessageFunc{
					Type: handlerType,
					Handler: func(msg net.Message) error {
						handlersFiredMutex.Lock()
						handlersFired = append(handlersFired, handlerType)
						handlersFiredMutex.Unlock()
						return nil
					},
				}

				if err := localChannel2.Recv(handler); err != nil {
					t.Fatalf("failed to register handler: [%v]", err)
				}
			}

			// Unregister specified handlers.
			for _, handlerType := range test.handlersUnregistered {
				if err := localChannel2.UnregisterRecv(handlerType); err != nil {
					t.Fatalf("failed to unregister handler: [%v]", err)
				}
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
