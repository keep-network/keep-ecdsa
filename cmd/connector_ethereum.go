//+build !celo

package cmd

import (
	"context"
	"fmt"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-ecdsa/config"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/ethereum"
)

func connectChain(
	ctx context.Context,
	config *config.Config,
) (chain.Handle, *operatorKeys, error) {
	ethereumKey, err := ethutil.DecryptKeyFile(
		config.Ethereum.Account.KeyFile,
		config.Ethereum.Account.KeyFilePassword,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to read key file [%s]: [%v]",
			config.Ethereum.Account.KeyFile,
			err,
		)
	}

	// DEPRECATED: config.Ethereum.ContractAddresses is the correct container
	// for the TBTCSystem address from now on; default to Extensions.TBTC and
	// warn if the ContractAddresses version is not set yet.
	_, exists := config.Ethereum.ContractAddresses[ethereum.TBTCSystemContractName]
	if !exists && len(config.Extensions.TBTC.TBTCSystem) != 0 {
		logger.Warn(
			"TBTCSystem address configuration in Extensions.TBTC.TBTCSystem " +
				"is DEPRECATED and will be removed. Please configure the " +
				"TBTCSystem address alongside BondedECDSAKeep under " +
				"Ethereum.ContractAddresses.",
		)
		config.Ethereum.ContractAddresses[ethereum.TBTCSystemContractName] =
			config.Extensions.TBTC.TBTCSystem
	}

	ethereumChain, err := ethereum.Connect(
		ctx,
		ethereumKey,
		&config.Ethereum,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to connect to ethereum node: [%v]",
			err,
		)
	}

	operatorKeys := &operatorKeys{
		public:  &ethereumKey.PrivateKey.PublicKey,
		private: ethereumKey.PrivateKey,
	}

	return ethereumChain, operatorKeys, nil
}

func extractKeyFilePassword(config *config.Config) string {
	return config.Ethereum.Account.KeyFilePassword
}
