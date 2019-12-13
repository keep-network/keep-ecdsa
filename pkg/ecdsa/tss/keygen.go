package tss

import (
	"fmt"
	"math/big"
	"time"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/tss"
)

const preParamsGenerationTimeout = 90 * time.Second

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

// initializeKeyGeneration initializes a signing group member to run a threshold
// multi-party key generation protocol.
//
// It expects unique identifiers of the current member as well as identifiers of
// all members of the signing group.
//
// TSS protocol requires pre-parameters such as safe primes to be generated for
// execution. The parameters should be generated prior to initializing the signer.
func initializeKeyGeneration(
	group *groupInfo,
	tssPreParams *keygen.LocalPreParams,
	network *networkBridge,
) (*member, error) {
	keyGenParty, endChan, errChan, err := initializeKeyGenerationParty(
		group,
		tssPreParams,
		network,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize key generation member: [%v]", err)
	}
	logger.Debugf("initialized key generation member: [%v]", keyGenParty.PartyID())

	return &member{
		groupInfo:     group,
		keygenParty:   keyGenParty,
		keygenEndChan: endChan,
		keygenErrChan: errChan,
		networkBridge: network,
	}, nil
}

// member represents an initialized member who is ready to start distributed key
// generation.
type member struct {
	*groupInfo

	networkBridge *networkBridge // network bridge used for messages transport

	keygenParty tss.Party
	// Channels where results of the key generation protocol execution will be written to.
	keygenEndChan <-chan keygen.LocalPartySaveData // data from a successful execution
	keygenErrChan chan error                       // error from a failed execution
}

// generateKey executes the protocol to generate a signing key. This function
// needs to be executed only after all members finished the initialization stage.
// As a result it will return a Signer who has completed key generation, or error
// if the key generation failed.
func (s *member) generateKey() (*ThresholdSigner, error) {
	defer s.networkBridge.close()

	if err := s.keygenParty.Start(); err != nil {
		return nil, fmt.Errorf(
			"failed to start key generation: [%v]",
			s.keygenParty.WrapError(err),
		)
	}

	for {
		select {
		case keygenData := <-s.keygenEndChan:
			signer := &ThresholdSigner{
				groupInfo:  s.groupInfo,
				keygenData: keygenData,
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

func generatePartiesIDs(
	thisMemberID MemberID,
	groupMemberIDs []MemberID,
) (
	*tss.PartyID,
	[]*tss.PartyID,
	error,
) {
	var thisPartyID *tss.PartyID
	groupPartiesIDs := []*tss.PartyID{}

	for _, memberID := range groupMemberIDs {
		if memberID.BigInt().Cmp(big.NewInt(0)) <= 0 {
			return nil, nil, fmt.Errorf("member ID must be greater than 0, but found [%v]", memberID.BigInt())
		}

		newPartyID := tss.NewPartyID(
			memberID.String(), // id - unique string representing this party in the network
			"",                // moniker - can be anything (even left blank)
			memberID.BigInt(), // key - unique identifying key
		)

		if thisMemberID == memberID {
			thisPartyID = newPartyID
		}

		groupPartiesIDs = append(groupPartiesIDs, newPartyID)
	}

	return thisPartyID, groupPartiesIDs, nil
}

func initializeKeyGenerationParty(
	groupInfo *groupInfo,
	tssPreParams *keygen.LocalPreParams,
	bridge *networkBridge,
) (
	tss.Party,
	<-chan keygen.LocalPartySaveData,
	chan error,
	error,
) {
	tssMessageChan := make(chan tss.Message, len(groupInfo.groupMemberIDs))
	endChan := make(chan keygen.LocalPartySaveData)
	errChan := make(chan error)

	currentPartyID, groupPartiesIDs, err := generatePartiesIDs(
		groupInfo.memberID,
		groupInfo.groupMemberIDs,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate parties IDs: [%v]", err)
	}

	params := tss.NewParameters(
		tss.NewPeerContext(tss.SortPartyIDs(groupPartiesIDs)),
		currentPartyID,
		len(groupPartiesIDs),
		groupInfo.dishonestThreshold,
	)

	party := keygen.NewLocalParty(params, tssMessageChan, endChan, *tssPreParams)

	if err := bridge.connect(
		tssMessageChan,
	); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect bridge network: [%v]", err)
	}

	bridge.registerTSSMessageHandler(party, params.Parties().IDs())

	return party, endChan, errChan, nil
}
