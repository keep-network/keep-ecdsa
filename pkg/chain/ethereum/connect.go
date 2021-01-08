package ethereum

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/keep-network/keep-common/pkg/chain/ethereum"
	"github.com/keep-network/keep-common/pkg/chain/ethereum/blockcounter"
	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/contract"
)

var (
	// DefaultMiningCheckInterval is the default interval in which transaction
	// mining status is checked. If the transaction is not mined within this
	// time, the gas price is increased and transaction is resubmitted.
	// This value can be overwritten in the configuration file.
	DefaultMiningCheckInterval = 60 * time.Second

	// DefaultMaxGasPrice specifies the default maximum gas price the client is
	// willing to pay for the transaction to be mined. The offered transaction
	// gas price can not be higher than the max gas price value. If the maximum
	// allowed gas price is reached, no further resubmission attempts are
	// performed. This value can be overwritten in the configuration file.
	DefaultMaxGasPrice = big.NewInt(500000000000) // 500 Gwei
)

// EthereumChain is an implementation of ethereum blockchain interface.
type EthereumChain struct {
	config                         *ethereum.Config
	accountKey                     *keystore.Key
	client                         ethutil.EthereumClient
	bondedECDSAKeepFactoryContract *contract.BondedECDSAKeepFactory
	blockCounter                   *blockcounter.EthereumBlockCounter
	miningWaiter                   *ethutil.MiningWaiter
	nonceManager                   *ethutil.NonceManager

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
func Connect(accountKey *keystore.Key, config *ethereum.Config) (*EthereumChain, error) {
	client, err := ethclient.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	wrappedClient := addClientWrappers(config, client)

	transactionMutex := &sync.Mutex{}

	nonceManager := ethutil.NewNonceManager(
		accountKey.Address,
		wrappedClient,
	)

	checkInterval := DefaultMiningCheckInterval
	maxGasPrice := DefaultMaxGasPrice
	if config.MiningCheckInterval != 0 {
		checkInterval = time.Duration(config.MiningCheckInterval) * time.Second
	}
	if config.MaxGasPrice != nil {
		maxGasPrice = config.MaxGasPrice.Int
	}

	logger.Infof("using [%v] mining check interval", checkInterval)
	logger.Infof("using [%v] wei max gas price", maxGasPrice)
	miningWaiter := ethutil.NewMiningWaiter(wrappedClient, checkInterval, maxGasPrice)

	bondedECDSAKeepFactoryContractAddress, err := config.ContractAddress(BondedECDSAKeepFactoryContractName)
	if err != nil {
		return nil, err
	}
	bondedECDSAKeepFactoryContract, err := contract.NewBondedECDSAKeepFactory(
		*bondedECDSAKeepFactoryContractAddress,
		accountKey,
		wrappedClient,
		nonceManager,
		miningWaiter,
		transactionMutex,
	)
	if err != nil {
		return nil, err
	}

	blockCounter, err := blockcounter.CreateBlockCounter(wrappedClient)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create Ethereum blockcounter: [%v]",
			err,
		)
	}

	return &EthereumChain{
		config:                         config,
		accountKey:                     accountKey,
		client:                         wrappedClient,
		bondedECDSAKeepFactoryContract: bondedECDSAKeepFactoryContract,
		blockCounter:                   blockCounter,
		nonceManager:                   nonceManager,
		miningWaiter:                   miningWaiter,
		transactionMutex:               transactionMutex,
	}, nil
}

func addClientWrappers(
	config *ethereum.Config,
	client ethutil.EthereumClient,
) ethutil.EthereumClient {
	loggingClient := ethutil.WrapCallLogging(logger, client)

	if config.RequestsPerSecondLimit > 0 || config.ConcurrencyLimit > 0 {
		logger.Infof(
			"enabled ethereum rate limiter; "+
				"rps limit [%v]; "+
				"concurrency limit [%v]",
			config.RequestsPerSecondLimit,
			config.ConcurrencyLimit,
		)

		return ethutil.WrapRateLimiting(
			loggingClient,
			&ethutil.RateLimiterConfig{
				RequestsPerSecondLimit: config.RequestsPerSecondLimit,
				ConcurrencyLimit:       config.ConcurrencyLimit,
			},
		)
	}

	return loggingClient
}
