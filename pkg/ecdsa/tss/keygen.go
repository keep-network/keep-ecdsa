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

// NewSigner runs a threshold multi-party key generation protocol to create a
// signer.
//
// It expects unique identifiers of members to be provided as `big.Int`.
// List of group members should contain all identifiers of members in the signing
// group including the current member.
//
// TSS protocol requires pre-parameters such as safe primes to be generated for
// execution. The parameters should be generated prior to initializing the signer.
//
// Network channel should support broadcast and unicast messages transport.
func NewSigner(
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
			fmt.Sprintf("moniker-%d", index), // moniker - can anything (even left blank)
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

	signer := &Signer{tssParty: keyGenParty}

	// TODO: We may want to sync all parties and start key generation at the same
	// time, when all parties finished the initialization.
	err := signer.startKeyGeneration()
	if err != nil {
		return nil, fmt.Errorf("failed to start key generation: [%v]", err)
	}

	for {
		select {
		case signer.keygenData = <-endChan:
			return signer, nil
		case err := <-errChan:
			return nil, fmt.Errorf("failed to generate signer: [%v]", err)
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
	recvChan := make(chan tss.Message, len(groupPartiesIDs))
	networkChannel.Receive(func(message tss.Message) error {
		recvChan <- message
		return nil
	})

	outChan := make(chan tss.Message)
	endChan := make(chan keygen.LocalPartySaveData)

	party := newParty(
		currentPartyID,
		groupPartiesIDs,
		threshold,
		*tssPreParams,
		outChan,
		endChan,
	)

	go func() {
		for {
			select {
			case msg := <-outChan:
				networkChannel.Send(msg)
			case msg := <-recvChan:
				go func() {
					// TODO: We should move marshalling of the message to step
					// where we send the message. This will be handled when we
					// implement proper network channel.
					bytes, _, err := msg.WireBytes()
					if err != nil {
						errChan <- party.WrapError(err)
						return
					}

					if _, err := party.UpdateFromBytes(bytes, msg.GetFrom(), msg.IsBroadcast()); err != nil {
						errChan <- party.WrapError(err)
					}
				}()
			}
		}
	}()

	return party, endChan
}

func (m *Signer) startKeyGeneration() error {
	if err := m.tssParty.Start(); err != nil {
		return m.tssParty.WrapError(err)
	}

	return nil
}
