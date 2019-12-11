// Package tecdsa defines Keep tECDSA protocol.
package tecdsa

import (
	crand "crypto/rand"
	"fmt"

	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss"
)

var logger = log.Logger("keep-tecdsa")

// TECDSA holds an interface to interact with the blockchain.
type TECDSA struct {
	EthereumChain eth.Handle
}

// RegisterForSignEvents registers for signature requested events emitted by
// specific keep contract.
func (t *TECDSA) RegisterForSignEvents(
	keepAddress eth.KeepAddress,
	signer *tss.ThresholdSigner,
) {
	t.EthereumChain.OnSignatureRequested(
		keepAddress,
		func(signatureRequestedEvent *eth.SignatureRequestedEvent) {
			logger.Debugf(
				"new signature requested from keep [%s] for digest: [%+x]",
				keepAddress.String(),
				signatureRequestedEvent.Digest,
			)

			go func() {
				// TODO: Replace it with Threshold Signer
				// err := t.calculateSignatureForKeep(
				// 	keepAddress,
				// 	nil,
				// 	signatureRequestedEvent.Digest,
				// )

				// if err != nil {
				// 	logger.Errorf("signature calculation failed: [%v]", err)
				// }
			}()
		},
	)
}

// GenerateSignerForKeep generates a new signer with ECDSA key pair. It publishes
// the signer's public key to the keep.
func (t *TECDSA) GenerateSignerForKeep(
	keepAddress eth.KeepAddress,
) (*ecdsa.Signer, error) {
	signer, err := generateSigner()

	logger.Debugf(
		"generated signer with public key: [%x]",
		signer.PublicKey().Marshal(),
	)

	// Publish signer's public key on ethereum blockchain in a specific keep
	// contract.
	serializedPublicKey, err := eth.SerializePublicKey(signer.PublicKey())
	if err != nil {
		return nil, fmt.Errorf("failed to serialize public key: [%v]", err)
	}

	err = t.EthereumChain.SubmitKeepPublicKey(
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

func generateSigner() (*ecdsa.Signer, error) {
	privateKey, err := ecdsa.GenerateKey(crand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: [%v]", err)
	}

	return ecdsa.NewSigner(privateKey), nil
}

func (t *TECDSA) calculateSignatureForKeep(
	keepAddress eth.KeepAddress,
	signer *ecdsa.Signer,
	digest [32]byte,
) error {
	signature, err := signer.CalculateSignature(crand.Reader, digest[:])

	logger.Debugf(
		"signature calculated:\nr: [%#x]\ns: [%#x]\nrecovery ID: [%d]\n",
		signature.R,
		signature.S,
		signature.RecoveryID,
	)

	err = t.EthereumChain.SubmitSignature(keepAddress, digest, signature)
	if err != nil {
		return fmt.Errorf("failed to submit signature: [%v]", err)
	}

	logger.Infof("submitted signature for digest: [%+x]", digest)

	return nil
}
