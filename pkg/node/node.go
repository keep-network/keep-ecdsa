// Package node defines a node executing the TSS protocol.
package node

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss"
	"github.com/keep-network/keep-tecdsa/pkg/net"
)

var logger = log.Logger("keep-tecdsa")

// Node holds interfaces to interact with the blockchain and network messages
// transport layer.
type Node struct {
	ethereumChain   eth.Handle
	networkProvider net.Provider
	tssParamsPool   *tssPreParamsPool
}

// NewNode initializes node struct with provided ethereum chain interface and
// network provider. It also initializes TSS Pre-Parameters pool. But does not
// start parameters generation. This should be called separately.
func NewNode(
	ethereumChain eth.Handle,
	networkProvider net.Provider,
) *Node {
	return &Node{
		ethereumChain:   ethereumChain,
		networkProvider: networkProvider,
	}
}

// GenerateSignerForKeep generates a new threshold signer with ECDSA key pair. The
// public key is a public key of the signing group. It publishes the public key
// to the keep. It uses keep address as unique signing group identifier.
func (n *Node) GenerateSignerForKeep(
	keepAddress common.Address,
	keepMembers []common.Address,
	keepMemberIndex uint,
) (*tss.ThresholdSigner, error) {
	groupMemberIDs := []tss.MemberID{}
	membersNetworkIDs := make(map[string]net.TransportIdentifier) // < memberID, networkID >

	for i, keepMember := range keepMembers {
		memberID := tss.MemberID(fmt.Sprintf("member-%d", i))

		groupMemberIDs = append(
			groupMemberIDs,
			memberID,
		)

		networkID := hex.EncodeToString(keepMember.Bytes())
		membersNetworkIDs[memberID.String()] = tss.NetworkID(networkID)
	}

	signer, err := tss.GenerateThresholdSigner(
		keepAddress.Hex(),
		groupMemberIDs[keepMemberIndex],
		groupMemberIDs,
		uint(len(keepMembers)-1),
		membersNetworkIDs,
		n.networkProvider,
		n.tssParamsPool.get(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate threshold signer: [%v]", err)
	}

	logger.Debugf(
		"generated threshold signer with public key: [%x]",
		signer.PublicKey().Marshal(),
	)

	// Publish signer's public key on ethereum blockchain in a specific keep
	// contract.
	serializedPublicKey, err := eth.SerializePublicKey(signer.PublicKey())
	if err != nil {
		return nil, fmt.Errorf("failed to serialize public key: [%v]", err)
	}

	// TODO: Publisher Selection: Temp solution only the first member in the group
	// publishes. We need to replace it with proper publisher selection.
	if signer.PublisherIndex() == 0 {
		err = n.ethereumChain.SubmitKeepPublicKey(
			keepAddress,
			serializedPublicKey,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to submit public key: [%v]", err)

		}

		logger.Debugf(
			"submitted public key to the keep [%s]: [%x]",
			keepAddress.String(),
			serializedPublicKey,
		)
	}

	return signer, nil
}

// CalculateSignature calculates a signature over a digest with threshold
// signer and publishes the result to the keep associated with the signer.
func (n *Node) CalculateSignature(
	signer *tss.ThresholdSigner,
	digest [32]byte,
) error {
	signature, err := signer.CalculateSignature(digest[:], n.networkProvider)
	if err != nil {
		return fmt.Errorf("failed to calculate signature: [%v]", err)
	}

	logger.Debugf(
		"signature calculated:\nr: [%#x]\ns: [%#x]\nrecovery ID: [%d]\n",
		signature.R,
		signature.S,
		signature.RecoveryID,
	)

	keepAddress := common.HexToAddress(signer.GroupID())

	// TODO: Publisher Selection: Temp solution only the first member in the group
	// publishes. We need to replace it with proper publisher selection.
	if signer.PublisherIndex() == 0 {
		err = n.ethereumChain.SubmitSignature(keepAddress, signature)
		if err != nil {
			return fmt.Errorf("failed to submit signature: [%v]", err)
		}

		logger.Infof("submitted signature for digest: [%+x]", digest)
	}

	return nil
}
