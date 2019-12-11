package tss

import (
	"encoding/json"
	"fmt"
)

// TSSMessage is a network message used to transport messages in TSS protocol
// execution.
type TSSMessage struct {
	SenderID    []byte
	Payload     []byte
	IsBroadcast bool
}

// Type returns a string type of the `TSSMessage` so that it conforms to
// `net.Message` interface.
func (m *TSSMessage) Type() string {
	return fmt.Sprintf("%T", m)
}

// Marshal converts this message to a byte array suitable for network communication.
func (m *TSSMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// Unmarshal converts a byte array produced by Marshal to a message.
func (m *TSSMessage) Unmarshal(bytes []byte) error {
	var message TSSMessage
	if err := json.Unmarshal(bytes, &message); err != nil {
		return err
	}

	m.Payload = message.Payload
	m.SenderID = message.SenderID
	m.IsBroadcast = message.IsBroadcast

	return nil
}
