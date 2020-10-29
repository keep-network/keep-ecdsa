// Package node defines a node executing the TSS protocol.
package node

import (
	"context"
	cecdsa "crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/keep-network/keep-ecdsa/pkg/registry"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/keep-network/keep-core/pkg/operator"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-core/pkg/net"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss/params"
)

var logger = log.Logger("keep-ecdsa")

const monitorKeepPublicKeySubmissionTimeout = 30 * time.Minute
const retryDelay = 1 * time.Second

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

// GenerateSignerForKeep generates a new threshold signer with ECDSA key pair
// and submits the public key to the on-chain keep.
//
// The attempt for generating signer is retried on failure until the provided
// context is done.
func (n *Node) GenerateSignerForKeep(
	ctx context.Context,
	operatorPublicKey *operator.PublicKey,
	keepAddress common.Address,
	members []common.Address,
	keepsRegistry *registry.Keeps,
) (*tss.ThresholdSigner, error) {
	memberID := tss.MemberIDFromPublicKey(operatorPublicKey)
	preParamsBox := params.NewBox(n.tssParamsPool.get())

	attemptCounter := 0
	for {
		attemptCounter++

		logger.Infof(
			"signer generation for keep [%s]; attempt [%v]",
			keepAddress.String(),
			attemptCounter,
		)

		isActive, err := n.ethereumChain.IsActive(keepAddress)
		if err != nil {
			logger.Warningf(
				"could not check if keep [%s] is still active: [%v]",
				keepAddress.String(),
				err,
			)
			time.Sleep(retryDelay) // TODO: #413 Replace with backoff.
			continue
		}

		// If the keep is not active there is no point in generating a signer as
		// the keep is either closed or terminated.
		if !isActive {
			return nil, fmt.Errorf("keep is no longer active")
		}

		// If we are re-attempting the key generation, pre-parameters in the box
		// could be destroyed because they were shared with other members.
		// In this case, we need to re-generate them.
		if preParamsBox.IsEmpty() {
			preParamsBox = params.NewBox(n.tssParamsPool.get())
		}

		// Global timeout for generating a signer exceeded.
		// We are giving up and leaving this function.
		if ctx.Err() != nil {
			return nil, fmt.Errorf("key generation timeout exceeded")
		}

		// Announce signer presence. Other members of the keep need to receive
		// the public key of this members. This member, need to receive public
		// keys of all other members. Up to this point, only addresses from
		// signer selection protocol are known.
		//
		// If signer announcement fails, we retry from the beginning.
		memberIDs, err := n.AnnounceSignerPresence(
			ctx,
			operatorPublicKey,
			keepAddress,
			members,
		)
		if err != nil {
			logger.Warningf("failed to announce signer presence: [%v]", err)
			time.Sleep(retryDelay) // TODO: #413 Replace with backoff.
			continue
		}

		// Generate threshold signer by generating threshold key with all other
		// keep members.
		//
		// If threshold key generation fails, we retry from the beginning.
		signer, err := tss.GenerateThresholdSigner(
			ctx,
			keepAddress.Hex(),
			memberID,
			memberIDs,
			uint(len(memberIDs)-1),
			n.networkProvider,
			preParamsBox,
		)
		if err != nil {
			logger.Errorf("failed to generate threshold signer: [%v]", err)
			time.Sleep(retryDelay) // TODO: #413 Replace with backoff.
			continue
		}

		// Make a snapshot of the generated signer before publishing the public
		// key to the keep. This guarantees the signer and their key share are
		// safely persisted before the public key is registered on-chain.
		// Then, the snapshot can be used for signer recovery in case something
		// bad occurs before the final signer registration will be done.
		err = keepsRegistry.SnapshotSigner(keepAddress, signer)
		if err != nil {
			return nil, fmt.Errorf(
				"could not make snapshot of signer for keep [%s]: [%v]",
				keepAddress.String(),
				err,
			)
		}

		// Serialize and publish public key to the keep.
		//
		// We don't retry in case of an error although the specific chain
		// implementation may implement its own retry policy. This action
		// should never fail and if it failed, something terrible happened.
		publicKey, err := eth.SerializePublicKey(signer.PublicKey())
		if err != nil {
			return nil, fmt.Errorf("failed to serialize public key: [%v]", err)
		}
		err = n.publishSignerPublicKey(ctx, keepAddress, publicKey)
		if err != nil {
			return nil, err
		}

		return signer, nil // key generation succeeded.
	}
}

func (n *Node) publishSignerPublicKey(
	ctx context.Context,
	keepAddress common.Address,
	publicKey [64]byte,
) error {
	logger.Debugf(
		"submitting public key to the keep [%s]: [%x]",
		keepAddress.String(),
		publicKey,
	)

	monitoringAbort := make(chan interface{})
	go n.monitorKeepPublicKeySubmission(monitoringAbort, keepAddress)

	err := n.ethereumChain.SubmitKeepPublicKey(keepAddress, publicKey)
	if err != nil {
		close(monitoringAbort)
		return fmt.Errorf("failed to submit public key: [%v]", err)
	}

	return nil
}

// CalculateSignature calculates a signature over a digest with threshold
// signer and publishes the result to the keep associated with the signer.
//
// The attempt for generating and publishing signature is retried on failure
// until the provided context is done.
func (n *Node) CalculateSignature(
	ctx context.Context,
	signer *tss.ThresholdSigner,
	digest [32]byte,
) error {
	keepAddress := common.HexToAddress(signer.GroupID())

	attemptCounter := 0
	for {
		attemptCounter++

		logger.Infof(
			"calculate signature for keep [%s]; attempt [%v]",
			keepAddress.String(),
			attemptCounter,
		)

		// Global timeout for generating a signature exceeded.
		// We are giving up and leaving this function.
		if ctx.Err() != nil {
			return fmt.Errorf("signing timeout exceeded")
		}

		// Calculate the signature executing threshold signing protocol with
		// other keep members.
		//
		// If threshold signing fails, we retry from the beginning.
		signature, err := signer.CalculateSignature(ctx, digest[:], n.networkProvider)
		if err != nil {
			logger.Errorf(
				"failed to calculate signature for keep [%s]: [%v]",
				keepAddress.String(),
				err,
			)
			time.Sleep(retryDelay) // TODO: #413 Replace with backoff.
			continue
		}

		logger.Debugf(
			"signature calculated:\nr: [%#x]\ns: [%#x]\nrecovery ID: [%d]\n",
			signature.R,
			signature.S,
			signature.RecoveryID,
		)

		// We have the signature so now we need to publish it.
		// This function implements internal retries so we do not need to
		// retry here.
		return n.publishSignature(ctx, keepAddress, digest, signature)
	}
}

// publishSignature takes the provided signature and attempts to publish it to
// the chain. It implements retry mechanism allowing to attempt to publish again
// in case of a failure.
//
// We do implement a retry in this function because the retry mechanism is much
// more complex than in case of e.g. publishSignerPublicKey. Although all keep
// members are supposed to try to publish the signature, only one transaction
// succeeds. For each attempt, we need to check if the keep still awaits
// a signature. Also, we need to implement some sane delay between attempts so
// that we do not waste gas.
func (n *Node) publishSignature(
	ctx context.Context,
	keepAddress common.Address,
	digest [32]byte,
	signature *ecdsa.Signature,
) error {
	attemptCounter := 0
	for {
		attemptCounter++

		// Global timeout for generating a signature exceeded.
		// We are giving up and leaving this function.
		if ctx.Err() != nil {
			return fmt.Errorf("context timeout exceeded")
		}

		// Check if keep is still active. There is no point in submitting the
		// request when keep is no longer active, which means that it was either
		// closed or terminated and signers' bonds might have been seized already.
		// We are giving up and leaving this function.
		isActive, err := n.ethereumChain.IsActive(keepAddress)
		if err != nil {
			logger.Errorf(
				"failed to verify if keep [%s] is still active: [%v]",
				keepAddress.String(),
				err,
			)
			time.Sleep(retryDelay) // TODO: #413 Replace with backoff.
			continue
		}
		if !isActive {
			return fmt.Errorf("keep is no longer active")
		}

		// Check if keep still awaits a signature for this digest.
		// We do this check here in case the attempt was retried because of
		// on-chain failure during submission. In this case we want to make sure
		// no other member published the signature in the meantime so that we
		// do not burn ether on redundant submission.
		//
		// If the check failed, we retry from the beginning.
		isAwaitingSignature, err := n.ethereumChain.IsAwaitingSignature(keepAddress, digest)
		if err != nil {
			logger.Errorf(
				"failed to verify if keep [%s] is still awaiting signature: [%v]",
				keepAddress.String(),
				err,
			)
			time.Sleep(retryDelay) // TODO: #413 Replace with backoff.
			continue
		}

		// Someone submitted the signature and it was accepted by the keep.
		// We are fine, leaving.
		if !isAwaitingSignature {
			logger.Infof(
				"signature for keep [%s] already submitted: [%+x]",
				keepAddress.String(),
				digest,
			)
			return nil
		}

		logger.Infof(
			"publishing signature for keep [%s]; attempt [%v]",
			keepAddress.String(),
			attemptCounter,
		)

		if submissionErr := n.ethereumChain.SubmitSignature(keepAddress, signature); submissionErr != nil {
			isAwaitingSignature, err := n.ethereumChain.IsAwaitingSignature(keepAddress, digest)
			if err != nil {
				logger.Errorf(
					"failed to verify if keep [%s] is still awaiting signature: [%v]",
					keepAddress.String(),
					err,
				)
				time.Sleep(retryDelay) // TODO: #413 Replace with backoff.
				continue
			}

			// Check if we failed because someone else submitted in the meantime
			// or because something wrong happened with our transaction.
			if !isAwaitingSignature {
				logger.Infof(
					"signature for keep [%s] already submitted: [%+x]",
					keepAddress.String(),
					digest,
				)
				return nil
			}

			// Our public key submission transaction failed. We are going to
			// wait for some time and then retry from the beginning.
			logger.Errorf(
				"failed to submit signature for keep [%s]: [%v]; "+
					"will retry after 1 minute",
				keepAddress.String(),
				submissionErr,
			)
			time.Sleep(1 * time.Minute)
			continue
		}

		logger.Infof("signature submitted for keep [%s]", keepAddress.String())
		return nil
	}
}

// monitorKeepPublicKeySubmission observes the chain until either the first
// conflicting public key is published or until keep established public key
// or until key generation timed out.
func (n *Node) monitorKeepPublicKeySubmission(
	abort chan interface{},
	keepAddress common.Address,
) {
	monitoringCtx, monitoringCancel := context.WithTimeout(
		context.Background(),
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
			"failed on watching public key published event for keep [%s]: [%v]",
			keepAddress.String(),
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
			"failed on watching conflicting public key event for keep [%s]: [%v]",
			keepAddress.String(),
			err,
		)
	}

	defer subscriptionConflictingPublicKey.Unsubscribe()
	defer subscriptionPublicKeyPublished.Unsubscribe()

	select {
	case event := <-publicKeyPublished:
		logger.Infof(
			"public key [%x] has been accepted by keep: [%s]",
			event.PublicKey,
			keepAddress.String(),
		)
	case event := <-conflictingPublicKey:
		logger.Errorf(
			"member [%x] has submitted conflicting public key for keep [%s]: [%x]",
			event.SubmittingMember,
			keepAddress.String(),
			event.ConflictingPublicKey,
		)
	case <-monitoringCtx.Done():
		logger.Warningf(
			"monitoring of public key submission for keep [%s] "+
				"has been cancelled: [%v]",
			keepAddress.String(),
			monitoringCtx.Err(),
		)
	case <-abort:
		logger.Warningf(
			"monitoring of public key submission for keep [%s] "+
				"has been aborted",
			keepAddress.String(),
		)
	}
}
