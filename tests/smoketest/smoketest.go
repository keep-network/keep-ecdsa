package smoketest

import (
	"bytes"
	cecdsa "crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/ethereum"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/gen/abi"
	"github.com/keep-network/keep-tecdsa/pkg/utils/byteutils"
)

var keepOwnerPrivateKeyHex = "bd03a0aa0b96c5cff1accafdc806aa7f655b6a9a13aeb79f4669c9cfad1eb265"

// Execute runs an ECDSA event smoke test. It tests if an event is emitted after
// new ECDSA keep is requested.
func Execute(config *ethereum.Config) error {
	chainAPI, err := ethereum.Connect(config)
	if err != nil {
		return err
	}

	client, err := initializeEthereumClient(config)
	if err != nil {
		return err
	}

	// Setup connection to ECDSA Keep Factory contract.
	ecdsaKeepFactory, err := initializeECDSAKeepFactory(config, client)
	if err != nil {
		return err
	}

	transactorOpts, err := createTransactorOpts(config.PrivateKey)
	if err != nil {
		return err
	}

	// We create a keep for specific owner, different than caller.
	keepOwnerTransactorOpts, err := createTransactorOpts(keepOwnerPrivateKeyHex)
	if err != nil {
		return err
	}

	// Register for Keep Created event.
	keepCreatedEventChan := make(chan *eth.ECDSAKeepCreatedEvent)

	keepCreatedHandle := func(event *eth.ECDSAKeepCreatedEvent) {
		keepCreatedEventChan <- event
	}

	keepCreatedSubscription, err := chainAPI.OnECDSAKeepCreated(keepCreatedHandle)
	defer keepCreatedSubscription.Unsubscribe()
	if err != nil {
		return err
	}

	// Request a new keep creation.
	groupSize := big.NewInt(10)
	honestThreshold := big.NewInt(5)
	keepOwnerAddress := keepOwnerTransactorOpts.From

	transaction, err := ecdsaKeepFactory.OpenKeep(
		transactorOpts,
		groupSize,
		honestThreshold,
		keepOwnerAddress,
	)
	if err != nil {
		return fmt.Errorf("call to contract failed: [%s]", err)
	}
	fmt.Printf(
		"New keep requested, transaction hash: [%s]\n",
		transaction.Hash().Hex(),
	)

	// Wait for event emission.
	keepCreatedEvent := <-keepCreatedEventChan

	keepAddress := keepCreatedEvent.KeepAddress

	// Validate received event.
	if !common.IsHexAddress(keepAddress.String()) {
		return fmt.Errorf("invalid hex address: [%v]", keepAddress)
	}

	fmt.Printf("ECDSA keep built with address: [%s]\n", keepAddress.String())

	// SIGN

	// Setup connection to ECDSA Keep contract.
	ecdsaKeep, err := initializeECDSAKeep(keepAddress, client)
	if err != nil {
		return err
	}

	// Register for Signature Requested event.
	signatureRequestedEventChan := make(chan *eth.SignatureRequestedEvent)

	signHandle := func(event *eth.SignatureRequestedEvent) {
		signatureRequestedEventChan <- event
	}

	signSubscription, err := chainAPI.OnSignatureRequested(keepAddress, signHandle)
	defer signSubscription.Unsubscribe()
	if err != nil {
		return err
	}

	// Register for Signature Submitted event.
	signatureSubmittedEventChan := make(chan *abi.ECDSAKeepSignatureSubmitted)

	signatureSubmittedSubscription, err := ecdsaKeep.WatchSignatureSubmitted(nil, signatureSubmittedEventChan)
	defer signatureSubmittedSubscription.Unsubscribe()
	if err != nil {
		return err
	}

	// Sign digest.
	hash, err := hex.DecodeString("54a6483b8aca55c9df2a35baf71d9965ddfd623468d81d51229bd5eb7d1e1c1b")
	if err != nil {
		return err
	}

	digest, err := byteutils.BytesTo32Byte(hash)
	if err != nil {
		return err
	}

	transaction, err = ecdsaKeep.Sign(keepOwnerTransactorOpts, digest)
	if err != nil {
		return fmt.Errorf("call to contract failed: [%s]", err)
	}

	fmt.Printf(
		"new signature requested, transaction hash: [%s]\n",
		transaction.Hash().Hex(),
	)

	// Wait for events emission.
	signatureRequestedEvent := <-signatureRequestedEventChan
	signatureSubmittedEvent := <-signatureSubmittedEventChan

	if !bytes.Equal(signatureRequestedEvent.Digest[:], digest[:]) {
		return fmt.Errorf(
			"digests don't match\nexpected: [%x]\nactual:   [%x]\n",
			digest,
			signatureRequestedEvent.Digest,
		)
	}

	keepPublicKey, err := ecdsaKeep.GetPublicKey(nil)
	if err != nil {
		return err
	}

	ecdsaPubicKey := &cecdsa.PublicKey{
		Curve: secp256k1.S256(),
		X:     new(big.Int).SetBytes(keepPublicKey[:32]),
		Y:     new(big.Int).SetBytes(keepPublicKey[32:]),
	}

	if !cecdsa.Verify(
		ecdsaPubicKey,
		signatureSubmittedEvent.Digest[:],
		new(big.Int).SetBytes(signatureSubmittedEvent.R[:]),
		new(big.Int).SetBytes(signatureSubmittedEvent.S[:]),
	) {
		return fmt.Errorf("signature is invalid")
	}

	// Log received event.
	fmt.Printf("Received Signature:\nr: [%#x]\ns: [%#x]\nRecovery ID: [%d]",
		signatureSubmittedEvent.R,
		signatureSubmittedEvent.S,
		signatureSubmittedEvent.RecoveryID,
	)

	return nil
}

func initializeEthereumClient(config *ethereum.Config) (*ethclient.Client, error) {
	client, err := ethclient.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func initializeECDSAKeepFactory(config *ethereum.Config, client *ethclient.Client) (*abi.ECDSAKeepFactory, error) {
	ecdsaKeepFactoryContractAddress, err := config.ContractAddress(ethereum.ECDSAKeepFactoryContractName)
	if err != nil {
		return nil, err
	}
	ecdsaKeepFactoryContract, err := abi.NewECDSAKeepFactory(
		ecdsaKeepFactoryContractAddress,
		client,
	)
	if err != nil {
		return nil, err
	}

	return ecdsaKeepFactoryContract, nil
}

func initializeECDSAKeep(keepAddress common.Address, client *ethclient.Client) (*abi.ECDSAKeep, error) {
	ecdsaKeepContract, err := abi.NewECDSAKeep(
		keepAddress,
		client,
	)
	if err != nil {
		return nil, err
	}

	return ecdsaKeepContract, nil
}

func createTransactorOpts(privateKeyHex string) (*bind.TransactOpts, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}
	senderAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	transactorOpts := bind.NewKeyedTransactor(privateKey)
	transactorOpts.Value = big.NewInt(0) // in wei
	transactorOpts.From = senderAddress

	return transactorOpts, nil
}
