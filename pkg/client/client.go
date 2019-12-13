// Package client defines ECDSA keep client.
package client

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss"
	"github.com/keep-network/keep-tecdsa/pkg/net"
	"github.com/keep-network/keep-tecdsa/pkg/registry"
	"github.com/keep-network/keep-tecdsa/pkg/tecdsa"
)

var logger = log.Logger("keep-tecdsa")

// Initialize initializes the tECDSA client with rules related to events handling.
func Initialize(
	ethereumChain eth.Handle,
	networkProvider net.Provider,
	persistence persistence.Handle,
) {
	keepsRegistry := registry.NewKeepsRegistry(persistence)

	tecdsa := tecdsa.NewTECDSA(ethereumChain, networkProvider)

	go func() {
		if err := tecdsa.GenerateTSSPreParams(); err != nil {
			logger.Errorf("failed to initialize tss pre parameters pool: [%v]", err)
		}

		logger.Infof("completed tss pre parameters pool initialization")
	}()

	// Load current keeps' signers from storage and register for signing events.
	keepsRegistry.LoadExistingKeeps()

	keepsRegistry.ForEachKeep(
		func(keepAddress common.Address, signer []*tss.ThresholdSigner) {
			for _, signer := range signer {
				tecdsa.RegisterForSignEvents(keepAddress, signer)
				logger.Debugf(
					"signer registered for events from keep: [%s]",
					keepAddress.String(),
				)
			}
		},
	)

	// Watch for new keeps creation.
	ethereumChain.OnECDSAKeepCreated(func(event *eth.ECDSAKeepCreatedEvent) {
		logger.Infof(
			"new keep [%s] created with members: [%x]\n",
			event.KeepAddress.String(),
			event.Members,
		)

		if event.IsMember(ethereumChain.Address()) {
			signer, err := tecdsa.GenerateSignerForKeep(
				event.KeepAddress,
				event.Members,
			)
			if err != nil {
				logger.Errorf("signer generation failed: [%v]", err)
				return
			}

			logger.Infof("initialized signer for keep [%s]", event.KeepAddress.String())

			err = keepsRegistry.RegisterSigner(event.KeepAddress, signer)
			if err != nil {
				logger.Errorf(
					"failed to register threshold signer for keep [%s]: [%v]",
					event.KeepAddress,
					err,
				)
			}

			tecdsa.RegisterForSignEvents(event.KeepAddress, signer)
		}
	})

	// Register client as a candidate member for keep.
	if err := ethereumChain.RegisterAsMemberCandidate(); err != nil {
		logger.Errorf("failed to register member: [%v]", err)
	}

	logger.Infof("client registered as member candidate in keep factory")
}
