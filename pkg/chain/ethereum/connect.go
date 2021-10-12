//+build !celo

package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/keep-network/keep-common/pkg/rate"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/chain/ethlike"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/keep-network/keep-common/pkg/chain/ethereum"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/ethereum/contract"
)

// Definitions of contract names.
const (
	BondedECDSAKeepFactoryContractName = "BondedECDSAKeepFactory"
	TBTCSystemContractName             = "TBTCSystem"
)

// ethereumChain is an implementation of ethereum blockchain interface.
type ethereumChain struct {
	config                         *ethereum.Config
	accountKey                     *keystore.Key
	client                         ethutil.EthereumClient
	chainID                        *big.Int
	bondedECDSAKeepFactoryContract *contract.BondedECDSAKeepFactory
	tbtcSystemAddress              common.Address
	blockCounter                   *ethlike.BlockCounter
	miningWaiter                   *ethutil.MiningWaiter
	nonceManager                   *ethlike.NonceManager

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
func Connect(
	ctx context.Context,
	accountKey *keystore.Key,
	config *ethereum.Config,
) (chain.Handle, error) {
	client, err := ethclient.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	wrappedClient := addClientWrappers(config, client)

	transactionMutex := &sync.Mutex{}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf(
			"failed to resolve Ethereum chain id: [%v]",
			err,
		)
	}

	nonceManager := ethutil.NewNonceManager(wrappedClient, accountKey.Address)

	miningWaiter := ethutil.NewMiningWaiter(wrappedClient, *config)

	blockCounter, err := ethutil.NewBlockCounter(wrappedClient)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create Ethereum blockcounter: [%v]",
			err,
		)
	}

	tbtcSystemAddress, err := config.ContractAddress(
		TBTCSystemContractName,
	)
	if err != nil {
		// If the contract address can't be looked up, let this fail later on,
		// but make sure an empty address is in place. A missing TBTCSystem
		// address should only mean that we fail to start the tBTC app handling,
		// not that the whole client fails to start.
		tbtcSystemAddress = common.Address{}
	}

	bondedECDSAKeepFactoryContractAddress, err := config.ContractAddress(
		BondedECDSAKeepFactoryContractName,
	)
	if err != nil {
		return nil, err
	}
	bondedECDSAKeepFactoryContract, err := contract.NewBondedECDSAKeepFactory(
		bondedECDSAKeepFactoryContractAddress,
		chainID,
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

	ethereum := &ethereumChain{
		config:                         config,
		accountKey:                     accountKey,
		client:                         wrappedClient,
		chainID:                        chainID,
		bondedECDSAKeepFactoryContract: bondedECDSAKeepFactoryContract,
		tbtcSystemAddress:              tbtcSystemAddress,
		blockCounter:                   blockCounter,
		nonceManager:                   nonceManager,
		miningWaiter:                   miningWaiter,
		transactionMutex:               transactionMutex,
	}

	ethereum.initializeBalanceMonitoring(ctx)

	return ethereum, nil
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
			&rate.LimiterConfig{
				RequestsPerSecondLimit: config.RequestsPerSecondLimit,
				ConcurrencyLimit:       config.ConcurrencyLimit,
			},
		)
	}

	return loggingClient
}
