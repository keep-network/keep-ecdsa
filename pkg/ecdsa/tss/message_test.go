package tss

import (
	"reflect"
	"testing"

	"github.com/keep-network/keep-tecdsa/pkg/utils/pbutils"
)

func TestTSSProtocolMessageMarshalling(t *testing.T) {
	msg := &TSSProtocolMessage{
		SenderID:    MemberID("member-1"),
		Payload:     []byte("very important message"),
		IsBroadcast: true,
	}

	unmarshaled := &TSSProtocolMessage{}

	if err := pbutils.RoundTrip(msg, unmarshaled); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(msg, unmarshaled) {
		t.Fatalf(
			"unexpected content of unmarshaled message\nexpected: [%+v]\nactual:   [%+v]\n",
			msg,
			unmarshaled,
		)
	}
}

func TestJoinMessageMarshalling(t *testing.T) {
	msg := &JoinMessage{
		SenderID: MemberID("member-1"),
	}

	unmarshaled := &JoinMessage{}

	if err := pbutils.RoundTrip(msg, unmarshaled); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(msg, unmarshaled) {
		t.Fatalf(
			"unexpected content of unmarshaled message\nexpected: [%+v]\nactual:   [%+v]\n",
			msg,
			unmarshaled,
		)
	}
}
