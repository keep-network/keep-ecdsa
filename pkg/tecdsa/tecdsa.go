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
) {
	client := &client{
		ethereumChain:    ethereumChain,
		bitcoinNetParams: bitcoinNetParams,
	}

	ethereumChain.OnECDSAKeepCreated(func(event *eth.ECDSAKeepCreatedEvent) {
		fmt.Printf("New ECDSA Keep created [%+v]\n", event)

		go func() {
			if err := client.generateSignerForKeep(event.KeepAddress.String()); err != nil {
				fmt.Fprintf(os.Stderr, "signer generation failed: [%s]", err)
			}
		}()
	})

	// // ecdsaKeepFactoryContractAddress, err := config.ContractAddress(ECDSAKeepFactoryContractName)
	// if err != nil {
	// 	return nil, err
	// }
	// ecdsaKeepFactoryContract, err := abi.NewECDSAKeepFactory(
	// 	ecdsaKeepFactoryContractAddress,
	// 	client,
	// )
	// if err != nil {
	// 	return nil, err
	// }

	// ethereumChain.OnECDSAKeepSignatureRequest(func(event *eth.ECDSAKeepSignatureRequestEvent) {
	// 	fmt.Printf("New signature requested [%+v]\n", event.Digest)
	// })
}

// generateSignerForKeep generates a new signer with ECDSA key pair and calculates
// bitcoin specific P2WPKH address based on signer's public key. It stores the
// signer in a map assigned to a provided keep address.
func (c *client) generateSignerForKeep(keepAddress string) error {
	signer, err := generateSigner()

	// Calculate bitcoin P2WPKH address.
	btcAddress, err := btc.PublicKeyToWitnessPubKeyHashAddress(
		signer.PublicKey(),
		c.bitcoinNetParams,
	)
	if err != nil {
		return fmt.Errorf("p2wpkh address conversion failed: [%s]", err)
	}

	// TODO: Publish it on ethereum chain in the specific keep contract.
	fmt.Printf(
		"Signer for keep [%s] initialized with Bitcoin P2WPKH address [%s]\n",
		keepAddress,
		btcAddress,
	)

	c.keepsSigners.Store(keepAddress, signer)

	return nil
}

func generateSigner() (*sign.Signer, error) {
	privateKey, err := sign.GenerateKey(crand.Reader)
	if err != nil {
		return nil, fmt.Errorf("private key generation failed: [%s]", err)
	}

	return sign.NewSigner(privateKey), nil
}
