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

func offlineChain(
	config *config.Config,
) (chain.OfflineHandle, error) {
	ethereumKey, err := ethutil.DecryptKeyFile(
		config.Ethereum.Account.KeyFile,
		config.Ethereum.Account.KeyFilePassword,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read key file [%s]: [%v]",
			config.Ethereum.Account.KeyFile,
			err,
		)
	}

	return ethereum.Offline(ethereumKey, &config.Ethereum), nil
}

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
	if len(config.Extensions.TBTC.TBTCSystem) != 0 {
		logger.Warn(
			"TBTCSystem address configuration in Extensions.TBTC.TBTCSystem " +
				"is DEPRECATED and will be removed. Please configure the " +
				"TBTCSystem address alongside BondedECDSAKeep under " +
				"Ethereum.ContractAddresses.",
		)

		if !exists {
			config.Ethereum.ContractAddresses[ethereum.TBTCSystemContractName] =
				config.Extensions.TBTC.TBTCSystem

			// Flag that the contract address entry now exists to skip the next
			// default.
			exists = true
		} else {
			if config.Ethereum.ContractAddresses[ethereum.TBTCSystemContractName] !=
				config.Extensions.TBTC.TBTCSystem {
				panic(
					"Configured TBTCSystem contract and Extensions.TBTC.TBTCSystem " +
						"do not match. Failing to boot to avoid misconfiguration. " +
						"Please ensure ethereum.ContractAddresses." +
						ethereum.TBTCSystemContractName + "is set to the correct " +
						"tBTC system contract and remove Extensions.TBTC.TBTCSystem " +
						"entry, then try starting again.",
				)
			}
		}
	}

	// DEPRECATED: config.Ethereum.ContractAddresses is the correct container
	// for the TBTCSystem address from now on; read SanctionedApplications and
	// assume it has a single entry that is TBTCSystem, warn if
	// SanctionedApplications needs to be used.
	applicationAddresses := config.SanctionedApplications.AddressesStrings
	if len(applicationAddresses) != 0 {
		logger.Warn(
			"TBTCSystem address configuration in SanctionedApplications.Addresses " +
				"is DEPRECATED and will be removed. Please configure the " +
				"TBTCSystem address alongside BondedECDSAKeep under " +
				"Ethereum.ContractAddresses.",
		)

		if !exists {
			config.Ethereum.ContractAddresses[ethereum.TBTCSystemContractName] =
				applicationAddresses[0]
		} else {
			if config.Ethereum.ContractAddresses[ethereum.TBTCSystemContractName] !=
				applicationAddresses[0] {
				panic(
					"Configured TBTCSystem contract and SanctionedApplications list " +
						"do not match. Failing to boot to avoid misconfiguration. " +
						"Please ensure ethereum.ContractAddresses." +
						ethereum.TBTCSystemContractName + "is set to the correct " +
						"tBTC system contract and remove SanctionedApplications " +
						"configuration list, then try starting again.",
				)
			}
		}
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
