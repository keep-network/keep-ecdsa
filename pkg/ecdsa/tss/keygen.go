package tss

import (
	"fmt"
	"math/big"
	"time"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/tss"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-core/pkg/beacon/relay/group"
	"github.com/keep-network/keep-tecdsa/pkg/net"
)

const preParamsGenerationTimeout = 90 * time.Second

var logger = log.Logger("keep-ecdsa")

// GenerateTSSPreParams calculates parameters required by TSS key generation.
// It times out after 90 seconds if the required parameters could not be generated.
// It is possible to generate the parameters way ahead of the TSS protocol
// execution.
// TODO: Consider pre-generating parameters to a pool and use them on protocol
// start.
func GenerateTSSPreParams() (*keygen.LocalPreParams, error) {
	preParams, err := keygen.GeneratePreParams(preParamsGenerationTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tss pre-params: [%v]", err)
	}

	return preParams, nil
}

// InitializeKeyGeneration initializes a member to run a threshold multi-party key
// generation protocol.
//
// It expects unique indices of members in the signing group as well as a group
// size to produce a unique members identifiers.
//
// TSS protocol requires pre-parameters such as safe primes to be generated for
// execution. The parameters should be generated prior to initializing the signer.
//
// Network provider needs to support broadcast and unicast transport.
func InitializeKeyGeneration(
	memberIndex group.MemberIndex,
	groupSize int,
	threshold int,
	tssPreParams *keygen.LocalPreParams,
	networkProvider net.Provider,
) (*Member, error) {
	if memberIndex <= 0 {
		return nil, fmt.Errorf("member index must be greater than 0")
	}

	thisPartyID, groupPartiesIDs := generateGroupPartiesIDs(memberIndex, groupSize)

	errChan := make(chan error)

	keyGenParty, params, endChan, err := initializeKeyGenerationParty(
		thisPartyID,
		groupPartiesIDs,
		threshold,
		tssPreParams,
		networkProvider,
		errChan,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize key generation party: [%v]", err)
	}
	logger.Debugf("initialized key generation party: [%v]", keyGenParty.PartyID())

	signer := &Member{
		keygenParty:     keyGenParty,
		keygenEndChan:   endChan,
		keygenErrChan:   errChan,
		tssParameters:   params,
		networkProvider: networkProvider,
	}

	return signer, nil
}

// Member represents an initialized member who is ready to start distributed key
// generation.
type Member struct {
	tssParameters   *tss.Parameters
	networkProvider net.Provider // network provider used for messages transport
	keygenParty     tss.Party
	// Channels where results of the key generation protocol execution will be written to.
	keygenEndChan <-chan keygen.LocalPartySaveData // data from a successful execution
	keygenErrChan chan error                       // errors emitted during the protocol execution

}

// GenerateKey executes the protocol to generate a signing key. This function
// needs to be executed only after all members finished the initialization stage.
// As a result it will return a Signer who has completed key generation.
func (s *Member) GenerateKey() (*Signer, error) {
	defer unregisterRecv(
		s.networkProvider,
		s.keygenParty,
		s.tssParameters,
		s.keygenErrChan,
	)

	if err := s.keygenParty.Start(); err != nil {
		return nil, fmt.Errorf(
			"failed to start key generation: [%v]",
			s.keygenParty.WrapError(err),
		)
	}

	for {
		select {
		case keygenData := <-s.keygenEndChan:
			signer := &Signer{
				tssParameters: s.tssParameters,
				keygenData:    keygenData,
			}

			return signer, nil
		case err := <-s.keygenErrChan:
			return nil, fmt.Errorf(
				"failed to generate signer key: [%v]",
				s.keygenParty.WrapError(err),
			)
		}
	}
}

func generateGroupPartiesIDs(
	memberIndex group.MemberIndex,
	groupSize int,
) (*tss.PartyID, []*tss.PartyID) {
	var thisPartyID *tss.PartyID
	groupPartiesIDs := []*tss.PartyID{}

	for i := 1; i <= groupSize; i++ {
		newPartyID := tss.NewPartyID(
			string(i),            // id - unique string representing this party in the network
			"",                   // moniker - can be anything (even left blank)
			big.NewInt(int64(i)), // key - unique identifying key
		)

		if memberIndex.Equals(i) {
			thisPartyID = newPartyID
		}

		groupPartiesIDs = append(groupPartiesIDs, newPartyID)
	}

	return thisPartyID, groupPartiesIDs
}

func initializeKeyGenerationParty(
	currentPartyID *tss.PartyID,
	groupPartiesIDs []*tss.PartyID,
	threshold int,
	tssPreParams *keygen.LocalPreParams,
	networkProvider net.Provider,
	errChan chan error,
) (
	tss.Party,
	*tss.Parameters,
	<-chan keygen.LocalPartySaveData,
	error,
) {
	outChan := make(chan tss.Message)
	endChan := make(chan keygen.LocalPartySaveData)

	ctx := tss.NewPeerContext(tss.SortPartyIDs(groupPartiesIDs))
	params := tss.NewParameters(ctx, currentPartyID, len(groupPartiesIDs), threshold)
	party := keygen.NewLocalParty(params, outChan, endChan, *tssPreParams)

	if err := bridgeNetwork(
		networkProvider,
		outChan,
		errChan,
		doneChan,
		party,
		params,
	); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize bridge network for signing: [%v]", err)
	}

	return party, params, endChan, nil
}
