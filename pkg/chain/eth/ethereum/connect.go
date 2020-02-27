package ethereum

import (
	cecdsa "crypto/ecdsa"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/keep-network/keep-common/pkg/chain/ethereum"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/gen/abi"
)

// EthereumChain is an implementation of ethereum blockchain interface.
type EthereumChain struct {
	config                         *ethereum.Config
	accountKey                     *keystore.Key
	client                         *ethclient.Client
	transactorOptions              *bind.TransactOpts
	callerOptions                  *bind.CallOpts
	bondedECDSAKeepFactoryContract *abi.BondedECDSAKeepFactory
}

// Connect performs initialization for communication with Ethereum blockchain
// based on provided config.
func Connect(accountKey *keystore.Key, config *ethereum.Config) (eth.Handle, error) {
	client, err := ethclient.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	transactorOptions := bind.NewKeyedTransactor(privateKey)
	callerOptions := &bind.CallOpts{From: transactorOptions.From}

	bondedECDSAKeepFactoryContractAddress, err := config.ContractAddress(BondedECDSAKeepFactoryContractName)
	if err != nil {
		return nil, err
	}
	bondedECDSAKeepFactoryContract, err := abi.NewBondedECDSAKeepFactory(
		bondedECDSAKeepFactoryContractAddress,
		client,
	)
	if err != nil {
		return nil, err
	}

	return &EthereumChain{
		config:                         config,
		accountKey:                     accountKey,
		client:                         client,
		transactorOptions:              transactorOptions,
		callerOptions:                  callerOptions,
		bondedECDSAKeepFactoryContract: bondedECDSAKeepFactoryContract,
	}, nil
}
