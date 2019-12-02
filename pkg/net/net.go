package net

// NetworkChannel represents a channel for members to exchange messages.
// TODO: Wrap `tss.Message` type with internal message type.
// TODO: Reuse keep-core?
type NetworkChannel interface {
	// TODO: Send function has to support broadcast and unicast or we can provide
	// separate functions for point-to-point and broadcast communication.
	Send(message Message) error
	// Receive registers handler function to be executed on message receive.
	Receive(handler func(message Message) error)
}

type Message struct {
	From        []byte // unique sender key
	To          []byte // unique destination key, required for unicast
	IsBroadcast bool
	Payload     []byte
}
