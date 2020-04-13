package ethereum

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/keep-network/keep-common/pkg/chain/ethereum"
	"github.com/keep-network/keep-common/pkg/chain/ethereum/blockcounter"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/contract"
)

// EthereumChain is an implementation of ethereum blockchain interface.
type EthereumChain struct {
	config                         *ethereum.Config
	accountKey                     *keystore.Key
	client                         *ethclient.Client
	bondedECDSAKeepFactoryContract *contract.BondedECDSAKeepFactory
	blockCounter                   *blockcounter.EthereumBlockCounter

	// transactionMutex allows interested parties to forcibly serialize
	// transaction submission.
	//
	// When transactions are submitted, they require a valid nonce. The nonce is
	// equal to the count of transactions the account has submitted so far, and
	// for a transaction to be accepted it should be monotonically greater than
	// any previous submitted transaction. To do this, transaction submission
	// asks the Ethereum client it is connected to for the next pending nonce,
	// and uses that value for the transaction. Unfortunately, if multiple
	// transactions are submitted in short order, they may all get the same
	// nonce. Serializing submission ensures that each nonce is requested after
	// a previous transaction has been submitted.
	transactionMutex *sync.Mutex
}

// Connect performs initialization for communication with Ethereum blockchain
// based on provided config.
func Connect(accountKey *keystore.Key, config *ethereum.Config) (chain.Handle, error) {
	client, err := ethclient.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	transactionMutex := &sync.Mutex{}

	bondedECDSAKeepFactoryContractAddress, err := config.ContractAddress(BondedECDSAKeepFactoryContractName)
	if err != nil {
		return nil, err
	}
	bondedECDSAKeepFactoryContract, err := contract.NewBondedECDSAKeepFactory(
		*bondedECDSAKeepFactoryContractAddress,
		accountKey,
		client,
		transactionMutex,
	)
	if err != nil {
		return nil, err
	}

	blockCounter, err := blockcounter.CreateBlockCounter(client)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create Ethereum blockcounter: [%v]",
			err,
		)
	}

	return &EthereumChain{
		config:                         config,
		accountKey:                     accountKey,
		client:                         client,
		transactionMutex:               transactionMutex,
		bondedECDSAKeepFactoryContract: bondedECDSAKeepFactoryContract,
		blockCounter:                   blockCounter,
	}, nil
}
