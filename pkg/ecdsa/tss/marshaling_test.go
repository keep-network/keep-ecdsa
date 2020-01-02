package tss

import (
	"fmt"
	"reflect"
	"testing"

	testdata "github.com/keep-network/keep-tecdsa/internal/testdata/tss"
	"github.com/keep-network/keep-tecdsa/pkg/utils/pbutils"
)

func TestSignerMarshalling(t *testing.T) {
	groupSize := 5
	dishonestThreshold := 4
	signerIndex := 2

	testData, err := testdata.LoadKeygenTestFixtures(groupSize)
	if err != nil {
		t.Fatalf("failed to load test data: [%v]", err)
	}

	groupMembersIDs := make([]MemberID, groupSize)

	for i := range groupMembersIDs {
		groupMembersIDs[i] = MemberID([]byte(fmt.Sprintf("member-%d", i)))
	}

	signer := &ThresholdSigner{
		groupInfo: &groupInfo{
			groupID:            "test-group-id-1",
			memberID:           groupMembersIDs[signerIndex],
			groupMemberIDs:     groupMembersIDs,
			dishonestThreshold: dishonestThreshold,
		},
		thresholdKey: ThresholdKey(testData[signerIndex]),
	}

	unmarshaled := &ThresholdSigner{}

	if err := pbutils.RoundTrip(signer, unmarshaled); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(signer, unmarshaled) {
		t.Fatalf(
			"unexpected content of unmarshaled signer\nexpected: [%+v]\nactual:   [%+v]\n",
			signer,
			unmarshaled,
		)
	}
}

func TestThresholdKeyMarshalling(t *testing.T) {
	testData, err := testdata.LoadKeygenTestFixtures(1)
	if err != nil {
		t.Fatalf("failed to load test data: [%v]", err)
	}

	key := ThresholdKey(testData[0])

	unmarshaled := &ThresholdKey{}

	if err := pbutils.RoundTrip(&key, unmarshaled); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(&key, unmarshaled) {
		t.Fatalf(
			"unexpected content of unmarshaled signer\nexpected: [%+v]\nactual:   [%+v]\n",
			&key,
			unmarshaled,
		)
	}
}

func TestTSSProtocolMessageMarshalling(t *testing.T) {
	msg := &TSSProtocolMessage{
		SenderID:    MemberID([]byte("member-1")),
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
		SenderID: MemberID([]byte("member-1")),
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
