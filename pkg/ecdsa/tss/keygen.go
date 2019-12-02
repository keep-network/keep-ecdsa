package tss

import (
	"fmt"
	"math/big"
	"time"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/tss"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-tecdsa/pkg/net"
)

const preParamsGenerationTimeout = 90 * time.Second

var logger = log.Logger("keep-ecdsa")

// GenerateTSSPreParams calculates parameters required by TSS key generation.
func GenerateTSSPreParams() (*keygen.LocalPreParams, error) {
	preParams, err := keygen.GeneratePreParams(preParamsGenerationTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tss pre-params: [%v]", err)
	}

	return preParams, nil
}

// InitializeSigner initializes a member to run a threshold multi-party key
// generation protocol.
//
// It expects unique identifiers of members to be provided as `big.Int`.
// List of group members should contain all identifiers of members in the signing
// group including the current member.
//
// TSS protocol requires pre-parameters such as safe primes to be generated for
// execution. The parameters should be generated prior to initializing the signer.
//
// Network channel should support broadcast and unicast messages transport.
func InitializeSigner(
	currentMemberKey *big.Int,
	groupMembersKeys []*big.Int,
	threshold int,
	tssPreParams *keygen.LocalPreParams,
	networkChannel net.NetworkChannel,
) (*Signer, error) {
	var thisPartyID *tss.PartyID
	groupPartiesIDs := []*tss.PartyID{}

	// Generate tss-lib specific parties IDs.
	for index, memberKey := range groupMembersKeys {
		newPartyID := tss.NewPartyID(
			fmt.Sprintf("party-%d", index),   // id - unique string representing this party in the network
			fmt.Sprintf("moniker-%d", index), // moniker - can be anything (even left blank)
			memberKey,                        // key - unique identifying key
		)

		if memberKey.Cmp(currentMemberKey) == 0 {
			thisPartyID = newPartyID
		}

		groupPartiesIDs = append(groupPartiesIDs, newPartyID)
	}

	errChan := make(chan error)

	keyGenParty, endChan := initializeKeyGenerationParty(
		thisPartyID,
		groupPartiesIDs,
		threshold,
		tssPreParams,
		networkChannel,
		errChan,
	)
	logger.Debugf("initialized key generation party: [%v]", keyGenParty.PartyID())

	signer := &Signer{
		keygenParty:   keyGenParty,
		keygenEndChan: endChan,
		keygenErrChan: errChan,
	}

	return signer, nil
}

// GenerateKey executes the protocol to generate a signing key. This function
// needs to be executed only after all members finished the initialization stage.
// As a result the signer will be updated with the key generation data.
func (s *Signer) GenerateKey() error {
	if err := s.keygenParty.Start(); err != nil {
		return fmt.Errorf(
			"failed to start key generation: [%v]",
			s.keygenParty.WrapError(err),
		)
	}

	for {
		select {
		case s.keygenData = <-s.keygenEndChan:
			return nil
		case err := <-s.keygenErrChan:
			return fmt.Errorf(
				"failed to generate signer key: [%v]",
				s.keygenParty.WrapError(err),
			)
		}
	}
}

func initializeKeyGenerationParty(
	currentPartyID *tss.PartyID,
	groupPartiesIDs []*tss.PartyID,
	threshold int,
	tssPreParams *keygen.LocalPreParams,
	networkChannel net.NetworkChannel,
	errChan chan error,
) (tss.Party, <-chan keygen.LocalPartySaveData) {
	recvChan := make(chan net.Message, len(groupPartiesIDs))
	networkChannel.Receive(func(message net.Message) error {
		recvChan <- message
		return nil
	})

	outChan := make(chan tss.Message)
	endChan := make(chan keygen.LocalPartySaveData)

	ctx := tss.NewPeerContext(tss.SortPartyIDs(groupPartiesIDs))
	params := tss.NewParameters(ctx, currentPartyID, len(groupPartiesIDs), threshold)
	party := keygen.NewLocalParty(params, outChan, endChan, *tssPreParams)

	go func() {
		for {
			select {
			case msg := <-outChan:
				bytes, routing, err := msg.WireBytes()
				if err != nil {
					errChan <- fmt.Errorf("failed to encode message: [%v]", party.WrapError(err))
					return
				}

				destinations := routing.To

				if destinations == nil {
					// broadcast
					message := net.Message{
						From:        party.PartyID().Key,
						IsBroadcast: true,
						Payload:     bytes,
					}

					networkChannel.Send(message)
				} else {
					for _, dest := range destinations {
						// unicast
						message := net.Message{
							From:        party.PartyID().Key,
							To:          dest.GetKey(),
							IsBroadcast: false,
							Payload:     bytes,
						}
						networkChannel.Send(message)
					}
				}
			case msg := <-recvChan:
				go func() {
					fromKeyInt := new(big.Int).SetBytes(msg.From)
					fromPartyID := params.Parties().IDs().FindByKey(fromKeyInt)

					if _, err := party.UpdateFromBytes(
						msg.Payload,
						fromPartyID,
						msg.IsBroadcast,
					); err != nil {
						errChan <- party.WrapError(err)
					}
				}()
			}
		}
	}()

	return party, endChan
}
