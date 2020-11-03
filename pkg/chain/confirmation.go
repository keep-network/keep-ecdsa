package eth

import (
	"fmt"
)

// WaitForChainConfirmation ensures that after receiving specific number of block
// confirmations the state of the chain is actually as expected. It waits for
// predefined number of blocks since the start block number provided. After the
// required block number is reached it performs a check of the chain state with
// a provided function returning a boolean value.
func WaitForChainConfirmation(
	ethereumChain Handle,
	startBlockNumber uint64,
	blockConfirmations uint64,
	stateCheck func() (bool, error),
) (bool, error) {
	blockHeight := startBlockNumber + blockConfirmations
	logger.Infof("waiting for block [%d] to confirm chain state", blockHeight)

	err := ethereumChain.BlockCounter().WaitForBlockHeight(blockHeight)
	if err != nil {
		return false, fmt.Errorf("failed to wait for block height: [%v]", err)
	}

	result, err := stateCheck()
	if err != nil {
		return false, fmt.Errorf("failed to get chain state confirmation: [%v]", err)
	}

	return result, nil
}
