package ethereum

import (
	cecdsa "crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/gen/abi"
)

// EthereumChain is an implementation of ethereum blockchain interface.
type EthereumChain struct {
	*ethereumBlockCounter

	config                   *Config
	client                   *ethclient.Client
	transactorOptions        *bind.TransactOpts
	callerOptions            *bind.CallOpts
	ecdsaKeepFactoryContract *abi.ECDSAKeepFactory
}

// Connect performs initialization for communication with Ethereum blockchain
// based on provided config.
func Connect(privateKey *cecdsa.PrivateKey, config *Config) (eth.Handle, error) {
	client, err := ethclient.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	transactorOptions := bind.NewKeyedTransactor(privateKey)
	callerOptions := &bind.CallOpts{From: transactorOptions.From}

	ecdsaKeepFactoryContractAddress, err := config.ContractAddress(ECDSAKeepFactoryContractName)
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

	blockCounter, err := createBlockCounter(client)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create Ethereum blockcounter: [%v]",
			err,
		)
	}

	return &EthereumChain{
		config:                   config,
		client:                   client,
		transactorOptions:        transactorOptions,
		callerOptions:            callerOptions,
		ecdsaKeepFactoryContract: ecdsaKeepFactoryContract,
		ethereumBlockCounter:     blockCounter,
	}, nil
}
