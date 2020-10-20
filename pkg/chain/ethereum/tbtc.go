package ethereum

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/tbtc/pkg/chain/ethereum/gen/contract"
)

// TBTCEthereumChain represents an Ethereum chain handle with
// TBTC-specific capabilities.
type TBTCEthereumChain struct {
	*EthereumChain

	tbtcSystemContract *contract.TBTCSystem
}

// WithTBTCExtension extends the Ethereum chain handle with
// TBTC-specific capabilities.
func WithTBTCExtension(
	ethereumChain *EthereumChain,
	tbtcSystemContractAddress string,
) (*TBTCEthereumChain, error) {
	if !common.IsHexAddress(tbtcSystemContractAddress) {
		return nil, fmt.Errorf("incorrect TBTCSystem contract address")
	}

	tbtcSystemContract, err := contract.NewTBTCSystem(
		common.HexToAddress(tbtcSystemContractAddress),
		ethereumChain.accountKey,
		ethereumChain.client,
		ethereumChain.nonceManager,
		ethereumChain.miningWaiter,
		ethereumChain.transactionMutex,
	)
	if err != nil {
		return nil, err
	}

	return &TBTCEthereumChain{
		EthereumChain:      ethereumChain,
		tbtcSystemContract: tbtcSystemContract,
	}, nil
}

// OnDepositCreated installs a callback that is invoked when an
// on-chain notification of a new deposit creation is seen.
func (tec *TBTCEthereumChain) OnDepositCreated(
	handler func(depositAddress string),
) (subscription.EventSubscription, error) {
	return tec.tbtcSystemContract.WatchCreated(
		func(
			DepositContractAddress common.Address,
			KeepAddress common.Address,
			Timestamp *big.Int,
			blockNumber uint64,
		) {
			handler(DepositContractAddress.Hex())
		},
		func(err error) error {
			return fmt.Errorf("watch deposit created failed: [%v]", err)
		},
		nil,
		nil,
	)
}

// OnDepositRegisteredPubkey installs a callback that is invoked when an
// on-chain notification of a deposit's pubkey registration is seen.
func (tec *TBTCEthereumChain) OnDepositRegisteredPubkey(
	handler func(depositAddress string),
) (subscription.EventSubscription, error) {
	return tec.tbtcSystemContract.WatchRegisteredPubkey(
		func(
			DepositContractAddress common.Address,
			SigningGroupPubkeyX [32]uint8,
			SigningGroupPubkeyY [32]uint8,
			Timestamp *big.Int,
			blockNumber uint64,
		) {
			handler(DepositContractAddress.Hex())
		},
		func(err error) error {
			return fmt.Errorf("watch deposit created failed: [%v]", err)
		},
		nil,
	)
}

// KeepAddress returns the underlying keep address for the
// provided deposit.
func (tec *TBTCEthereumChain) KeepAddress(
	depositAddress string,
) (string, error) {
	deposit, err := tec.getDepositContract(depositAddress)
	if err != nil {
		return "", err
	}

	keepAddress, err := deposit.KeepAddress()
	if err != nil {
		return "", err
	}

	return keepAddress.Hex(), nil
}

// RetrieveSignerPubkey retrieves the signer public key for the
// provided deposit.
func (tec *TBTCEthereumChain) RetrieveSignerPubkey(
	depositAddress string,
) error {
	deposit, err := tec.getDepositContract(depositAddress)
	if err != nil {
		return err
	}

	transaction, err := deposit.RetrieveSignerPubkey()
	if err != nil {
		return err
	}

	logger.Debugf(
		"submitted RetrieveSignerPubkey transaction with hash: [%x]",
		transaction.Hash(),
	)

	return nil
}

func (tec *TBTCEthereumChain) getDepositContract(
	address string,
) (*contract.Deposit, error) {
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("incorrect deposit contract address")
	}

	depositContract, err := contract.NewDeposit(
		common.HexToAddress(address),
		tec.accountKey,
		tec.client,
		tec.nonceManager,
		tec.miningWaiter,
		tec.transactionMutex,
	)
	if err != nil {
		return nil, err
	}

	return depositContract, nil
}
