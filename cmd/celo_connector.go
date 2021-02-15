//+build celo

package cmd

import (
	"context"
	"fmt"

	"github.com/keep-network/keep-common/pkg/chain/celo/celoutil"
	"github.com/keep-network/keep-ecdsa/config"
	eth "github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/celo"
)

func connectChain(
	ctx context.Context,
	config *config.Config,
) (eth.Handle, *operatorKeys, error) {
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

	celoChain, err := celo.Connect(celoKey, &config.Celo)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to connect to celo node: [%v]",
			err,
		)
	}

	// TODO: initialize Celo extensions
	initializeBalanceMonitoring(
		ctx,
		celoChain,
		&config.Celo,
		celoKey.Address.Hex(),
	)

	operatorKeys := &operatorKeys{
		public:  &ethereumKey.PrivateKey.PublicKey,
		private: ethereumKey.PrivateKey,
	}

	return celoChain, operatorKeys, nil
}

func extractChainConfig(config *config.Config) chainConfig {
	return &config.Celo
}
