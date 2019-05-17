package ethereum

import (
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/keep-network/keep-tecdsa/pkg/eth/chain/gen/abi"
)

// EthereumChain is an implementation of ethereum blockchain interface.
type EthereumChain struct {
	config                   *Config
	client                   *ethclient.Client
	keepTECDSAGroupContract  *abi.KeepTECDSAGroup // TODO: Remove
	ecdsaKeepFactoryContract *abi.ECDSAKeepFactory
}

// Connect performs initialization for communication with Ethereum blockchain
// based on provided config.
func Connect(config *Config) (*EthereumChain, error) {
	client, err := ethclient.Dial(config.URL)
	if err != nil {
		log.Fatal(err)
	}

	keepTECDSAGroupContractAddress, err := config.ContractAddress(KeepTECDSAGroupContractName)
	if err != nil {
		return nil, err
	}

	keepTECDSAGroupContract, err := abi.NewKeepTECDSAGroup(
		keepTECDSAGroupContractAddress,
		client,
	)
	if err != nil {
		return nil, err
	}

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

	return &EthereumChain{
		config:                   config,
		client:                   client,
		keepTECDSAGroupContract:  keepTECDSAGroupContract,
		ecdsaKeepFactoryContract: ecdsaKeepFactoryContract,
	}, nil
}
