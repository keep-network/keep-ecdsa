package tss

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-core/pkg/net/key"
)

func TestJoinNotifier(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

	err := log.SetLogLevel("*", "INFO")
	if err != nil {
		t.Fatalf("logger initialization failed: [%v]", err)
	}

	groupSize := 5

	groupMembers, err := generateMemberKeys(groupSize)
	if err != nil {
		t.Fatalf("failed to generate members keys: [%v]", err)
	}

	errChan := make(chan error)

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(groupSize)

	mutex := &sync.RWMutex{}
	joinedCount := 0

	for _, memberID := range groupMembers {
		go func(memberID MemberID) {
			groupInfo := &groupInfo{
				groupID:        "test-group-1",
				memberID:       memberID,
				groupMemberIDs: groupMembers,
			}

			memberPublicKey, err := memberID.PublicKey()
			if err != nil {
				errChan <- err
				return
			}

			memberNetworkKey := key.NetworkPublic(*memberPublicKey)
			networkProvider := newTestNetProvider(&memberNetworkKey)

			defer waitGroup.Done()

			if err := joinProtocol(ctx, groupInfo, networkProvider); err != nil {
				errChan <- err
				return
			}

			mutex.Lock()
			joinedCount++
			mutex.Unlock()
		}(memberID)
	}

	go func() {
		waitGroup.Wait()
		cancel()
	}()

	select {
	case <-ctx.Done():
		if joinedCount != groupSize {
			t.Errorf(
				"invalid number of received notifications\nexpected: [%d]\nactual:  [%d]",
				groupSize,
				joinedCount,
			)
		}
	case err := <-errChan:
		t.Fatal(err)
	}

}
