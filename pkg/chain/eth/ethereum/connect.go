package ethereum

import (
	cecdsa "crypto/ecdsa"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	coreAbi "github.com/keep-network/keep-core/pkg/chain/gen/abi"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/gen/abi"
)

// EthereumChain is an implementation of ethereum blockchain interface.
type EthereumChain struct {
	config                           *Config
	client                           *ethclient.Client
	transactorOptions                *bind.TransactOpts
	ecdsaKeepFactoryContract         *abi.ECDSAKeepFactory
	keepRandomBeaconOperatorContract *coreAbi.KeepRandomBeaconOperator
	publicKey                        *cecdsa.PublicKey
}

// Connect performs initialization for communication with Ethereum blockchain
// based on provided config.
func Connect(privateKey *cecdsa.PrivateKey, config *Config) (eth.Handle, error) {
	client, err := ethclient.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	transactorOptions := bind.NewKeyedTransactor(privateKey)

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

	keepRandomBeaconOperatorContractAddress, err := config.ContractAddress(
		KeepRandomBeaconOperatorContractName,
	)
	if err != nil {
		return nil, err
	}
	// TODO: can we use a generated dependency or should we generate it on our own?
	keepRandomBeaconOperatorContract, err := coreAbi.NewKeepRandomBeaconOperator(
		keepRandomBeaconOperatorContractAddress,
		client,
	)
	if err != nil {
		return nil, err
	}

	return &EthereumChain{
		config:                           config,
		client:                           client,
		transactorOptions:                transactorOptions,
		ecdsaKeepFactoryContract:         ecdsaKeepFactoryContract,
		keepRandomBeaconOperatorContract: keepRandomBeaconOperatorContract,
		publicKey:                        &privateKey.PublicKey,
	}, nil
}
