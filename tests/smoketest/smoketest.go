package smoketest

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/ethereum"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/gen/abi"
)

var keepOwnerAddress = common.HexToAddress("0x316F8eaf0b6065a53f0eaB3DD19aC6a07af95b3D")

// Execute runs an ECDSA event smoke test. It tests if an event is emitted after
// new ECDSA keep is requested.
func Execute(config *ethereum.Config) error {
	chainAPI, err := ethereum.Connect(config)
	if err != nil {
		return err
	}

	// Setup connection to ECDSA Keep Factory contract.
	tecdsaKeepFactory, err := initializeTECDSAKeepFactory(config)
	if err != nil {
		return err
	}

	transactorOpts, err := createTransactorOpts(config)
	if err != nil {
		return err
	}

	// Define callback on event.
	eventChan := make(chan *eth.TECDSAKeepCreatedEvent)

	handle := func(event *eth.TECDSAKeepCreatedEvent) {
		eventChan <- event
	}

	// Register for events.
	subscription, err := chainAPI.OnTECDSAKeepCreated(handle)
	defer subscription.Unsubscribe()
	if err != nil {
		return err
	}

	groupSize := big.NewInt(10)
	honestThreshold := big.NewInt(5)

	// Request a new keep creation.
	transaction, err := tecdsaKeepFactory.OpenKeep(
		transactorOpts,
		groupSize,
		honestThreshold,
		keepOwnerAddress,
	)
	if err != nil {
		return fmt.Errorf("call to contract failed: [%s]", err)
	}
	fmt.Printf(
		"New keep requested, transaction hash: [%s].\n",
		transaction.Hash().Hex(),
	)

	// Wait for event emission.
	actualEvent := <-eventChan

	// Log received event.
	fmt.Printf("Received event: [%#v]\n", actualEvent)

	// Validate received event.
	if !common.IsHexAddress(actualEvent.KeepAddress.String()) {
		return fmt.Errorf("invalid hex address: [%v]", actualEvent.KeepAddress)
	}

	fmt.Printf("ECDSA keep built with address: [%s]\n", actualEvent.KeepAddress.String())

	return nil
}

func initializeTECDSAKeepFactory(config *ethereum.Config) (*abi.TECDSAKeepFactory, error) {
	client, err := ethclient.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	tecdsaKeepFactoryContractAddress, err := config.ContractAddress(ethereum.TECDSAKeepFactoryContractName)
	if err != nil {
		return nil, err
	}
	tecdsaKeepFactoryContract, err := abi.NewTECDSAKeepFactory(
		tecdsaKeepFactoryContractAddress,
		client,
	)
	if err != nil {
		return nil, err
	}

	return tecdsaKeepFactoryContract, nil
}

func createTransactorOpts(config *ethereum.Config) (*bind.TransactOpts, error) {
	privateKey, err := crypto.HexToECDSA(config.PrivateKey)
	if err != nil {
		return nil, err
	}
	senderAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	transactorOpts := bind.NewKeyedTransactor(privateKey)
	transactorOpts.Value = big.NewInt(0) // in wei
	transactorOpts.From = senderAddress

	return transactorOpts, nil
}
