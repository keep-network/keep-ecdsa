package tss

import "encoding/json"

const TSSmessageType = "TSSMessage"

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
	return TSSmessageType
}

// Marshal converts this message to a byte array suitable for network communication.
func (pm *TSSMessage) Marshal() ([]byte, error) {
	return json.Marshal(pm)
}

// Unmarshal converts a byte array produced by Marshal to a message.
func (pm *TSSMessage) Unmarshal(bytes []byte) error {
	var message TSSMessage
	if err := json.Unmarshal(bytes, &message); err != nil {
		return err
	}

	pm.Payload = message.Payload
	pm.SenderID = message.SenderID
	pm.IsBroadcast = message.IsBroadcast

	return nil
}
