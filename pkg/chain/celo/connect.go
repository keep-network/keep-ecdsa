package celo

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/celo-org/celo-blockchain/accounts/keystore"
	celoclient "github.com/celo-org/celo-blockchain/ethclient"
	"github.com/keep-network/keep-common/pkg/chain/celo/celoutil"
	"github.com/keep-network/keep-common/pkg/chain/ethereum"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/contract"
)

// Definitions of contract names.
const (
	BondedECDSAKeepFactoryContractName = "BondedECDSAKeepFactory"
)

// TODO: revisit those constants values
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

// TODO: Replace `*ethereum.Config` and `*contract.BondedECDSAKeepFactory`
// CeloChain is an implementation of Celo blockchain interface.
type CeloChain struct {
	config                         *ethereum.Config
	accountKey                     *keystore.Key
	client                         celoutil.CeloClient
	bondedECDSAKeepFactoryContract *contract.BondedECDSAKeepFactory
	blockCounter                   *ethlike.BlockCounter
	miningWaiter                   *ethlike.MiningWaiter
	nonceManager                   *ethlike.NonceManager

	// transactionMutex allows interested parties to forcibly serialize
	// transaction submission.
	//
	// When transactions are submitted, they require a valid nonce. The nonce is
	// equal to the count of transactions the account has submitted so far, and
	// for a transaction to be accepted it should be monotonically greater than
	// any previous submitted transaction. To do this, transaction submission
	// asks the Celo client it is connected to for the next pending nonce,
	// and uses that value for the transaction. Unfortunately, if multiple
	// transactions are submitted in short order, they may all get the same
	// nonce. Serializing submission ensures that each nonce is requested after
	// a previous transaction has been submitted.
	transactionMutex *sync.Mutex
}

// Connect performs initialization for communication with Celo blockchain
// based on provided config.
func Connect(
	accountKey *keystore.Key,
	config *ethereum.Config,
) (*CeloChain, error) {
	client, err := celoclient.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	wrappedClient := addClientWrappers(config, client)

	transactionMutex := &sync.Mutex{}

	nonceManager := ethlike.NewNonceManager(
		accountKey.Address.Hex(),
		celoutil.NewNonceSourceAdapter(wrappedClient),
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
	miningWaiter := ethlike.NewMiningWaiter(
		celoutil.NewTransactionSourceAdapter(wrappedClient),
		checkInterval,
		maxGasPrice,
	)

	blockCounter, err := ethlike.CreateBlockCounter(
		celoutil.NewBlockSourceAdapter(wrappedClient),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create Celo blockcounter: [%v]",
			err,
		)
	}

	bondedECDSAKeepFactoryContractAddress, err := config.ContractAddress(
		BondedECDSAKeepFactoryContractName,
	)
	if err != nil {
		return nil, err
	}

	// TODO: create Celo contract bindings
	bondedECDSAKeepFactoryContract, err := contract.NewBondedECDSAKeepFactory(
		*bondedECDSAKeepFactoryContractAddress,
		accountKey,
		wrappedClient,
		nonceManager,
		miningWaiter,
		blockCounter,
		transactionMutex,
	)
	if err != nil {
		return nil, err
	}

	return &CeloChain{
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
	client celoutil.CeloClient,
) celoutil.CeloClient {
	loggingClient := celoutil.WrapCallLogging(logger, client)

	if config.RequestsPerSecondLimit > 0 || config.ConcurrencyLimit > 0 {
		logger.Infof(
			"enabled Celo rate limiter; "+
				"rps limit [%v]; "+
				"concurrency limit [%v]",
			config.RequestsPerSecondLimit,
			config.ConcurrencyLimit,
		)

		return celoutil.WrapRateLimiting(
			loggingClient,
			&ethlike.RateLimiterConfig{
				RequestsPerSecondLimit: config.RequestsPerSecondLimit,
				ConcurrencyLimit:       config.ConcurrencyLimit,
			},
		)
	}

	return loggingClient
}
