package client

import (
	"context"
	"fmt"

	"github.com/keep-network/keep-core/pkg/net"
	"github.com/keep-network/keep-core/pkg/operator"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/bitcoin"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss"
	"github.com/keep-network/keep-ecdsa/pkg/extensions/tbtc"
	"github.com/keep-network/keep-ecdsa/pkg/extensions/tbtc/recovery"
	"github.com/keep-network/keep-ecdsa/pkg/node"
	"github.com/keep-network/keep-ecdsa/pkg/registry"
)

const (
	defaultVbyteFee               = 75
	estimatedTransactionSizeVByte = 175
)

// TODO: Should this function be moved to `node` package under tss.Node?
func handleLiquidationRecovery(
	ctx context.Context,
	hostChain chain.Handle,
	tbtcHandle chain.TBTCHandle,
	bitcoinHandle bitcoin.Handle,
	networkProvider net.Provider,
	tbtcConfig *tbtc.Config,
	tssNode *node.Node,
	operatorPublicKey *operator.PublicKey,
	keep chain.BondedECDSAKeepHandle,
	keepsRegistry *registry.Keeps,
	derivationIndexStorage *recovery.DerivationIndexStorage,
) error {
	logger.Infof(
		"starting liquidation recovery protocol for keep [%s]",
		keep.ID(),
	)

	members, err := keep.GetMembers()
	if err != nil {
		return fmt.Errorf(
			"failed to retrieve members from keep [%s]: [%w]",
			keep.ID(),
			err,
		)
	}

	memberID := tss.MemberIDFromPublicKey(operatorPublicKey)

	memberIDs, err := tssNode.AnnounceSignerPresence(
		ctx,
		operatorPublicKey,
		keep.ID(),
		members,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to announce signer presence on keep [%s] termination: [%w]",
			keep.ID(),
			err,
		)
	}

	chainParams, err := tbtcConfig.Bitcoin.ChainParams()
	if err != nil {
		return fmt.Errorf(
			"failed to parse the configured net params: [%w]",
			err,
		)
	}

	beneficiaryAddress, err := recovery.ResolveAddress(
		tbtcConfig.Bitcoin.BeneficiaryAddress,
		derivationIndexStorage,
		chainParams,
		bitcoinHandle,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to resolve a btc address for keep [%s] address: [%s]: [%w]",
			keep.ID(),
			tbtcConfig.Bitcoin.BeneficiaryAddress,
			err,
		)
	}

	depositAddress, err := keep.GetOwner()
	if err != nil {
		return fmt.Errorf(
			"failed to retrieve the owner for keep [%s]: [%w]",
			keep.ID(),
			err,
		)
	}

	fundingInfo, err := tbtcHandle.FundingInfo(depositAddress.String())
	if err != nil {
		return fmt.Errorf(
			"failed to retrieve the funding info of deposit [%s] for keep [%s]: [%w]",
			depositAddress,
			keep.ID(),
			err,
		)
	}
	previousOutputValue := int32(chain.UtxoValueBytesToUint32(fundingInfo.UtxoValueBytes))

	vbyteFee := resolveVbyteFee(bitcoinHandle, tbtcConfig, previousOutputValue)

	btcAddresses, maxFeePerVByte, err := tss.BroadcastRecoveryAddress(
		ctx,
		beneficiaryAddress,
		vbyteFee,
		keep.ID().String(),
		memberID,
		memberIDs,
		uint(len(memberIDs)-1),
		networkProvider,
		hostChain.Signing().PublicKeyToAddress,
		chainParams,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to communicate recovery details for keep [%s]: [%w]",
			keep.ID(),
			err,
		)
	}

	signer, err := keepsRegistry.GetSigner(keep.ID())
	if err != nil {
		// If there are no signer for loaded keep then something is clearly
		// wrong. We don't want to continue processing for this keep.
		return fmt.Errorf("no signer for keep [%s]: [%w]", keep.ID(), err)
	}

	logger.Infof(
		"building liquidation recovery transaction for keep [%s] "+
			"with receiving addresses [%v] and maxFeePerVByte [%d]",
		keep.ID(),
		btcAddresses,
		maxFeePerVByte,
	)

	recoveryTransactionHex, err := recovery.BuildBitcoinTransaction(
		ctx,
		networkProvider,
		hostChain,
		fundingInfo,
		signer,
		chainParams,
		btcAddresses,
		maxFeePerVByte,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to build the transaction for keep [%s]: [%w]",
			keep.ID(),
			err,
		)
	}

	logger.Debugf(
		"broadcasting liquidation recovery transaction for keep [%s]: [%s]",
		keep.ID(),
		recoveryTransactionHex,
	)

	broadcastError := bitcoinHandle.Broadcast(recoveryTransactionHex)
	if broadcastError != nil {
		logger.Errorf(
			"failed to broadcast liquidation recovery transaction for keep [%s]: [%v]",
			keep.ID(),
			broadcastError,
		)

		for i := 0; i < 5; i++ {
			logger.Warningf("Please broadcast Bitcoin transaction %s", recoveryTransactionHex)
		}
	}

	return nil
}

// resolveVbyteFee fetches vByte fee for 25 blocks from the bitcoin handle. If a
// call to Bitcoin API fails the function catches and logs the error but doesn't
// fail the execution.
//
// If a value of vByte fee was returned from the bitcoin handle it is used to
// calculate transaction fee estimate. If the estimated transaction fee exceeds
// the transaction value more than 5%, then the lesser of the suggested vByte fee
// or configured MaxFeePerVByte is used.
//
// If a value of vByte fee was not fetched from the bitcoin handle the function
// tries to read it from a config file. If the value is not defined in the config file
// it returns a default vByte fee.
func resolveVbyteFee(
	bitcoinHandle bitcoin.Handle,
	tbtcConfig *tbtc.Config,
	previousOutputValue int32,
) int32 {
	vbyteFee, vbyteFeeError := bitcoinHandle.VbyteFeeFor25Blocks()
	if vbyteFeeError != nil {
		logger.Errorf(
			"failed to retrieve a vbyte fee estimate: [%v]",
			vbyteFeeError,
		)
		// Since the electrs connection is optional, we don't return the error.
	}
	if vbyteFee > 0 {
		// Fee computation requires that the full transaction be assembled. The
		// transaction fee should be estimated by assuming the final transaction
		// will be 175 vBytes, which should be very close to accurate given the
		// regular structure of liquidation refund transactions.
		estimatedTransactionFee := vbyteFee * estimatedTransactionSizeVByte

		fivePercentPreviousOutput := previousOutputValue / 20 // 5% of UTXO

		// There is one exception to the rule that the 25-block suggested fee
		// should be used in the presence of a Bitcoin connection. If this
		// suggested fee would result in a fee consuming more than 5% of the UTXO
		// value that is being split, then the lesser of the 25-block suggested
		// fee and MaxFeePerVByte configured fee should be used.
		if estimatedTransactionFee > fivePercentPreviousOutput {
			if tbtcConfig.Bitcoin.MaxFeePerVByte > 0 {
				vbyteFee = min(vbyteFee, tbtcConfig.Bitcoin.MaxFeePerVByte)
			}

			// If the final fee is computed as being >5% of UTXO value, the client
			// should log a WARN message so that the operator can be aware of this
			// and revise their max fee per vByte if desired.
			logger.Warnf(
				"estimated transaction fee [%d] is greater than 5%% of the UTXO value [%d]; "+
					"using [%d] as fee per vByte",
				estimatedTransactionFee,
				previousOutputValue,
				vbyteFee,
			)
		}
	}
	if vbyteFee == 0 {
		vbyteFee = tbtcConfig.Bitcoin.MaxFeePerVByte
	}
	if vbyteFee == 0 {
		vbyteFee = defaultVbyteFee
	}

	return vbyteFee
}

func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
