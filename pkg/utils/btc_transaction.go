package utils

import (
	"bytes"
	"fmt"

	"github.com/btcsuite/btcd/wire"
)

// DeserializeTransaction decodes a transaction from bitcoin hexadecimal format
// to a bitcoin message.
func DeserializeTransaction(transactionHex []byte) (*wire.MsgTx, error) {
	msgTx := wire.NewMsgTx(wire.TxVersion)

	reader := bytes.NewReader(transactionHex)
	err := msgTx.Deserialize(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize transaction: [%v]", err)
	}

	return msgTx, nil
}

// SerializeTransaction encodes a bitcoin transaction message to a hexadecimal
// format.
func SerializeTransaction(msgTx *wire.MsgTx) ([]byte, error) {
	var buffer bytes.Buffer

	err := msgTx.Serialize(&buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: [%v]", err)
	}

	return buffer.Bytes(), nil
}
