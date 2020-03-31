// Package node defines a node executing the TSS protocol.
package node

import (
	"context"
	cecdsa "crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/keep-network/keep-core/pkg/operator"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-core/pkg/net"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss"
)

var logger = log.Logger("keep-ecdsa")

const monitorKeepPublicKeySubmissionTimeout = 30 * time.Minute

// Node holds interfaces to interact with the blockchain and network messages
// transport layer.
type Node struct {
	ethereumChain   eth.Handle
	networkProvider net.Provider
	tssParamsPool   *tssPreParamsPool
	tssConfig       *tss.Config
}

// NewNode initializes node struct with provided ethereum chain interface and
// network provider. It also initializes TSS Pre-Parameters pool. But does not
// start parameters generation. This should be called separately.
func NewNode(
	ethereumChain eth.Handle,
	networkProvider net.Provider,
	tssConfig *tss.Config,
) *Node {
	return &Node{
		ethereumChain:   ethereumChain,
		networkProvider: networkProvider,
		tssConfig:       tssConfig,
	}
}

// AnnounceSignerPresence triggers the announce protocol in order to signal
// signer presence and gather information about other signers.
func (n *Node) AnnounceSignerPresence(
	ctx context.Context,
	operatorPublicKey *operator.PublicKey,
	keepAddress common.Address,
	keepMembersAddresses []common.Address,
) ([]tss.MemberID, error) {
	broadcastChannel, err := n.networkProvider.BroadcastChannelFor(keepAddress.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize broadcast channel: [%v]", err)
	}

	tss.RegisterUnmarshalers(broadcastChannel)

	if err := broadcastChannel.SetFilter(
		createAddressFilter(keepMembersAddresses),
	); err != nil {
		return nil, fmt.Errorf("failed to set broadcast channel filter: [%v]", err)
	}

	return tss.AnnounceProtocol(
		ctx,
		operatorPublicKey,
		len(keepMembersAddresses),
		broadcastChannel,
	)
}

func createAddressFilter(
	addresses []common.Address,
) net.BroadcastChannelFilter {
	authorizations := make(map[string]bool, len(addresses))
	for _, address := range addresses {
		authorizations[hex.EncodeToString(address.Bytes())] = true
	}

	return func(authorPublicKey *cecdsa.PublicKey) bool {
		authorAddress := hex.EncodeToString(
			crypto.PubkeyToAddress(*authorPublicKey).Bytes(),
		)
		_, isAuthorized := authorizations[authorAddress]

		if !isAuthorized {
			logger.Warningf(
				"rejecting message from [%v]; author is not authorized",
				authorAddress,
			)
		}

		return isAuthorized
	}
}

// GenerateSignerForKeep generates a new threshold signer with ECDSA key pair. The
// public key is a public key of the signing group. It publishes the public key
// to the keep. It uses keep address as unique signing group identifier.
func (n *Node) GenerateSignerForKeep(
	ctx context.Context,
	operatorPublicKey *operator.PublicKey,
	keepAddress common.Address,
	keepMembersIDs []tss.MemberID,
) (*tss.ThresholdSigner, error) {
	memberID := tss.MemberIDFromPublicKey(operatorPublicKey)

	signer, err := tss.GenerateThresholdSigner(
		ctx,
		keepAddress.Hex(),
		memberID,
		keepMembersIDs,
		uint(len(keepMembersIDs)-1),
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

	monitoringAbort := make(chan interface{})
	go n.monitorKeepPublicKeySubmission(ctx, monitoringAbort, keepAddress)

	err = n.ethereumChain.SubmitKeepPublicKey(
		keepAddress,
		serializedPublicKey,
	)
	if err != nil {
		close(monitoringAbort)
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
	ctx context.Context,
	signer *tss.ThresholdSigner,
	digest [32]byte,
) error {
	signature, err := signer.CalculateSignature(ctx, digest[:], n.networkProvider)
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

// monitorKeepPublicKeySubmission observes the chain until either the first
// conflicting public key is published or until keep established public key
// or until key generation timed out.
func (n *Node) monitorKeepPublicKeySubmission(
	ctx context.Context,
	abort chan interface{},
	keepAddress common.Address,
) {
	monitoringCtx, monitoringCancel := context.WithTimeout(
		ctx,
		monitorKeepPublicKeySubmissionTimeout,
	)
	defer monitoringCancel()

	publicKeyPublished := make(chan *eth.PublicKeyPublishedEvent)
	conflictingPublicKey := make(chan *eth.ConflictingPublicKeySubmittedEvent)

	subscriptionPublicKeyPublished, err := n.ethereumChain.OnPublicKeyPublished(
		keepAddress,
		func(event *eth.PublicKeyPublishedEvent) {
			publicKeyPublished <- event
		},
	)
	if err != nil {
		logger.Errorf(
			"failed on watching public key published event: [%v]",
			err,
		)
	}

	subscriptionConflictingPublicKey, err := n.ethereumChain.OnConflictingPublicKeySubmitted(
		keepAddress,
		func(event *eth.ConflictingPublicKeySubmittedEvent) {
			conflictingPublicKey <- event
		},
	)
	if err != nil {
		logger.Errorf(
			"failed on watching conflicting public key event: [%v]",
			err,
		)
	}

	defer subscriptionConflictingPublicKey.Unsubscribe()
	defer subscriptionPublicKeyPublished.Unsubscribe()

	select {
	case event := <-publicKeyPublished:
		logger.Infof(
			"public key [%x] has been accepted by keep",
			event.PublicKey,
		)
	case event := <-conflictingPublicKey:
		logger.Errorf(
			"member [%x] has submitted conflicting public key: [%x]",
			event.SubmittingMember,
			event.ConflictingPublicKey,
		)
	case <-monitoringCtx.Done():
		if monitoringCtx.Err() == context.DeadlineExceeded {
			logger.Warningf(
				"monitoring of public key submission for keep [%s] "+
					"has been cancelled due to exceeded timeout",
				keepAddress.String(),
			)
		}
	case <-abort:
		logger.Warningf(
			"monitoring of public key submission for keep [%s] "+
				"has been aborted",
			keepAddress.String(),
		)
	}
}
