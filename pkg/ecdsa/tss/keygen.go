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
	broadcastChannel net.BroadcastChannel,
	unicastChannels map[string]net.UnicastChannel,
) (*Signer, error) {
	thisPartyID, groupPartiesIDs := generateGroupPartiesIDs(currentMemberKey, groupMembersKeys)

	errChan := make(chan error)

	keyGenParty, endChan := initializeKeyGenerationParty(
		thisPartyID,
		groupPartiesIDs,
		threshold,
		tssPreParams,
		broadcastChannel,
		unicastChannels,
		errChan,
	)
	logger.Debugf("initialized key generation party: [%v]", keyGenParty.PartyID())

	signer := &Signer{
		keygenParty:      keyGenParty,
		broadcastChannel: broadcastChannel,
		unicastChannels:  unicastChannels,
		keygenEndChan:    endChan,
		keygenErrChan:    errChan,
	}

	return signer, nil
}

// GenerateKey executes the protocol to generate a signing key. This function
// needs to be executed only after all members finished the initialization stage.
// As a result the signer will be updated with the key generation data.
func (s *Signer) GenerateKey() error {
	defer s.broadcastChannel.UnregisterRecv(TSSmessageType)

	for _, unicastChannel := range s.unicastChannels {
		defer unicastChannel.UnregisterRecv(TSSmessageType)
	}

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

func generateGroupPartiesIDs(currentMemberKey *big.Int, groupMembersKeys []*big.Int) (*tss.PartyID, []*tss.PartyID) {
	var thisPartyID *tss.PartyID
	groupPartiesIDs := []*tss.PartyID{}

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

	return thisPartyID, groupPartiesIDs
}

func initializeKeyGenerationParty(
	currentPartyID *tss.PartyID,
	groupPartiesIDs []*tss.PartyID,
	threshold int,
	tssPreParams *keygen.LocalPreParams,
	broadcastChannel net.BroadcastChannel,
	unicastChannels map[string]net.UnicastChannel,
	errChan chan error,
) (tss.Party, <-chan keygen.LocalPartySaveData) {
	recvMessage := make(chan *TSSMessage, len(groupPartiesIDs))

	handleMessageFunc := func(channel chan *TSSMessage) net.HandleMessageFunc {
		return net.HandleMessageFunc{
			Type: TSSmessageType,
			Handler: func(msg net.Message) error {
				switch tssMessage := msg.Payload().(type) {
				case *TSSMessage:
					channel <- tssMessage
				default:
					return fmt.Errorf("unexpected message: [%v]", msg.Payload())
				}

				return nil
			},
		}
	}

	broadcastChannel.Recv(handleMessageFunc(recvMessage))

	for _, unicastChannel := range unicastChannels {
		unicastChannel.Recv(handleMessageFunc(recvMessage))
	}

	outChan := make(chan tss.Message)
	endChan := make(chan keygen.LocalPartySaveData)

	ctx := tss.NewPeerContext(tss.SortPartyIDs(groupPartiesIDs))
	params := tss.NewParameters(ctx, currentPartyID, len(groupPartiesIDs), threshold)
	party := keygen.NewLocalParty(params, outChan, endChan, *tssPreParams)

	go func() {
		for {
			select {
			case tssLibMsg := <-outChan:
				bytes, routing, err := tssLibMsg.WireBytes()
				if err != nil {
					errChan <- fmt.Errorf("failed to encode message: [%v]", party.WrapError(err))
					break
				}

				msg := &TSSMessage{
					SenderID:    routing.From.GetKey(),
					Payload:     bytes,
					IsBroadcast: routing.IsBroadcast,
				}

				if routing.To == nil {
					broadcastChannel.Send(msg)
				} else {
					for _, destination := range routing.To {
						peerID := MemberID(destination.GetKey())

						unicastChannel, ok := unicastChannels[peerID.String()]
						if !ok {
							errChan <- fmt.Errorf("failed to find unicast channel for: [%v]", peerID)
							continue
						}

						unicastChannel.Send(msg)
					}
				}
			case msg := <-recvMessage:
				go func() {
					senderKey := new(big.Int).SetBytes(msg.SenderID)
					senderPartyID := params.Parties().IDs().FindByKey(senderKey)

					bytes := msg.Payload

					if _, err := party.UpdateFromBytes(
						bytes,
						senderPartyID,
						msg.IsBroadcast,
					); err != nil {
						errChan <- party.WrapError(err)
						return
					}
				}()
			}
		}
	}()

	return party, endChan
}
