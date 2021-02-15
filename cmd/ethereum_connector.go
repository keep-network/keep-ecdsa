//+build !celo

package cmd

import (
	"context"
	"fmt"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-ecdsa/config"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/ethereum"
	"github.com/keep-network/keep-ecdsa/pkg/extensions/tbtc"
)

func connectChain(
	ctx context.Context,
	config *config.Config,
) (eth.Handle, *operatorKeys, error) {
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

	ethereumChain, err := ethereum.Connect(ethereumKey, &config.Ethereum)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to connect to ethereum node: [%v]",
			err,
		)
	}

	initializeExtensions(ctx, config.Extensions, ethereumChain)
	initializeBalanceMonitoring(
		ctx,
		ethereumChain,
		&config.Ethereum,
		ethereumKey.Address.Hex(),
	)

	operatorKeys := &operatorKeys{
		public:  &ethereumKey.PrivateKey.PublicKey,
		private: ethereumKey.PrivateKey,
	}

	return ethereumChain, operatorKeys, nil
}

func initializeExtensions(
	ctx context.Context,
	config config.Extensions,
	ethereumChain *ethereum.EthereumChain,
) {
	if len(config.TBTC.TBTCSystem) > 0 {
		tbtcEthereumChain, err := ethereum.WithTBTCExtension(
			ethereumChain,
			config.TBTC.TBTCSystem,
		)
		if err != nil {
			logger.Errorf(
				"could not initialize tbtc chain extension: [%v]",
				err,
			)
			return
		}

		tbtc.Initialize(ctx, tbtcEthereumChain)
	}
}

func extractChainConfig(config *config.Config) chainConfig {
	return &config.Ethereum
}
