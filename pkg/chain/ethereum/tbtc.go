package ethereum

import (
	"fmt"
	"math/big"
	"sort"

	chain "github.com/keep-network/keep-ecdsa/pkg/chain"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/tbtc/pkg/chain/ethereum/gen/contract"
	"github.com/keep-network/tbtc/pkg/chain/ethereum/gen/eventlog"
)

// TBTCEthereumChain represents an Ethereum chain handle with
// TBTC-specific capabilities.
type TBTCEthereumChain struct {
	*EthereumChain

	tbtcSystemContract *contract.TBTCSystem
	tbtcSystemEventLog *eventlog.TBTCSystemEventLog
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

	tbtcSystemEventLog, err := eventlog.NewTBTCSystemEventLog(
		common.HexToAddress(tbtcSystemContractAddress),
		ethereumChain.client,
	)
	if err != nil {
		return nil, err
	}

	return &TBTCEthereumChain{
		EthereumChain:      ethereumChain,
		tbtcSystemContract: tbtcSystemContract,
		tbtcSystemEventLog: tbtcSystemEventLog,
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
			return fmt.Errorf(
				"watch deposit registered pubkey failed: [%v]",
				err,
			)
		},
		nil,
	)
}

// OnDepositRedemptionRequested installs a callback that is invoked when an
// on-chain notification of a deposit redemption request is seen.
func (tec *TBTCEthereumChain) OnDepositRedemptionRequested(
	handler func(depositAddress string),
) (subscription.EventSubscription, error) {
	return tec.tbtcSystemContract.WatchRedemptionRequested(
		func(
			DepositContractAddress common.Address,
			Requester common.Address,
			Digest [32]uint8,
			UtxoValue *big.Int,
			RedeemerOutputScript []uint8,
			RequestedFee *big.Int,
			Outpoint []uint8,
			blockNumber uint64,
		) {
			handler(DepositContractAddress.Hex())
		},
		func(err error) error {
			return fmt.Errorf(
				"watch deposit redemption requested failed: [%v]",
				err,
			)
		},
		nil,
		nil,
		nil,
	)
}

// OnDepositGotRedemptionSignature installs a callback that is invoked when an
// on-chain notification of a deposit receiving a redemption signature is seen.
func (tec *TBTCEthereumChain) OnDepositGotRedemptionSignature(
	handler func(depositAddress string),
) (subscription.EventSubscription, error) {
	return tec.tbtcSystemContract.WatchGotRedemptionSignature(
		func(
			DepositContractAddress common.Address,
			Digest [32]uint8,
			R [32]uint8,
			S [32]uint8,
			Timestamp *big.Int,
			blockNumber uint64,
		) {
			handler(DepositContractAddress.Hex())
		},
		func(err error) error {
			return fmt.Errorf(
				"watch deposit got redemption signature failed: [%v]",
				err,
			)
		},
		nil,
		nil,
	)
}

// OnDepositRedeemed installs a callback that is invoked when an
// on-chain notification of a deposit redemption is seen.
func (tec *TBTCEthereumChain) OnDepositRedeemed(
	handler func(depositAddress string),
) (subscription.EventSubscription, error) {
	return tec.tbtcSystemContract.WatchRedeemed(
		func(
			DepositContractAddress common.Address,
			Txid [32]uint8,
			Timestamp *big.Int,
			blockNumber uint64,
		) {
			handler(DepositContractAddress.Hex())
		},
		func(err error) error {
			return fmt.Errorf(
				"watch deposit redeemed failed: [%v]",
				err,
			)
		},
		nil,
		nil,
	)
}

// PastDepositRedemptionRequestedEvents returns all redemption requested
// events for the given deposit which occurred after the provided start block.
// Returned events are sorted by the block number in the ascending order.
func (tec *TBTCEthereumChain) PastDepositRedemptionRequestedEvents(
	depositAddress string,
	startBlock uint64,
) ([]*chain.DepositRedemptionRequestedEvent, error) {
	if !common.IsHexAddress(depositAddress) {
		return nil, fmt.Errorf("incorrect deposit contract address")
	}

	events, err := tec.tbtcSystemEventLog.PastRedemptionRequestedEvents(
		[]common.Address{
			common.HexToAddress(depositAddress),
		},
		startBlock,
		nil,
	)
	if err != nil {
		return nil, err
	}

	result := make([]*chain.DepositRedemptionRequestedEvent, 0)

	for _, event := range events {
		result = append(result, &chain.DepositRedemptionRequestedEvent{
			DepositAddress:       event.DepositContractAddress.Hex(),
			RequesterAddress:     event.Requester.Hex(),
			Digest:               event.Digest,
			UtxoValue:            event.UtxoValue,
			RedeemerOutputScript: event.RedeemerOutputScript,
			RequestedFee:         event.RequestedFee,
			Outpoint:             event.Outpoint,
			BlockNumber:          event.BlockNumber,
		})
	}

	// Make sure events are sorted by block number in ascending order.
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].BlockNumber < result[j].BlockNumber
	})

	return result, nil
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

// ProvideRedemptionSignature provides the redemption signature for the
// provided deposit.
func (tec *TBTCEthereumChain) ProvideRedemptionSignature(
	depositAddress string,
	v uint8,
	r [32]uint8,
	s [32]uint8,
) error {
	deposit, err := tec.getDepositContract(depositAddress)
	if err != nil {
		return err
	}

	transaction, err := deposit.ProvideRedemptionSignature(v, r, s)
	if err != nil {
		return err
	}

	logger.Debugf(
		"submitted ProvideRedemptionSignature transaction with hash: [%x]",
		transaction.Hash(),
	)

	return nil
}

// IncreaseRedemptionFee increases the redemption fee for the provided deposit.
func (tec *TBTCEthereumChain) IncreaseRedemptionFee(
	depositAddress string,
	previousOutputValueBytes [8]uint8,
	newOutputValueBytes [8]uint8,
) error {
	deposit, err := tec.getDepositContract(depositAddress)
	if err != nil {
		return err
	}

	transaction, err := deposit.IncreaseRedemptionFee(
		previousOutputValueBytes,
		newOutputValueBytes,
	)
	if err != nil {
		return err
	}

	logger.Debugf(
		"submitted IncreaseRedemptionFee transaction with hash: [%x]",
		transaction.Hash(),
	)

	return nil
}

// ProvideRedemptionProof provides the redemption proof for the provided deposit.
func (tec *TBTCEthereumChain) ProvideRedemptionProof(
	depositAddress string,
	txVersion [4]uint8,
	txInputVector []uint8,
	txOutputVector []uint8,
	txLocktime [4]uint8,
	merkleProof []uint8,
	txIndexInBlock *big.Int,
	bitcoinHeaders []uint8,
) error {
	deposit, err := tec.getDepositContract(depositAddress)
	if err != nil {
		return err
	}

	transaction, err := deposit.ProvideRedemptionProof(
		txVersion,
		txInputVector,
		txOutputVector,
		txLocktime,
		merkleProof,
		txIndexInBlock,
		bitcoinHeaders,
	)
	if err != nil {
		return err
	}

	logger.Debugf(
		"submitted ProvideRedemptionProof transaction with hash: [%x]",
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
