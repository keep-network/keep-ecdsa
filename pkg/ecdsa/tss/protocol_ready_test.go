package tss

import (
	"context"
	cecdsa "crypto/ecdsa"
	"crypto/elliptic"
	"sync"
	"testing"
	"time"

	"github.com/keep-network/keep-core/pkg/net"

	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-core/pkg/net/key"
)

func TestReadyProtocol(t *testing.T) {
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

	pubKeyToAddressFn := func(publicKey cecdsa.PublicKey) []byte {
		return elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
	}

	errChan := make(chan error)

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(groupSize)

	mutex := &sync.RWMutex{}
	readyCount := 0

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

			broadcastChannel, err := networkProvider.BroadcastChannelFor("test-group-1")
			if err != nil {
				errChan <- err
				return
			}

			broadcastChannel.SetUnmarshaler(func() net.TaggedUnmarshaler {
				return &ReadyMessage{}
			})

			defer waitGroup.Done()

			if err := readyProtocol(
				ctx,
				groupInfo,
				broadcastChannel,
				pubKeyToAddressFn,
			); err != nil {
				errChan <- err
				return
			}

			mutex.Lock()
			readyCount++
			mutex.Unlock()
		}(memberID)
	}

	go func() {
		waitGroup.Wait()
		cancel()
	}()

	select {
	case <-ctx.Done():
		if readyCount != groupSize {
			t.Errorf(
				"invalid number of received notifications\nexpected: [%d]\nactual:  [%d]",
				groupSize,
				readyCount,
			)
		}
	case err := <-errChan:
		t.Fatal(err)
	}

}
