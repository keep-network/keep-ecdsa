package smoketest

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/ethereum"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/gen/abi"
)

var logger = log.Logger("keep-smoketest")

var keepOwnerAddress = common.HexToAddress("0x316F8eaf0b6065a53f0eaB3DD19aC6a07af95b3D")

// Execute runs an ECDSA event smoke test. It tests if an event is emitted after
// new ECDSA keep is requested.
func Execute(config *ethereum.Config) error {
	chainAPI, err := ethereum.Connect(config)
	if err != nil {
		return err
	}

	// Setup connection to ECDSA Keep Factory contract.
	ecdsaKeepFactory, err := initializeECDSAKeepFactory(config)
	if err != nil {
		return err
	}

	transactorOpts, err := createTransactorOpts(config)
	if err != nil {
		return err
	}

	// Define callback on event.
	eventChan := make(chan *eth.ECDSAKeepCreatedEvent)

	handle := func(event *eth.ECDSAKeepCreatedEvent) {
		eventChan <- event
	}

	// Register for events.
	subscription, err := chainAPI.OnECDSAKeepCreated(handle)
	defer subscription.Unsubscribe()
	if err != nil {
		return err
	}

	groupSize := big.NewInt(10)
	honestThreshold := big.NewInt(5)

	// Request a new keep creation.
	transaction, err := ecdsaKeepFactory.OpenKeep(
		transactorOpts,
		groupSize,
		honestThreshold,
		keepOwnerAddress,
	)
	if err != nil {
		return fmt.Errorf("call to contract failed: [%s]", err)
	}
	logger.Infof(
		"new keep requested, transaction hash: [%s]",
		transaction.Hash().Hex(),
	)

	// Wait for event emission.
	actualEvent := <-eventChan

	// Log received event.
	logger.Infof("received event: [%#v]", actualEvent)

	// Validate received event.
	if !common.IsHexAddress(actualEvent.KeepAddress.String()) {
		return fmt.Errorf("invalid hex address: [%v]", actualEvent.KeepAddress)
	}

	logger.Infof("ECDSA keep built with address: [%s]", actualEvent.KeepAddress.String())

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
