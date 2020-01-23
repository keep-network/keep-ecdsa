package tss

import (
	"context"
	"fmt"
	"reflect"
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

	testMembers, err := generateTestMembers(groupSize, groupSize-2)
	if err != nil {
		t.Fatalf("failed to generate members keys: [%v]", err)
	}

	errChan := make(chan error)

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(groupSize)

	mutex := &sync.RWMutex{}
	joinedCount := 0

	for i, tm := range testMembers {
		go func(i int, tm *testMember) {
			groupInfo := &groupInfo{
				groupID:        "test-group-1",
				memberID:       MemberID([]byte(fmt.Sprintf("member-%d", i))),
				groupMemberIDs: testMembers.memberIDs(),
			}

			defer waitGroup.Done()

			membersPublicKeys, err := joinProtocol(ctx, groupInfo, tm.networkProvider)
			if err != nil {
				errChan <- err
				return
			}

			if !reflect.DeepEqual(membersPublicKeys, testMembers.groupPublicKeys()) {
				t.Errorf(
					"invalid list of members public keys\nexpected: [%v]\nactual:   [%v]",
					testMembers.groupPublicKeys(),
					membersPublicKeys,
				)
			}

			mutex.Lock()
			joinedCount++
			mutex.Unlock()
		}(i, tm)
	}

	go func() {
		waitGroup.Wait()
		cancel()
	}()

	select {
	case <-ctx.Done():
		if joinedCount != groupSize {
			t.Errorf(
				"invalid number of received notifications\nexpected: [%d]\nactual:   [%d]",
				groupSize,
				joinedCount,
			)
		}
	case err := <-errChan:
		t.Fatal(err)
	}
}
