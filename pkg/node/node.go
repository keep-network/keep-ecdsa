// Package node defines a node executing the TSS protocol.
package node

import (
	"context"
	cecdsa "crypto/ecdsa"
	"fmt"
	"time"

	"github.com/keep-network/keep-common/pkg/chain/ethlike"

	"github.com/keep-network/keep-ecdsa/pkg/registry"

	"github.com/keep-network/keep-core/pkg/operator"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-core/pkg/net"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss/params"
)

var logger = log.Logger("keep-ecdsa")

const (
	// Determines the delay which should be preserved before retrying
	// actions within the key generation and signing process.
	retryDelay = 1 * time.Second

	// Number of blocks which should elapse before confirming
	// the given chain state expectations.
	blockConfirmations = uint64(12)

	// Used to calculate the publication delay factor for the given signer index
	// to avoid all signers publishing the same signature for given keep at the
	// same time.
	signaturePublicationDelayStep = 90 * time.Second
)

// Node holds interfaces to interact with the blockchain and network messages
// transport layer.
type Node struct {
	chain           chain.Handle
	networkProvider net.Provider
	tssParamsPool   *tssPreParamsPool
	tssConfig       *tss.Config
}

// NewNode initializes node struct with provided chain interface and
// network provider. It also initializes TSS Pre-Parameters pool. But does not
// start parameters generation. This should be called separately.
func NewNode(
	chain chain.Handle,
	networkProvider net.Provider,
	tssConfig *tss.Config,
) *Node {
	return &Node{
		chain:           chain,
		networkProvider: networkProvider,
		tssConfig:       tssConfig,
	}
}

// AnnounceSignerPresence triggers the announce protocol in order to signal
// signer presence and gather information about other signers.
func (n *Node) AnnounceSignerPresence(
	ctx context.Context,
	operatorPublicKey *operator.PublicKey,
	keepID chain.ID,
	keepMemberIDs []chain.ID,
) ([]tss.MemberID, error) {
	broadcastChannel, err := n.networkProvider.BroadcastChannelFor(keepID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize broadcast channel: [%v]", err)
	}

	tss.RegisterUnmarshalers(broadcastChannel)

	if err := broadcastChannel.SetFilter(
		createAddressFilter(
			keepMemberIDs,
			n.chain.PublicKeyToOperatorID,
		),
	); err != nil {
		return nil, fmt.Errorf("failed to set broadcast channel filter: [%v]", err)
	}

	return tss.AnnounceProtocol(
		ctx,
		operatorPublicKey,
		keepID,
		keepMemberIDs,
		broadcastChannel,
		n.chain.PublicKeyToOperatorID,
	)
}

func createAddressFilter(
	keepMemberIDs []chain.ID,
	publicKeyToOperatorIDFunc func(*cecdsa.PublicKey) chain.ID,
) net.BroadcastChannelFilter {
	operatorAuthorizations := make(map[string]bool, len(keepMemberIDs))
	for _, keepMemberID := range keepMemberIDs {
		operatorAuthorizations[keepMemberID.String()] = true
	}

	return func(authorPublicKey *cecdsa.PublicKey) bool {
		authorAddress := publicKeyToOperatorIDFunc(authorPublicKey)

		_, isAuthorized := operatorAuthorizations[authorAddress.String()]

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
	keep chain.BondedECDSAKeepHandle,
	members []chain.ID,
	keepsRegistry *registry.Keeps,
) (*tss.ThresholdSigner, error) {
	memberID := tss.MemberIDFromPublicKey(operatorPublicKey)
	preParamsBox := params.NewBox(n.tssParamsPool.get())

	attemptCounter := 0
	for {
		attemptCounter++

		logger.Infof(
			"signer generation for keep [%s]; attempt [%v]",
			keep.ID(),
			attemptCounter,
		)

		isActive, err := keep.IsActive()
		if err != nil {
			logger.Warningf(
				"could not check if keep [%s] is still active: [%v]",
				keep.ID(),
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
			keep.ID(),
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
			keep.ID().String(),
			memberID,
			memberIDs,
			uint(len(memberIDs)-1),
			n.networkProvider,
			n.chain.Signing().PublicKeyToAddress,
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
		err = keepsRegistry.SnapshotSigner(keep.ID(), signer)
		if err != nil {
			return nil, fmt.Errorf(
				"could not make snapshot of signer for keep [%s]: [%v]",
				keep.ID(),
				err,
			)
		}

		// Serialize and submit public key to the keep.
		//
		// We don't retry in case of an error although the specific chain
		// implementation may implement its own retry policy. This action
		// should never fail and if it failed, something terrible happened.
		publicKey, err := chain.SerializePublicKey(signer.PublicKey())
		if err != nil {
			return nil, fmt.Errorf("failed to serialize public key: [%v]", err)
		}

		err = keep.SubmitKeepPublicKey(publicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to submit public key: [%v]", err)
		}

		go n.monitorKeepPublicKeySubmission(keep, publicKey)

		return signer, nil // key generation succeeded.
	}
}

// CalculateSignature calculates a signature over a digest with threshold
// signer and publishes the result to the keep associated with the signer.
//
// The attempt for generating and publishing signature is retried on failure
// until the provided context is done.
func (n *Node) CalculateSignature(
	ctx context.Context,
	keep chain.BondedECDSAKeepHandle,
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
		signature, err := signer.CalculateSignature(
			ctx,
			digest[:],
			n.networkProvider,
			n.chain.Signing().PublicKeyToAddress,
		)
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
		return n.publishSignature(ctx, keep, digest, signature)
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
	keep chain.BondedECDSAKeepHandle,
	digest [32]byte,
	signature *ecdsa.Signature,
) error {
	n.waitSignaturePublicationDelay(keep)

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
		isActive, err := keep.IsActive()
		if err != nil {
			logger.Errorf(
				"failed to verify if keep [%s] is still active: [%v]",
				keep.ID(),
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
		isAwaitingSignature, err := keep.IsAwaitingSignature(digest)
		if err != nil {
			logger.Errorf(
				"failed to verify if keep [%s] is still awaiting signature: [%v]",
				keep.ID(),
				err,
			)
			time.Sleep(retryDelay) // TODO: #413 Replace with backoff.
			continue
		}

		// Someone submitted the signature, it was accepted by the keep,
		// and there are enough confirmations from the chain.
		// We are fine, leaving.
		if !isAwaitingSignature && n.confirmSignature(keep, digest) {
			return nil
		}

		logger.Infof(
			"publishing signature for keep [%s]; attempt [%v]",
			keep.ID(),
			attemptCounter,
		)

		if submissionErr := keep.SubmitSignature(signature); submissionErr != nil {
			isAwaitingSignature, err := keep.IsAwaitingSignature(digest)
			if err != nil {
				logger.Errorf(
					"failed to verify if keep [%s] is still awaiting signature: [%v]",
					keep.ID(),
					err,
				)
				time.Sleep(retryDelay) // TODO: #413 Replace with backoff.
				continue
			}

			// Check if we failed because someone else submitted in the meantime
			// or because something wrong happened with our transaction.
			// If someone else submitted in the meantime, wait for enough
			// confirmations from the chain before making a decision about
			// leaving the submission process.
			if !isAwaitingSignature && n.confirmSignature(keep, digest) {
				return nil
			}

			// Our public key submission transaction failed. We are going to
			// wait for some time and then retry from the beginning.
			logger.Errorf(
				"failed to submit signature for keep [%s]: [%v]; "+
					"will retry after 1 minute",
				keep.ID(),
				submissionErr,
			)
			time.Sleep(1 * time.Minute)
			continue
		}

		if !(n.waitForSignature(keep, digest) && n.confirmSignature(keep, digest)) {
			time.Sleep(retryDelay) // TODO: #413 Replace with backoff.
			continue
		}

		return nil
	}
}

// waitSignaturePublicationDelay waits a certain amount of time appropriately
// for the given signer index to avoid all signers publishing the same signature
// for given keep at the same time.
func (n *Node) waitSignaturePublicationDelay(keep chain.BondedECDSAKeepHandle) {
	signerIndex, err := keep.OperatorIndex()
	if err != nil {
		logger.Errorf(
			"could not determine signature publication delay for keep [%s]: "+
				"[%v]; the signature publication will not be delayed",
			keep.ID(),
			err,
		)
		return
	}

	// just in case this function is not invoked in the right context
	if signerIndex < 0 {
		logger.Errorf(
			"could not determine signature publication delay for keep [%s], "+
				"signer index is less than zero; the signature publication "+
				"will not be delayed",
			keep.ID(),
		)
		return
	}

	delay := time.Duration(signerIndex) * signaturePublicationDelayStep

	logger.Infof(
		"waiting [%v] before publishing signature for keep [%s]",
		delay,
		keep.ID(),
	)

	time.Sleep(delay)
}

func (n *Node) waitForSignature(
	keep chain.BondedECDSAKeepHandle,
	digest [32]byte,
) bool {
	const waitTimeout = 30 * time.Minute
	const checkTick = 1 * time.Minute

	ctx, cancelCtx := context.WithTimeout(context.Background(), waitTimeout)
	defer cancelCtx()

	checkTicker := time.NewTicker(checkTick)
	defer checkTicker.Stop()

	logger.Infof(
		"waiting for signature for keep [%s] to appear on-chain",
		keep.ID(),
	)

	for {
		select {
		case <-checkTicker.C:
			// We check the signature periodically instead of relying on
			// incoming events. The main motivation is that events could not be
			// trusted here because they may come from a forked chain or
			// the same event can be delivered multiple times.
			isAwaitingSignature, err := keep.IsAwaitingSignature(digest)
			if err != nil {
				logger.Errorf(
					"failed to perform signature check while waiting "+
						"for signature for keep [%s]: [%v]",
					keep.ID(),
					err,
				)
				continue
			}

			if !isAwaitingSignature {
				logger.Infof(
					"signature for keep [%s] appeared on-chain",
					keep.ID(),
				)
				return true
			}
		case <-ctx.Done():
			logger.Errorf(
				"signature for keep [%s] has not appeared on the chain "+
					"after [%v] from submitting it",
				keep.ID(),
				waitTimeout,
			)
			return false
		}
	}
}

func (n *Node) confirmSignature(
	keep chain.BondedECDSAKeepHandle,
	digest [32]byte,
) bool {
	logger.Infof(
		"confirming on-chain signature submission for keep [%s]",
		keep.ID(),
	)

	currentBlock, err := n.chain.BlockCounter().CurrentBlock()
	if err != nil {
		logger.Errorf(
			"could not get current block while confirming "+
				"signature submission for keep [%s]: [%v]",
			keep.ID(),
			err,
		)
		return false
	}

	isSignatureConfirmed, err := ethlike.WaitForBlockConfirmations(
		n.chain.BlockCounter(),
		currentBlock,
		blockConfirmations,
		func() (bool, error) {
			isAwaitingSignature, err := keep.IsAwaitingSignature(digest)
			if err != nil {
				return false, err
			}

			return !isAwaitingSignature, nil
		},
	)
	if err != nil {
		logger.Errorf(
			"could not confirm signature submission for keep [%s]: [%v]",
			keep.ID(),
			err,
		)
		return false
	}

	if !isSignatureConfirmed {
		logger.Errorf(
			"signature submission for keep [%s] not confirmed; "+
				"trying to submit the signature again",
			keep.ID(),
		)
		return false
	}

	logger.Infof(
		"signature for keep [%s] successfully submitted "+
			"and confirmed on-chain",
		keep.ID(),
	)

	return true
}

// monitorKeepPublicKeySubmission observes the chain until either the first
// conflicting public key is published or until keep established public key
// or until the key generation timed out. It also tries to re-submit the public
// key if it has not been submitted in the given time frame as a way to protect
// against chain forks (i.e. transaction mined in a fork that has been dropped).
//
// When event informing about conflicting public key is emitted from the chain,
// monitoring stops immediately. Even if this event was emitted in a minority
// fork, it is clear that the operator who submitted the conflicting key is
// dishonest and it is better to abandon this keep.
//
// Function checks the chain periodically and inspects whether the public key
// has been registered on the chain. If the key is there, the function waits
// for the required number of confirmations and repeats the check. If the public
// key for the keep is still there, the monitoring exits successfully.
// In the case when public key for the keep is not yet established during the
// periodic check or it is gone after waiting for a certain number of
// confirmations (chain reorganization), this function will attempt to submit
// the public key again.
func (n *Node) monitorKeepPublicKeySubmission(
	keep chain.BondedECDSAKeepHandle,
	publicKey [64]byte,
) {
	conflictingPublicKey := make(chan *chain.ConflictingPublicKeySubmittedEvent)

	subscriptionConflictingPublicKey, err := keep.OnConflictingPublicKeySubmitted(
		func(event *chain.ConflictingPublicKeySubmittedEvent) {
			conflictingPublicKey <- event
		},
	)
	if err != nil {
		logger.Errorf(
			"failed on watching conflicting public key event for keep [%s]: [%v]",
			keep.ID(),
			err,
		)
	}

	defer subscriptionConflictingPublicKey.Unsubscribe()

	pubkeyChecksCounter := 0
	// There is no way to determine whether keep waits for public key submission
	// from this client or some other client. Given that the consequences are
	// not that serious as for not submitting a signature and to minimize gas
	// expenditure in case the current client's pub key has been properly
	// registered and we are waiting for someone else, we retry only three times.
	const maxPubkeyChecksCount = 2
	// All three operators need to submit public key to the chain so we are
	// less aggressive with check ticks than in case of signature submission
	// where only one signature is enough.
	const pubkeyCheckTick = 60 * time.Minute

	pubkeyCheckTicker := time.NewTicker(pubkeyCheckTick)
	defer pubkeyCheckTicker.Stop()

	for {
		select {
		case event := <-conflictingPublicKey:
			// Even in the case of a fork, a transaction with a conflicting
			// public key submitted to the chain had to be signed with the
			// operator's pubkey. For an honest operator, this situation should
			// never happen, and if it happened, it is better to exit the
			// monitoring immediately and do not re-attempt to submit public
			// key for this node. It is clear that something is wrong with the
			// operator that published the conflicting key and it is safer to
			// abandon this keep.
			logger.Errorf(
				"member [%x] has submitted conflicting public key for keep [%s]: [%x]",
				event.SubmittingMember,
				keep.ID(),
				event.ConflictingPublicKey,
			)
			return
		case <-pubkeyCheckTicker.C:
			pubkeyChecksCounter++

			if pubkeyChecksCounter > maxPubkeyChecksCount {
				logger.Infof(
					"monitoring of public key submission for keep [%s] "+
						"has been cancelled because maximum checks count [%v] "+
						"has been reached",
					keep.ID(),
					maxPubkeyChecksCount,
				)
				return
			}

			logger.Infof(
				"confirming on-chain public key submission for keep [%s]",
				keep.ID(),
			)

			// We check the public key periodically instead of relying on
			// incoming events. The main motivation is that events could not be
			// trusted here because they may come from a forked chain or
			// the same event can be delivered multiple times.
			keepPublicKey, err := keep.GetPublicKey()
			if err != nil {
				logger.Errorf(
					"failed to get keep public key during "+
						"public key submission monitoring for keep [%s]: [%v]",
					keep.ID(),
					err,
				)
				continue
			} else {
				if len(keepPublicKey) > 0 {
					currentBlock, err := n.chain.BlockCounter().CurrentBlock()
					if err != nil {
						logger.Errorf(
							"failed to get the current block while "+
								"performing public key submission confirmation "+
								"for keep [%s]: [%v]",
							keep.ID(),
							err,
						)
						continue
					}

					isConfirmed, err := ethlike.WaitForBlockConfirmations(
						n.chain.BlockCounter(),
						currentBlock,
						blockConfirmations,
						func() (bool, error) {
							key, err := keep.GetPublicKey()
							if err != nil {
								return false, err
							}

							return len(key) > 0, nil
						},
					)
					if err != nil {
						logger.Errorf(
							"failed to perform keep public key "+
								"confirmation during public key submission "+
								"monitoring for keep [%s]: [%v]",
							keep.ID(),
							err,
						)
						continue
					}

					if isConfirmed {
						logger.Infof(
							"public key [%x] for keep [%s] successfully "+
								"submitted and confirmed on-chain",
							keepPublicKey,
							keep.ID(),
						)
						return
					}
				}
			}

			logger.Infof(
				"keep [%s] still does not have a confirmed public key; "+
					"re-submitting public key [%x]",
				keep.ID(),
				publicKey,
			)

			err = keep.SubmitKeepPublicKey(publicKey)
			if err != nil {
				logger.Errorf(
					"keep [%s] still does not have a confirmed public key "+
						"and resubmission by this member failed with: [%v]",
					keep.ID(),
					err,
				)
				return
			}
		}
	}
}
