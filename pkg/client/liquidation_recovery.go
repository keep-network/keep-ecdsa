package client

import (
	"context"

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
	defaultVbyteFee = 75
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
		logger.Errorf(
			"failed to retrieve members from keep [%s]: [%v]",
			keep.ID(),
			err,
		)
		return err
	}

	memberID := tss.MemberIDFromPublicKey(operatorPublicKey)

	memberIDs, err := tssNode.AnnounceSignerPresence(
		ctx,
		operatorPublicKey,
		keep.ID(),
		members,
	)
	if err != nil {
		logger.Errorf(
			"failed to announce signer presence on keep [%s] termination: [%v]",
			keep.ID(),
			err,
		)
		return err
	}

	chainParams, err := tbtcConfig.Bitcoin.ChainParams()
	if err != nil {
		logger.Errorf(
			"failed to parse the configured net params: [%v]",
			err,
		)
		return err
	}

	beneficiaryAddress, err := recovery.ResolveAddress(
		tbtcConfig.Bitcoin.BeneficiaryAddress,
		derivationIndexStorage,
		chainParams,
		bitcoinHandle,
		false,
	)
	if err != nil {
		logger.Errorf(
			"failed to resolve a btc address for keep: [%s] address: [%s] err: [%v]",
			keep.ID(),
			tbtcConfig.Bitcoin.BeneficiaryAddress,
			err,
		)
		return err
	}

	vbyteFee := resolveVbyteFee(bitcoinHandle, tbtcConfig)

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
		logger.Errorf(
			"failed to communicate recovery details for keep [%s]: [%v]",
			keep.ID(),
			err,
		)
		return err
	}

	signer, err := keepsRegistry.GetSigner(keep.ID())
	if err != nil {
		// If there are no signer for loaded keep then something is clearly
		// wrong. We don't want to continue processing for this keep.
		logger.Errorf(
			"no signer for keep [%s]: [%v]",
			keep.ID(),
			err,
		)
		return err
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
		tbtcHandle,
		keep,
		signer,
		chainParams,
		btcAddresses,
		maxFeePerVByte,
	)
	if err != nil {
		logger.Errorf(
			"failed to build the transaction for keep [%s]: [%v]",
			keep.ID(),
			err,
		)
		return err
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
// fail the execution. If a value of vByte fee was not fetched from the bitcoin
// handle the function tries to read it from a config file. If the value is not
// defined in the config file it returns a default vByte fee.
func resolveVbyteFee(bitcoinHandle bitcoin.Handle, tbtcConfig *tbtc.Config) int32 {
	vbyteFee, vbyteFeeError := bitcoinHandle.VbyteFeeFor25Blocks()
	if vbyteFeeError != nil {
		logger.Errorf(
			"failed to retrieve a vbyte fee estimate: [%v]",
			vbyteFeeError,
		)
		// Since the electrs connection is optional, we don't return the error
	}
	if vbyteFee == 0 {
		vbyteFee = tbtcConfig.Bitcoin.MaxFeePerVByte
	}
	if vbyteFee == 0 {
		vbyteFee = defaultVbyteFee
	}

	return vbyteFee
}
