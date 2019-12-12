package tss

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ipfs/go-log"
)

func TestJoinNotifier(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

	err := log.SetLogLevel("*", "INFO")
	if err != nil {
		t.Fatalf("logger initialization failed: [%v]", err)
	}

	groupSize := 5

	groupMembers, groupMembersKeys, err := generateMemberKeys(groupSize)
	if err != nil {
		t.Fatalf("failed to generate members keys: [%v]", err)
	}

	errChan := make(chan error)

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

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(groupSize)

	mutex := &sync.RWMutex{}
	joinedCount := 0

	for _, jn := range joinNotifiers {
		go func(jn *joinNotifier) {
			defer waitGroup.Done()

			if err := jn.notifyReady(); err != nil {
				errChan <- err
				return
			}

			mutex.Lock()
			joinedCount++
			mutex.Unlock()
		}(jn)
	}

	go func() {
		defer cancel()
		waitGroup.Wait()
	}()

	select {
	case <-ctx.Done():
		if joinedCount != groupSize {
			t.Errorf(
				"invalid number of received notifications\nexpected: [%d]\nactual:  [%d]",
				groupSize-1,
				joinedCount,
			)
		}
	case err := <-errChan:
		t.Fatal(err)
	}

}
