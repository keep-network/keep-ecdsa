// Package tecdsa defines Keep tECDSA protocol.
package tecdsa

import (
	"fmt"
	"sync"
	"time"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss"
	"github.com/keep-network/keep-tecdsa/pkg/net"
)

var logger = log.Logger("keep-tecdsa")

// TECDSA holds interfaces to interact with the blockchain and network messages
// transport layer.
type TECDSA struct {
	ethereumChain   eth.Handle
	networkProvider net.Provider
	tssParamsPool   *tssPreParamsPool
}

// NewTECDSA initializes TECDSA struct with provided ethereum chain interface and
// network provider. It also initializes TSS Pre-Parameters pool. But does not
// start parameters generation. This should be called separately.
func NewTECDSA(
	ethereumChain eth.Handle,
	networkProvider net.Provider,
) *TECDSA {
	return &TECDSA{
		ethereumChain:   ethereumChain,
		networkProvider: networkProvider,
		tssParamsPool: &tssPreParamsPool{
			poolMutex: &sync.Mutex{},
			pool:      []*keygen.LocalPreParams{},
		},
	}
}

// RegisterForSignEvents registers for signature requested events emitted by
// specific keep contract.
func (t *TECDSA) RegisterForSignEvents(
	keepAddress eth.KeepAddress,
	signer *tss.ThresholdSigner,
) {
	t.ethereumChain.OnSignatureRequested(
		keepAddress,
		func(signatureRequestedEvent *eth.SignatureRequestedEvent) {
			logger.Infof(
				"new signature requested from keep [%s] for digest: [%+x]",
				keepAddress.String(),
				signatureRequestedEvent.Digest,
			)

			// TODO: Temp Sync
			tss.SigningSync.Add(1)
			time.Sleep(1 * time.Second)

			go func() {
				err := t.calculateSignatureForKeep(
					keepAddress,
					signer,
					signatureRequestedEvent.Digest,
				)

				if err != nil {
					logger.Errorf("signature calculation failed: [%v]", err)
				}
			}()
		},
	)
}

// GenerateSignerForKeep generates a new signer with ECDSA key pair. It publishes
// the signer's public key to the keep.
func (t *TECDSA) GenerateSignerForKeep(
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

	memberID := tss.MemberID(t.ethereumChain.Address().String())

	signer, err := tss.GenerateThresholdSigner(
		keepAddress.Hex(),
		memberID,
		groupMemberIDs,
		uint(len(keepMembers)-1),
		t.networkProvider,
		t.tssParamsPool.get(),
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
		err = t.ethereumChain.SubmitKeepPublicKey(
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

func (t *TECDSA) calculateSignatureForKeep(
	keepAddress eth.KeepAddress,
	signer *tss.ThresholdSigner,
	digest [32]byte,
) error {
	signature, err := signer.CalculateSignature(digest[:], t.networkProvider)
	if err != nil {
		return fmt.Errorf("failed to calculate signature: [%v]", err)
	}

	logger.Debugf(
		"signature calculated:\nr: [%#x]\ns: [%#x]\nrecovery ID: [%d]\n",
		signature.R,
		signature.S,
		signature.RecoveryID,
	)

	err = t.ethereumChain.SubmitSignature(keepAddress, digest, signature)
	if err != nil {
		return fmt.Errorf("failed to submit signature: [%v]", err)
	}

	logger.Infof("submitted signature for digest: [%+x]", digest)

	return nil
}
