// Package node defines a node executing the TSS protocol.
package node

import (
	"fmt"
	"time"

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
	tssParamsPool   *TSSPreParamsPool
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

// GenerateSignerForKeep generates a new signer with ECDSA key pair. It publishes
// the signer's public key to the keep.
func (n *Node) GenerateSignerForKeep(
	keepAddress eth.KeepAddress,
	keepMembers []common.Address,
) (*tss.ThresholdSigner, error) {
	// TODO: Temp Sync
	tss.KeyGenSync.Add(1)
	time.Sleep(2 * time.Second)

	groupMemberIDs := []tss.MemberID{}
	for _, member := range keepMembers {
		groupMemberIDs = append(
			groupMemberIDs,
			tss.MemberID(member.String()),
		)
	}

	memberID := tss.MemberID(n.ethereumChain.Address().String())

	signer, err := tss.GenerateThresholdSigner(
		keepAddress.Hex(),
		memberID,
		groupMemberIDs,
		uint(len(keepMembers)-1),
		n.networkProvider,
		n.tssParamsPool.Get(),
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

	// TODO: Temp solution only the first member in the group publishes.
	// We need to replace it with proper publisher selection.
	if memberID == groupMemberIDs[0] {
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

// CalculateSignatureForKeep calculates a signature over a digest with threshold
// signer and publishes the result to the keep.
func (n *Node) CalculateSignatureForKeep(
	keepAddress eth.KeepAddress,
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

	err = n.ethereumChain.SubmitSignature(keepAddress, digest, signature)
	if err != nil {
		return fmt.Errorf("failed to submit signature: [%v]", err)
	}

	logger.Infof("submitted signature for digest: [%+x]", digest)

	return nil
}
