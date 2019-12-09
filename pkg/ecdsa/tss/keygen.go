package tss

import (
	"fmt"
	"math/big"
	"time"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/tss"
	"github.com/ipfs/go-log"
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

// InitializeKeyGeneration initializes a signing group member to run a threshold
// multi-party key generation protocol.
//
// It expects unique identifiers of the current member as well as identifiers of
// all members of the signing group.
//
// TSS protocol requires pre-parameters such as safe primes to be generated for
// execution. The parameters should be generated prior to initializing the signer.
func InitializeKeyGeneration(
	groupInfo *GroupInfo,
	tssPreParams *keygen.LocalPreParams,
	networkBridge *NetworkBridge,
) (*Member, error) {
	if len(groupInfo.groupMemberIDs) <= groupInfo.dishonestThreshold {
		return nil, fmt.Errorf(
			"group size [%d], should be greater than dishonest threshold [%d]",
			len(groupInfo.groupMemberIDs),
			groupInfo.dishonestThreshold,
		)
	}

	keyGenParty, endChan, errChan, err := initializeKeyGenerationParty(
		groupInfo,
		tssPreParams,
		networkBridge,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize key generation member: [%v]", err)
	}
	logger.Debugf("initialized key generation member: [%v]", keyGenParty.PartyID())

	return &Member{
		GroupInfo:     groupInfo,
		keygenParty:   keyGenParty,
		keygenEndChan: endChan,
		keygenErrChan: errChan,
		networkBridge: networkBridge,
	}, nil
}

// Member represents an initialized member who is ready to start distributed key
// generation.
type Member struct {
	*GroupInfo

	networkBridge *NetworkBridge // network bridge used for messages transport

	keygenParty tss.Party
	// Channels where results of the key generation protocol execution will be written to.
	keygenEndChan <-chan keygen.LocalPartySaveData // data from a successful execution
	keygenErrChan chan error                       // error from a failed execution
}

// GenerateKey executes the protocol to generate a signing key. This function
// needs to be executed only after all members finished the initialization stage.
// As a result it will return a Signer who has completed key generation, or error
// if the key generation failed.
func (s *Member) GenerateKey() (*Signer, error) {
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
			signer := &Signer{
				GroupInfo:  s.GroupInfo,
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
	if len(groupMemberIDs) < 1 {
		return nil, nil, fmt.Errorf("group should have at least one member")
	}

	var thisPartyID *tss.PartyID
	groupPartiesIDs := []*tss.PartyID{}

	for _, memberID := range groupMemberIDs {
		if memberID.bigInt().Cmp(big.NewInt(0)) <= 0 {
			return nil, nil, fmt.Errorf("member ID must be greater than 0, but found [%v]", memberID.bigInt())
		}

		newPartyID := tss.NewPartyID(
			string(memberID),  // id - unique string representing this party in the network
			"",                // moniker - can be anything (even left blank)
			memberID.bigInt(), // key - unique identifying key
		)

		if thisMemberID == memberID {
			thisPartyID = newPartyID
		}

		groupPartiesIDs = append(groupPartiesIDs, newPartyID)
	}

	return thisPartyID, groupPartiesIDs, nil
}

func initializeKeyGenerationParty(
	groupInfo *GroupInfo,
	tssPreParams *keygen.LocalPreParams,
	bridge *NetworkBridge,
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
		groupInfo.groupID,
		party,
		params,
		tssMessageChan,
		errChan,
	); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect bridge network: [%v]", err)
	}

	return party, endChan, errChan, nil
}
