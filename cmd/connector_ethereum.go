//+build !celo

package cmd

import (
	"context"
	"fmt"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-ecdsa/config"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/ethereum"
	"github.com/keep-network/keep-ecdsa/pkg/extensions/tbtc"
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

	// FIXME tBTC stuff needs some luuuuv.
	ethereumChain, err := ethereum.Connect(
		ctx,
		ethereumKey,
		&config.Ethereum,
		config.Extensions.TBTC.TBTCSystem,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to connect to ethereum node: [%v]",
			err,
		)
	}

	tbtcHandle, err := ethereumChain.TBTCApplicationHandle()
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to set up tBTC application: [%v]",
			err,
		)
	}

	initializeExtensions(
		ctx,
		ethereumChain,
		tbtcHandle,
	)

	operatorKeys := &operatorKeys{
		public:  &ethereumKey.PrivateKey.PublicKey,
		private: ethereumKey.PrivateKey,
	}

	return ethereumChain, operatorKeys, nil
}

func initializeExtensions(
	ctx context.Context,
	chain chain.Handle,
	tbtcEthereumChain chain.TBTCHandle,
) {
	if tbtcEthereumChain != nil {
		tbtc.Initialize(
			ctx,
			tbtcEthereumChain,
			chain.BlockCounter(),
			chain.BlockTimestamp,
		)
	} else {
		logger.Errorf(
			"could not initialize tbtc chain extension",
		)
	}
}

func extractKeyFilePassword(config *config.Config) string {
	return config.Ethereum.Account.KeyFilePassword
}
