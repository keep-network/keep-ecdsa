package tss

import (
	"encoding/json"
	"fmt"

	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss/gen/pb"
)

// Marshal converts Signer to byte array.
func (s *ThresholdSigner) Marshal() ([]byte, error) {
	keygenData, err := json.Marshal(s.keygenData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key data: [%v]", err)
	}

	// Group Info
	groupMemberIDs := make([]string, len(s.groupMemberIDs))
	for i, memberID := range s.groupMemberIDs {
		groupMemberIDs[i] = string(memberID)
	}

	groupInfo := &pb.Signer_GroupInfo{
		GroupID:            s.groupID,
		MemberID:           string(s.memberID),
		GroupMemberIDs:     groupMemberIDs,
		DishonestThreshold: int32(s.dishonestThreshold),
	}

	return (&pb.Signer{
		GroupInfo:  groupInfo,
		KeygenData: keygenData,
	}).Marshal()
}

// Unmarshal converts a byte array back to Signer.
func (s *ThresholdSigner) Unmarshal(bytes []byte) error {
	pbSigner := pb.Signer{
		GroupInfo: &pb.Signer_GroupInfo{},
	}
	if err := pbSigner.Unmarshal(bytes); err != nil {
		return fmt.Errorf("failed to unmarshal signer: [%v]", err)
	}

	if err := json.Unmarshal(pbSigner.GetKeygenData(), &s.keygenData); err != nil {
		return fmt.Errorf("failed to unmarshal key data: [%v]", err)
	}

	// Group Info
	pbGroupInfo := pbSigner.GetGroupInfo()

	groupMemberIDs := make([]MemberID, len(pbGroupInfo.GetGroupMemberIDs()))
	for i, memberID := range pbGroupInfo.GetGroupMemberIDs() {
		groupMemberIDs[i] = MemberID(memberID)
	}

	s.GroupInfo = &GroupInfo{
		groupID:            pbGroupInfo.GetGroupID(),
		memberID:           MemberID(pbGroupInfo.GetMemberID()),
		groupMemberIDs:     groupMemberIDs,
		dishonestThreshold: int(pbGroupInfo.GetDishonestThreshold()),
	}

	return nil
}
