package tss

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/keep-network/keep-tecdsa/internal/testdata"
	"github.com/keep-network/keep-tecdsa/pkg/utils/pbutils"
)

func TestMarshalling(t *testing.T) {
	groupSize := 5
	dishonestThreshold := 4
	signerIndex := 2

	testData, err := testdata.LoadKeygenTestFixtures(groupSize)
	if err != nil {
		t.Fatalf("failed to load test data: [%v]", err)
	}

	groupMembersIDs := make([]MemberID, groupSize)

	for i := range groupMembersIDs {
		groupMembersIDs[i] = MemberID(fmt.Sprintf("member-%d", i))
	}

	signer := &ThresholdSigner{
		groupInfo: &groupInfo{
			groupID:            "test-group-id-1",
			memberID:           groupMembersIDs[signerIndex],
			groupMemberIDs:     groupMembersIDs,
			dishonestThreshold: dishonestThreshold,
		},
		keygenData: testData[signerIndex],
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
