package tss

import (
	"context"
	cecdsa "crypto/ecdsa"
	"fmt"
	"time"

	"github.com/keep-network/keep-core/pkg/net"
)

// BroadcastRecoveryAddress broadcasts and receives the BTC recovery addresses
// of each client so that each client can retrieve the underlying bitcoin in
// the case that a keep is terminated.
func BroadcastRecoveryAddress(
	parentCtx context.Context,
	groupID string,
	memberID MemberID,
	groupMemberIDs []MemberID,
	dishonestThreshold uint,
	networkProvider net.Provider,
	pubKeyToAddressFn func(cecdsa.PublicKey) []byte,
) error {
	const protocolReadyTimeout = 2 * time.Minute

	group := &groupInfo{
		groupID:            groupID,
		memberID:           memberID,
		groupMemberIDs:     groupMemberIDs,
		dishonestThreshold: int(dishonestThreshold),
	}

	netBridge, _ := newNetworkBridge(group, networkProvider)
	broadcastChannel, _ := netBridge.getBroadcastChannel()
	ctx, cancel := context.WithTimeout(parentCtx, protocolReadyTimeout)
	defer cancel()

	msgInChan := make(chan *LiquidationRecoveryAnnounceMessage, len(group.groupMemberIDs))
	handleLiquidationRecoveryAnnounceMessage := func(netMsg net.Message) {
		switch msg := netMsg.Payload().(type) {
		case *LiquidationRecoveryAnnounceMessage:
			msgInChan <- msg
		}
	}
	broadcastChannel.Recv(ctx, handleLiquidationRecoveryAnnounceMessage)

	memberBTCAddresses := make(map[string]string)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgInChan:
				for _, memberID := range group.groupMemberIDs {
					if msg.SenderID.Equal(memberID) {
						memberAddress, err := memberIDToAddress(
							memberID,
							pubKeyToAddressFn,
						)
						if err != nil {
							logger.Errorf(
								"could not convert member ID to address for "+
									"a member of keep [%s]: [%v]",
								group.groupID,
								err,
							)
							break
						}
						memberBTCAddresses[memberAddress] = msg.BtcRecoveryAddress

						logger.Infof(
							"member [%s] from keep [%s] announced supplied btc address [%s] for liquidation recovery",
							memberAddress,
							group.groupID,
							msg.BtcRecoveryAddress,
						)

						break
					}
				}

				if len(memberBTCAddresses) == len(group.groupMemberIDs) {
					cancel()
				}
			}
		}
	}()

	go func() {
		sendMessage := func() {
			if err := broadcastChannel.Send(ctx,
				&LiquidationRecoveryAnnounceMessage{SenderID: group.memberID, BtcRecoveryAddress: "abc123"},
			); err != nil {
				logger.Errorf("failed to send btc recovery address: [%v]", err)
			}
		}

		// Send the message first time. It will be periodically retransmitted
		// by the broadcast channel for the entire lifetime of the context.
		sendMessage()

		<-ctx.Done()
		// Send the message once again as the member received messages
		// from all peer members but not all peer members could receive
		// the message from the member as some peer member could join
		// the protocol after the member sent the last message.
		sendMessage()
		return
	}()

	<-ctx.Done()

	switch ctx.Err() {
	case context.DeadlineExceeded:
		for _, memberID := range group.groupMemberIDs {
			memberAddress, err := memberIDToAddress(memberID, pubKeyToAddressFn)
			if err != nil {
				logger.Errorf(
					"could not convert member ID to address for a member of "+
						"keep [%s]: [%v]",
					group.groupID,
					err,
				)
				continue
			}
			if _, present := memberBTCAddresses[memberAddress]; !present {
				logger.Errorf(
					"member [%s] has not supplied a btc recovery address for keep [%s]; "+
						"check if keep client for that operator is active and "+
						"connected",
					memberAddress,
					group.groupID,
				)
			}
		}
		return fmt.Errorf(
			"waiting for btc recovery addresses timed out after: [%v]", protocolReadyTimeout,
		)
	case context.Canceled:
		logger.Infof("successfully gathered all btc addresses")

		return nil
	default:
		return fmt.Errorf("unexpected context error: [%v]", ctx.Err())
	}
}
