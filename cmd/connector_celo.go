//+build celo

package cmd

import (
	"context"
	"fmt"

	"github.com/keep-network/keep-common/pkg/chain/celo/celoutil"
	"github.com/keep-network/keep-ecdsa/config"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/celo"
)

func connectChain(
	ctx context.Context,
	config *config.Config,
) (chain.Handle, *operatorKeys, error) {
	celoKey, err := celoutil.DecryptKeyFile(
		config.Celo.Account.KeyFile,
		config.Celo.Account.KeyFilePassword,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to read key file [%s]: [%v]",
			config.Celo.Account.KeyFile,
			err,
		)
	}

	// DEPRECATED: config.Celo.ContractAddresses is the correct container
	// for the TBTCSystem address from now on; default to Extensions.TBTC and
	// warn if the ContractAddresses version is not set yet.
	_, exists := config.Celo.ContractAddresses[celo.TBTCSystemContractName]
	if !exists && len(config.Extensions.TBTC.TBTCSystem) != 0 {
		logger.Warn(
			"TBTCSystem address configuration in Extensions.TBTC.TBTCSystem " +
				"is DEPRECATED and will be removed. Please configure the " +
				"TBTCSystem address alongside BondedECDSAKeep under " +
				"Celo.ContractAddresses.",
		)
		config.Celo.ContractAddresses[celo.TBTCSystemContractName] =
			config.Extensions.TBTC.TBTCSystem
	}

	celoChain, err := celo.Connect(ctx, celoKey, &config.Celo)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to connect to celo node: [%v]",
			err,
		)
	}

	operatorKeys := &operatorKeys{
		public:  &celoKey.PrivateKey.PublicKey,
		private: celoKey.PrivateKey,
	}

	return celoChain, operatorKeys, nil
}

func extractKeyFilePassword(config *config.Config) string {
	return config.Celo.Account.KeyFilePassword
}
