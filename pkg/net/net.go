package net

import (
	"github.com/binance-chain/tss-lib/tss"
)

// NetworkChannel represents a channel for members to exchange messages.
// TODO: Wrap `tss.Message` type with internal message type.
// TODO: Reuse keep-core?
type NetworkChannel interface {
	// TODO: Send function has to support broadcast and unicast or we can provide
	// separate functions for point-to-point and broadcast communication.
	// ```
	// destinations := message.GetTo()
	// if destinations == nil {
	//   // broadcast
	// } else {
	//   for _, dest := range destinations {
	//    // unicast
	//   }
	// }
	// ```
	Send(message tss.Message) error
	// Receive registers handler function to be executed on message receive.
	Receive(handler func(message tss.Message) error)
}
