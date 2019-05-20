// Package client defines Keep tECDSA client.
package client

import (
	crand "crypto/rand"
	"fmt"
	"sync"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/keep-network/keep-tecdsa/pkg/btc"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/sign"
)

// Client holds blockchain specific configuration, interfaces to interact with the
// blockchain and a map of signers for given keeps.
type Client struct {
	ethereumChain    eth.Interface
	bitcoinNetParams *chaincfg.Params
	keepsSigners     sync.Map //<keepAddress, signer>
}

// NewClient creates new client.
func NewClient(
	ethereumChain eth.Interface,
	bitcoinNetParams *chaincfg.Params,
) *Client {
	return &Client{
		ethereumChain:    ethereumChain,
		bitcoinNetParams: bitcoinNetParams,
	}
}

// Initialize initializes the client with rules related to events handling.
func (c *Client) Initialize() {
	c.ethereumChain.OnECDSAKeepCreated(func(event *eth.ECDSAKeepCreatedEvent) {
		fmt.Printf("New ECDSA Keep created [%+v]\n", event)

		go c.GenerateSignerForKeep(event.KeepAddress.String())
	})
}

// GenerateSignerForKeep generates a new signer with ECDSA key pair and calculates
// bitcoin specific P2WPKH address based on signer's public key. It stores the
// signer in a map assigned to a provided keep address.
func (c *Client) GenerateSignerForKeep(keepAddress string) error {
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
