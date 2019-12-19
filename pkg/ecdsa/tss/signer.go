package tss

import (
	"encoding/json"
	"fmt"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	tssLib "github.com/binance-chain/tss-lib/tss"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-tecdsa/pkg/registry/gen/pb"
)

// ThresholdSigner is a threshold signer who completed key generation stage.
type ThresholdSigner struct {
	*groupInfo

	// keygenData contains output of key generation stage. This data should be
	// persisted to local storage.
	keygenData keygen.LocalPartySaveData
}

// MemberID returns member's unique identifer.
func (s *ThresholdSigner) MemberID() MemberID {
	return s.memberID
}

// GroupID return signing group unique identifer.
func (s *ThresholdSigner) GroupID() string {
	return s.groupID
}

// PublicKey returns signer's ECDSA public key which is also the signing group's
// public key.
func (s *ThresholdSigner) PublicKey() *ecdsa.PublicKey {
	pkX, pkY := s.keygenData.ECDSAPub.X(), s.keygenData.ECDSAPub.Y()

	curve := tssLib.EC()
	publicKey := ecdsa.PublicKey{
		Curve: curve,
		X:     pkX,
		Y:     pkY,
	}

	return (*ecdsa.PublicKey)(&publicKey)
}

// Marshal converts ThresholdSigner to byte array.
func (s *ThresholdSigner) Marshal() ([]byte, error) {
	keygenData, err := json.Marshal(s.keygenData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key data: [%v]", err)
	}

	// Group Info
	groupMemberIDs := make([]string, len(s.groupMemberIDs))
	for i, memberID := range s.groupMemberIDs {
		groupMemberIDs[i] = memberID.string()
	}

	group := &pb.ThresholdSigner_GroupInfo{
		GroupID:            s.groupID,
		MemberID:           s.memberID.string(),
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
