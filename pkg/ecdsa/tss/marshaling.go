package tss

import (
	"encoding/json"
	"fmt"

	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss/gen/pb"
)

// Marshal converts ThresholdSigner to byte array.
func (s *ThresholdSigner) Marshal() ([]byte, error) {
	keygenData, err := json.Marshal(s.keygenData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key data: [%v]", err)
	}

	// Group Info
	groupMemberIDs := make([][]byte, len(s.groupMemberIDs))
	for i, memberID := range s.groupMemberIDs {
		groupMemberIDs[i] = memberID
	}

	group := &pb.ThresholdSigner_GroupInfo{
		GroupID:            s.groupID,
		MemberID:           s.memberID,
		GroupMemberIDs:     groupMemberIDs,
		DishonestThreshold: int32(s.dishonestThreshold),
	}

	return (&pb.ThresholdSigner{
		GroupInfo:    group,
		ThresholdKey: keygenData,
	}).Marshal()
}

// Unmarshal converts a byte array back to ThresholdSigner.
func (s *ThresholdSigner) Unmarshal(bytes []byte) error {
	pbSigner := pb.ThresholdSigner{
		GroupInfo: &pb.ThresholdSigner_GroupInfo{},
	}
	if err := pbSigner.Unmarshal(bytes); err != nil {
		return fmt.Errorf("failed to unmarshal signer: [%v]", err)
	}

	if err := json.Unmarshal(pbSigner.GetThresholdKey(), &s.keygenData); err != nil {
		return fmt.Errorf("failed to unmarshal key data: [%v]", err)
	}

	// Group Info
	pbGroupInfo := pbSigner.GetGroupInfo()

	groupMemberIDs := make([]MemberID, len(pbGroupInfo.GetGroupMemberIDs()))
	for i, memberID := range pbGroupInfo.GetGroupMemberIDs() {
		groupMemberIDs[i] = MemberID(memberID)
	}

	s.groupInfo = &groupInfo{
		groupID:            pbGroupInfo.GetGroupID(),
		memberID:           MemberID(pbGroupInfo.GetMemberID()),
		groupMemberIDs:     groupMemberIDs,
		dishonestThreshold: int(pbGroupInfo.GetDishonestThreshold()),
	}

	return nil
}

// Marshal converts this message to a byte array suitable for network communication.
func (m *TSSProtocolMessage) Marshal() ([]byte, error) {
	return (&pb.TSSProtocolMessage{
		SenderID:    m.SenderID,
		Payload:     m.Payload,
		IsBroadcast: m.IsBroadcast,
	}).Marshal()
}

// Unmarshal converts a byte array produced by Marshal to a message.
func (m *TSSProtocolMessage) Unmarshal(bytes []byte) error {
	pbMsg := &pb.TSSProtocolMessage{}
	if err := pbMsg.Unmarshal(bytes); err != nil {
		return err
	}

	m.SenderID = MemberID(pbMsg.SenderID)
	m.Payload = pbMsg.Payload
	m.IsBroadcast = pbMsg.IsBroadcast

	return nil
}

// Marshal converts this message to a byte array suitable for network communication.
func (m *JoinMessage) Marshal() ([]byte, error) {
	return (&pb.JoinMessage{
		SenderID: m.SenderID,
	}).Marshal()
}

// Unmarshal converts a byte array produced by Marshal to a message.
func (m *JoinMessage) Unmarshal(bytes []byte) error {
	pbMsg := &pb.JoinMessage{}
	if err := pbMsg.Unmarshal(bytes); err != nil {
		return err
	}

	m.SenderID = MemberID(pbMsg.SenderID)

	return nil
}
