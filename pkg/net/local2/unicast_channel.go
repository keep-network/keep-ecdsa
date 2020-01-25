package local2

import "github.com/keep-network/keep-tecdsa/pkg/net"

type unicastChannel struct {
	senderTransportID   net.TransportIdentifier
	receiverTransportID net.TransportIdentifier
}
