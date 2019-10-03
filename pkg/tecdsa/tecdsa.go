// Package tecdsa defines Keep tECDSA client.
package tecdsa

import (
	crand "crypto/rand"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-tecdsa/pkg/registry"
)

var logger = log.Logger("keep-tecdsa")

// client holds blockchain specific configuration, interfaces to interact with the
// blockchain and a map of signers for given keeps.
type client struct {
	ethereumChain    eth.Handle
	bitcoinNetParams *chaincfg.Params
	keepsRegistry    *registry.Keeps
}

// Initialize initializes the tECDSA client with rules related to events handling.
func Initialize(
	ethereumChain eth.Handle,
	bitcoinNetParams *chaincfg.Params,
	persistence persistence.Handle,
) {
	keepsRegistry := registry.NewKeepsRegistry(persistence)

	client := &client{
		ethereumChain:    ethereumChain,
		bitcoinNetParams: bitcoinNetParams,
		keepsRegistry:    keepsRegistry,
	}

	// Load current keeps signers from storage and register for signing events.
	keepsRegistry.LoadExistingKeeps()

	for _, keepAddress := range keepsRegistry.GetKeepsAddresses() {
		client.registerForSignEvents(keepAddress)
	}

	// Watch for new keeps creation.
	ethereumChain.OnECDSAKeepCreated(func(event *eth.ECDSAKeepCreatedEvent) {
		logger.Infof(
			"new keep created with address: [%s]",
			event.KeepAddress.String(),
		)

		if event.ContainsMember(ethereumChain.Address()) {
			go func() {
				if err := client.generateSignerForKeep(event.KeepAddress); err != nil {
					logger.Errorf("signer generation failed: [%v]", err)
				}

				client.registerForSignEvents(event.KeepAddress)
			}()
		}
	})

	// Register client as a candidate member for keep.
	if err := ethereumChain.RegisterAsMemberCandidate(); err != nil {
		logger.Errorf("failed to register member: [%v]", err)
	}
}

// registerForSignEvents registers for signature requested events emitted by
// specific keep contract.
func (c *client) registerForSignEvents(keepAddress eth.KeepAddress) {
	c.ethereumChain.OnSignatureRequested(
		keepAddress,
		func(signatureRequestedEvent *eth.SignatureRequestedEvent) {
			logger.Debugf(
				"new signature requested for digest: [%+x]",
				signatureRequestedEvent.Digest,
			)

			go func() {
				err := c.calculateSignatureForKeep(
					keepAddress,
					signatureRequestedEvent.Digest,
				)

				if err != nil {
					logger.Errorf("signature calculation failed: [%v]", err)
				}
			}()
		},
	)
}

// generateSignerForKeep generates a new signer with ECDSA key pair and calculates
// bitcoin specific P2WPKH address based on signer's public key. It registers
// signer in a keeps registry for a given keep address.
func (c *client) generateSignerForKeep(keepAddress eth.KeepAddress) error {
	signer, err := generateSigner()

	logger.Debugf(
		"generated signer with public key: [x: [%x], y: [%x]]",
		signer.PublicKey().X,
		signer.PublicKey().Y,
	)

	// Publish signer's public key on ethereum blockchain in a specific keep
	// contract.
	serializedPublicKey, err := eth.SerializePublicKey(signer.PublicKey())
	if err != nil {
		return fmt.Errorf("failed to serialize public key: [%v]", err)
	}

	err = c.ethereumChain.SubmitKeepPublicKey(
		keepAddress,
		serializedPublicKey,
	)
	if err != nil {
		return fmt.Errorf("failed to submit public key: [%v]", err)
	}

	logger.Infof(
		"initialized signer for keep [%s]",
		keepAddress.String(),
	)

	// Store the signer in a map, with the keep address as a key.
	c.keepsRegistry.RegisterSigner(keepAddress, signer)

	return nil
}

func generateSigner() (*ecdsa.Signer, error) {
	privateKey, err := ecdsa.GenerateKey(crand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: [%v]", err)
	}

	return ecdsa.NewSigner(privateKey), nil
}

func (c *client) calculateSignatureForKeep(keepAddress eth.KeepAddress, digest [32]byte) error {
	signer, err := c.keepsRegistry.GetSigner(keepAddress)
	if err != nil {
		return fmt.Errorf("failed to get group for keep [%s]: [%v]", keepAddress.String(), err)
	}

	signature, err := signer.CalculateSignature(
		crand.Reader,
		digest[:],
	)

	logger.Debugf(
		"signature calculated:\nr: [%#x]\ns: [%#x]\nrecovery ID: [%d]\n",
		signature.R,
		signature.S,
		signature.RecoveryID,
	)

	err = c.ethereumChain.SubmitSignature(
		keepAddress,
		digest,
		signature,
	)
	if err != nil {
		return fmt.Errorf("failed to submit signature: [%v]", err)
	}

	return nil
}
