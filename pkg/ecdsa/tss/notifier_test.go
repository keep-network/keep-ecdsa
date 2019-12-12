package tss

import (
	"fmt"
	"testing"

	"github.com/ipfs/go-log"
)

func TestJoinNotifier(t *testing.T) {
	err := log.SetLogLevel("*", "INFO")
	if err != nil {
		t.Fatalf("logger initialization failed: [%v]", err)
	}

	groupMembers, groupMembersKeys, err := generateMemberKeys(2)
	if err != nil {
		t.Fatalf("failed to generate members keys: [%v]", err)
	}

	errChan := make(chan error)

	go func() {
		for {
			fmt.Printf("%v", <-errChan)
		}
	}()

	joinNotifiers := []*joinNotifier{}

	for memberID, memberNetworkKey := range groupMembersKeys {
		groupInfo := &groupInfo{
			groupID:        "test-group-1",
			memberID:       memberID,
			groupMemberIDs: groupMembers,
		}

		networkProvider := newTestNetProvider(memberNetworkKey)

		joinNotifier, err := newJoinNotifier(groupInfo, networkProvider)
		if err != nil {
			t.Fatalf("failed to initialize join notifier: [%v]", err)
		}

		joinNotifiers = append(joinNotifiers, joinNotifier)
	}

	for _, joinNotifier := range joinNotifiers {
		if err := joinNotifier.notifyReady(); err != nil {
			t.Errorf("failed to notify: [%v]", err)
		}
	}

	joinNotifiers[0].waitForAll()
}
