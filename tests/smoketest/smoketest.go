package smoketest

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/ethereum"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/gen/abi"
)

// Private key of a transactions to contract sender.
const senderPrivateKeyString = "bd03a0aa0b96c5cff1accafdc806aa7f655b6a9a13aeb79f4669c9cfad1eb265"

// Execute runs an ECDSA event smoke test. It tests if an event is emitted after
// new ECDSA keep is requested.
func Execute(config *ethereum.Config) error {
	chainAPI, err := ethereum.Connect(config)
	if err != nil {
		return err
	}

	// Define callback on event.
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	var actualEvent *eth.ECDSAKeepCreatedEvent
	handle := func(event *eth.ECDSAKeepCreatedEvent) {
		actualEvent = event
		waitGroup.Done()
	}

	// Register for events.
	subscription, err := chainAPI.OnECDSAKeepCreated(handle)
	defer subscription.Unsubscribe()
	if err != nil {
		return err
	}

	// Setup connection to ECDSA Keep Factory contract.
	ecdsaKeepFactory, err := initializeECDSAKeepFactory(config)
	if err != nil {
		return err
	}

	transactorOpts, err := createTransactorOpts()
	if err != nil {
		return err
	}

	groupSize := big.NewInt(10)
	honestThreshold := big.NewInt(5)
	keepOwnerAddress := common.HexToAddress("0x316F8eaf0b6065a53f0eaB3DD19aC6a07af95b3D")

	// Request a new keep creation.
	transaction, err := ecdsaKeepFactory.CreateNewKeep(
		transactorOpts,
		groupSize,
		honestThreshold,
		keepOwnerAddress,
	)
	if err != nil {
		return fmt.Errorf("call to contract failed: [%s]", err)
	}
	fmt.Printf(
		"New keep requested, transaction hash: %s.\n",
		transaction.Hash().Hex(),
	)

	// Wait for event emission.
	waitGroup.Wait()

	// Log received event.
	marshaled, err := json.MarshalIndent(actualEvent, "", " ")
	if err != nil {
		return fmt.Errorf("cannot marshal received event: [%s]", err)
	}
	fmt.Printf("Received event:\n%s\n", marshaled)

	// Validate received event.
	if !common.IsHexAddress(actualEvent.KeepAddress.String()) {
		return fmt.Errorf("invalid hex address: %v", actualEvent.KeepAddress)
	}

	fmt.Printf("ECDSA keep built with address: %s\n", actualEvent.KeepAddress.String())

	return nil
}

func initializeECDSAKeepFactory(config *ethereum.Config) (*abi.ECDSAKeepFactory, error) {
	client, err := ethclient.Dial(config.URL)
	if err != nil {
		return nil, err
	}

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

func createTransactorOpts() (*bind.TransactOpts, error) {
	privateKey, err := crypto.HexToECDSA(senderPrivateKeyString)
	if err != nil {
		return nil, err
	}
	senderAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	transactorOpts := bind.NewKeyedTransactor(privateKey)
	transactorOpts.Value = big.NewInt(0) // in wei
	transactorOpts.From = senderAddress

	return transactorOpts, nil
}
