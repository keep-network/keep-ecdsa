// Package tecdsa defines Keep tECDSA client.
package tecdsa

import (
	crand "crypto/rand"
	"fmt"
	"os"
	"sync"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/keep-network/keep-tecdsa/pkg/btc"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/sign"
)

// client holds blockchain specific configuration, interfaces to interact with the
// blockchain and a map of signers for given keeps.
type client struct {
	ethereumChain    eth.Interface
	bitcoinNetParams *chaincfg.Params
	keepsSigners     sync.Map //<keepAddress, signer>
}

// Initialize initializes the tECDSA client with rules related to events handling.
func Initialize(
	ethereumChain eth.Interface,
	bitcoinNetParams *chaincfg.Params,
) error {
	client := &client{
		ethereumChain:    ethereumChain,
		bitcoinNetParams: bitcoinNetParams,
	}

	ethereumChain.OnECDSAKeepCreated(func(event *eth.ECDSAKeepCreatedEvent) {
		fmt.Printf("New ECDSA Keep created [%+v]\n", event)

		go func() {
			if err := client.generateSignerForKeep(event.KeepAddress); err != nil {
				fmt.Fprintf(os.Stderr, "signer generation failed: [%s]", err)
			}
		}()

		ethereumChain.OnSignatureRequested(
			event.KeepAddress,
			func(signatureRequestedEvent *eth.SignatureRequestedEvent) {
				fmt.Printf("New signature requested [%+x]\n", signatureRequestedEvent)

				go func() {
					err := client.calculateSignatureForKeep(
						event.KeepAddress,
						signatureRequestedEvent.Digest,
					)

					if err != nil {
						fmt.Fprintf(os.Stderr, "signature calculation failed: [%s]", err)
					}
				}()
			},
		)
	})

	return nil
}

// generateSignerForKeep generates a new signer with ECDSA key pair and calculates
// bitcoin specific P2WPKH address based on signer's public key. It stores the
// signer in a map assigned to a provided keep address.
func (c *client) generateSignerForKeep(keepAddress eth.KeepAddress) error {
	signer, err := generateSigner()

	// Calculate bitcoin P2WPKH address.
	// TODO: We don't need to calculate the address here. It should be done in
	// tBTC utils tool. But we do it for development simplification.
	btcAddress, err := btc.PublicKeyToWitnessPubKeyHashAddress(
		signer.PublicKey(),
		c.bitcoinNetParams,
	)
	if err != nil {
		return fmt.Errorf("p2wpkh address conversion failed: [%s]", err)
	}

	// Publish signer's public key on ethereum blockchain in a specific keep
	// contract.
	serializedPublicKey, err := eth.SerializePublicKey(signer.PublicKey())
	if err != nil {
		return fmt.Errorf("public key serialization failed: [%s]", err)
	}

	err = c.ethereumChain.SubmitKeepPublicKey(
		keepAddress,
		serializedPublicKey,
	)
	if err != nil {
		return fmt.Errorf("public key submission failed: [%s]", err)
	}

	fmt.Printf(
		"Signer for keep [%s] initialized with Bitcoin P2WPKH address: [%s]\n",
		keepAddress.String(),
		btcAddress,
	)

	// Store the signer in a map, with the keep address as a key. Keep address
	// is converted to a hexadecimal string prefiexed with `0x`.
	c.keepsSigners.Store(keepAddress.String(), signer)

	return nil
}

func generateSigner() (*sign.Signer, error) {
	privateKey, err := sign.GenerateKey(crand.Reader)
	if err != nil {
		return nil, fmt.Errorf("private key generation failed: [%s]", err)
	}

	return sign.NewSigner(privateKey), nil
}

func (c *client) calculateSignatureForKeep(keepAddress eth.KeepAddress, digest [32]byte) error {
	signer, ok := c.keepsSigners.Load(keepAddress.String())
	if !ok {
		return fmt.Errorf("cannot load signer for keep: [%s]", keepAddress.String())
	}

	signature, err := signer.(*sign.Signer).CalculateSignature(
		crand.Reader,
		digest[:],
	)

	fmt.Printf("Signature calculated: [%+v]\n", signature)

	err = c.ethereumChain.SubmitSignature(
		keepAddress,
		digest,
		signature,
	)
	if err != nil {
		return fmt.Errorf("signature submission failed: [%s]", err)
	}

	return nil
}
