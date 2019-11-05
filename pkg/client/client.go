// Package client defines ECDSA keep client.
package client

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-tecdsa/pkg/registry"
	"github.com/keep-network/keep-tecdsa/pkg/tecdsa"
)

var logger = log.Logger("keep-tecdsa")

// Initialize initializes the tECDSA client with rules related to events handling.
func Initialize(
	ethereumChain eth.Handle,
	persistence persistence.Handle,
) {
	keepsRegistry := registry.NewKeepsRegistry(persistence)

	tecdsa := &tecdsa.TECDSA{
		EthereumChain: ethereumChain,
	}

	// Load current keeps' signers from storage and register for signing events.
	keepsRegistry.LoadExistingKeeps()

	keepsRegistry.ForEachKeep(
		func(keepAddress common.Address, signer *ecdsa.Signer) {
			tecdsa.RegisterForSignEvents(keepAddress, signer)
			logger.Debugf(
				"signer registered for events from keep: [%s]",
				keepAddress.String(),
			)
		},
	)

	// Watch for new keeps creation.
	ethereumChain.OnECDSAKeepCreated(func(event *eth.ECDSAKeepCreatedEvent) {
		logger.Infof(
			"new keep created with address: [%s]",
			event.KeepAddress.String(),
		)

		if event.IsMember(ethereumChain.Address()) {
			go func(keepAddress common.Address) {
				signer, err := tecdsa.GenerateSignerForKeep(event.KeepAddress)
				if err != nil {
					logger.Errorf("signer generation failed: [%v]", err)
					return
				}

				logger.Infof("initialized signer for keep [%s]", keepAddress.String())

				// Store the signer in a map, with the keep address as a key.
				keepsRegistry.RegisterSigner(keepAddress, signer)

				tecdsa.RegisterForSignEvents(keepAddress, signer)
			}(event.KeepAddress)
		}
	})

	// Register client as a candidate member for keep.
	if err := ethereumChain.RegisterAsMemberCandidate(); err != nil {
		logger.Errorf("failed to register member: [%v]", err)
	}

	logger.Infof("client registered as member candidate in keep factory")
}
