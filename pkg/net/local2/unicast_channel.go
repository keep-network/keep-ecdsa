package local2

import (
	"context"
	"fmt"
	"sync"

	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-tecdsa/pkg/net"
)

type unicastChannel struct {
	structMutex *sync.RWMutex

	senderTransportID net.TransportIdentifier
	senderPublicKey   *key.NetworkPublic

	receiverTransportID net.TransportIdentifier

	messageReceivers   []*unicastChannelRecv
	unmarshalersByType map[string]func() net.TaggedUnmarshaler
}

type unicastChannelRecv struct {
	ctx      context.Context
	handleFn func(message net.Message)
}

func newUnicastChannel(
	senderTransportID net.TransportIdentifier,
	senderPublicKey *key.NetworkPublic,
	receiverTransportID net.TransportIdentifier,
) *unicastChannel {
	return &unicastChannel{
		structMutex:         &sync.RWMutex{},
		senderTransportID:   senderTransportID,
		senderPublicKey:     senderPublicKey,
		receiverTransportID: receiverTransportID,
		messageReceivers:    make([]*unicastChannelRecv, 0),
		unmarshalersByType:  make(map[string]func() net.TaggedUnmarshaler),
	}
}

func (uc *unicastChannel) Send(message net.TaggedMarshaler) error {
	panic("not implemented yet")
}

func (uc *unicastChannel) Recv(
	ctx context.Context,
	handler func(message net.Message),
) {
	uc.structMutex.Lock()
	defer uc.structMutex.Unlock()

	uc.messageReceivers = append(
		uc.messageReceivers,
		&unicastChannelRecv{ctx: ctx, handleFn: handler},
	)
}

func (uc *unicastChannel) RegisterUnmarshaler(
	unmarshaler func() net.TaggedUnmarshaler,
) error {
	uc.structMutex.Lock()
	defer uc.structMutex.Unlock()

	tpe := unmarshaler().Type()

	_, exists := uc.unmarshalersByType[tpe]
	if exists {
		return fmt.Errorf("type %s already has an associated unmarshaler", tpe)
	}

	uc.unmarshalersByType[tpe] = unmarshaler
	return nil
}

func (uc *unicastChannel) receiveMessage(message net.Message) {
	uc.structMutex.Lock()
	defer uc.structMutex.Unlock()

	i := 0
	for _, receiver := range uc.messageReceivers {
		// check if still active...
		if receiver.ctx.Err() == nil {
			// still active, should remain in the slice
			uc.messageReceivers[i] = receiver
			i++

			// firing handler asynchronously to
			// do not block the loop
			go receiver.handleFn(message)
		}
	}

	// cleaning up those no longer active
	uc.messageReceivers = uc.messageReceivers[:i]
}
