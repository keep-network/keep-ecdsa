package tss

import (
	"encoding/json"
	"fmt"

	"github.com/keep-network/keep-tecdsa/pkg/registry/gen/pb"
)

// Marshal converts ThresholdSigner to byte array.
func (s *ThresholdSigner) Marshal() ([]byte, error) {
	keygenData, err := json.Marshal(s.keygenData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key data: [%v]", err)
	}

	// Group Info
	groupMemberIDs := make([]string, len(s.groupMemberIDs))
	for i, memberID := range s.groupMemberIDs {
		groupMemberIDs[i] = memberID.String()
	}

	group := &pb.ThresholdSigner_GroupInfo{
		GroupID:            s.groupID,
		MemberID:           s.memberID.String(),
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
