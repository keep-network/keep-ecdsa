package registry

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-tecdsa/pkg/registry/gen/pb"
)

// Marshal converts Membership to a byte array.
func (m *Membership) Marshal() ([]byte, error) {
	signer, err := m.Signer.Marshal()
	if err != nil {
		return nil, err
	}

	return (&pb.Membership{
		Signer:      signer,
		KeepAddress: m.KeepAddress.Bytes(),
	}).Marshal()
}

// Unmarshal converts a byte array produced by Marshal to Membership.
func (m *Membership) Unmarshal(bytes []byte) error {
	pbMembership := pb.Membership{}
	if err := pbMembership.Unmarshal(bytes); err != nil {
		return err
	}

	signer := &ecdsa.Signer{}

	err := signer.Unmarshal(pbMembership.Signer)
	if err != nil {
		return fmt.Errorf("unexpected error occurred: [%v]", err)
	}

	m.Signer = signer
	m.KeepAddress = common.BytesToAddress(pbMembership.KeepAddress)

	return nil
}
