package tss

import (
	"context"
	cecdsa "crypto/ecdsa"
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

	groupMembers, groupMembersKeys, err := generateMemberKeys(groupSize)
	if err != nil {
		t.Fatalf("failed to generate members keys: [%v]", err)
	}

	errChan := make(chan error)

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(groupSize)

	mutex := &sync.RWMutex{}
	result := make(map[string]map[string]cecdsa.PublicKey)

	for memberIDstring, memberPublicKey := range groupMembersKeys {
		memberID, _ := MemberIDFromHex(memberIDstring)

		go func(memberID MemberID, memberPublicKey cecdsa.PublicKey) {
			groupInfo := &groupInfo{
				groupID:         "test-group-1",
				memberID:        memberID,
				memberPublicKey: memberPublicKey,
				groupMemberIDs:  groupMembers,
			}

			memberNetworkKey := key.NetworkPublic(memberPublicKey)
			networkProvider := newTestNetProvider(&memberNetworkKey)

			defer waitGroup.Done()

			groupMemberPublicKeys, err := announceProtocol(ctx, groupInfo, networkProvider)
			if err != nil {
				errChan <- err
				return
			}

			mutex.Lock()
			result[memberID.String()] = groupMemberPublicKeys
			mutex.Unlock()
		}(memberID, memberPublicKey)
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

		for memberID, _ := range groupMembersKeys {
			if memberResult, ok := result[memberID]; ok {
				for otherMemberID, _ := range groupMembersKeys {
					if _, ok := memberResult[otherMemberID]; !ok {
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
