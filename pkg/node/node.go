// Package node defines a node executing the TSS protocol.
package node

import (
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
) (*tss.ThresholdSigner, error) {
	groupMemberIDs := []tss.MemberID{}
	for _, memberAddress := range keepMembers {
		memberID, err := tss.MemberIDFromHex(memberAddress.Hex())
		if err != nil {
			return nil, fmt.Errorf("failed to convert member address to member ID: [%v]", err)
		}

		groupMemberIDs = append(
			groupMemberIDs,
			memberID,
		)
	}

	memberID, err := tss.MemberIDFromHex(n.ethereumChain.Address().String())
	if err != nil {
		return nil, fmt.Errorf("failed to convert client's address to member ID: [%v]", err)
	}

	signer, err := tss.GenerateThresholdSigner(
		keepAddress.Hex(),
		memberID,
		groupMemberIDs,
		uint(len(keepMembers)-1),
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

	return signer, nil
}

// CalculateSignature calculates a signature over a digest with threshold
// signer and publishes the result to the keep associated with the signer.
// In case of failure on signature submission we need to check if the keep is
// still waiting for the signature. It is possible that other member was faster
// than the current one and submitted the signature first.
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

	if err := n.ethereumChain.SubmitSignature(keepAddress, signature); err != nil {
		isAwaitingSignature, err := n.ethereumChain.IsAwaitingSignature(keepAddress, digest)
		if err != nil {
			return fmt.Errorf("failed to verify if keep is still awaiting signature: [%v]", err)
		}

		if !isAwaitingSignature {
			logger.Infof("signature submitted by another member: [%+x]", digest)

			return nil
		}

		return fmt.Errorf("failed to submit signature: [%v]", err)
	}

	logger.Infof("submitted signature for digest: [%+x]", digest)

	return nil
}
