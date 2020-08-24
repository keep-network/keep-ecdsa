package tss

import (
	"context"
	"github.com/keep-network/keep-core/pkg/net"
	"sync"
	"testing"
	"time"

	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-core/pkg/net/key"
)

func TestAnnounceProtocol(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

	err := log.SetLogLevel("*", "INFO")
	if err != nil {
		t.Fatalf("logger initialization failed: [%v]", err)
	}

	groupSize := 3

	groupMembers, err := generateMemberKeys(groupSize)
	if err != nil {
		t.Fatalf("failed to generate members keys: [%v]", err)
	}

	errChan := make(chan error)

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(groupSize)

	mutex := &sync.RWMutex{}
	result := make(map[string][]MemberID)

	for _, memberID := range groupMembers {
		go func(memberID MemberID) {
			memberPublicKey, err := memberID.PublicKey()
			if err != nil {
				errChan <- err
				return
			}

			memberNetworkKey := key.NetworkPublic(*memberPublicKey)
			networkProvider := newTestNetProvider(&memberNetworkKey)

			broadcastChannel, err := networkProvider.BroadcastChannelFor("test-group-1")
			if err != nil {
				errChan <- err
				return
			}

			broadcastChannel.SetUnmarshaler(func() net.TaggedUnmarshaler {
				return &AnnounceMessage{}
			})

			defer waitGroup.Done()

			memberIDs, err := AnnounceProtocol(
				ctx,
				memberPublicKey,
				groupSize,
				broadcastChannel,
			)
			if err != nil {
				errChan <- err
				return
			}

			mutex.Lock()
			result[memberID.String()] = memberIDs
			mutex.Unlock()
		}(memberID)
	}

	go func() {
		waitGroup.Wait()
		cancel()
	}()

	select {
	case <-ctx.Done():
		if len(result) != groupSize {
			t.Errorf(
				"invalid number of results\nexpected: [%d]\nactual:  [%d]",
				groupSize,
				len(result),
			)
		}

		for _, memberID := range groupMembers {
			if memberResult, ok := result[memberID.String()]; ok {
				for _, otherMemberID := range groupMembers {
					exists := false
					for _, result := range memberResult {
						if result.Equal(otherMemberID) {
							exists = true
							break
						}
					}
					if !exists {
						t.Errorf(
							"result of member [%v] doesn't contain "+
								"key for other member [%v]",
							memberID,
							otherMemberID,
						)
					}
				}
			} else {
				t.Errorf("missing result for member [%v]", memberID)
			}
		}
	case err := <-errChan:
		t.Fatal(err)
	}
}
