package local2

import (
	"fmt"
	"sync"

	"github.com/keep-network/keep-tecdsa/pkg/net"
)

type unicastChannelManager struct {
	structMutex *sync.RWMutex
	channels    map[net.TransportIdentifier]*unicastChannel
}

func newUnicastChannelManager() *unicastChannelManager {
	return &unicastChannelManager {
		structMutex: &sync.RWMutex{},
		channels: make(map[net.TransportIdentifier]*unicastChannel),
	}
}

func (ucm *unicastChannelManager) getChannel(
	receiver net.TransportIdentifier,
) (*unicastChannel, error) {
	ucm.structMutex.RLock()
	defer ucm.structMutex.RUnlock()

	channel, ok := ucm.channels[receiver]
	if !ok {
		return nil, fmt.Errorf("no channel with [%v]", receiver)
	}

	return channel, nil
}

func (ucm *unicastChannelManager) addChannel(channel *unicastChannel) {
	ucm.structMutex.Lock()
	defer ucm.structMutex.Unlock()

	ucm.channels[channel.receiverTransportID] = channel
}
