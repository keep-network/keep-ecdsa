package local2

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetNonExistingChannel(t *testing.T) {
	manager := newUnicastChannelManager()
	_, err := manager.getChannel(localIdentifier("0x111"))

	expectedErr := fmt.Errorf("no channel with [0x111]")
	if !reflect.DeepEqual(err, expectedErr) {
		t.Errorf(
			"unexpected error\nactual:   [%v]\nexpected: [%v]",
			err,
			expectedErr,
		)
	}
}

func TestAddAndGetChannel(t *testing.T) {
	channel := &unicastChannel{
		senderTransportID:   localIdentifier("0xAAA"),
		receiverTransportID: localIdentifier("0xEEE"),
	}

	channel2 := &unicastChannel{
		senderTransportID:   localIdentifier("0xAAA"),
		receiverTransportID: localIdentifier("0xFFF"),
	}

	manager := newUnicastChannelManager()
	manager.addChannel(channel)
	manager.addChannel(channel2)

	actual, err := manager.getChannel(localIdentifier("0xEEE"))
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(actual, channel) {
		t.Errorf(
			"unexpected channel\nactual:   [%v]\nexpected: [%v]",
			actual,
			channel,
		)
	}

}
