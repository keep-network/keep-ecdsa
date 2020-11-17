package eventlog

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/abi"
)

// FIXME: This is a temporary structure allowing to access past events
//  emitted by the `BondedECDSAKeep` contract. This structure is
//  here because the generated contract wrappers from `gen/contract`
//  don't support `Filter*` methods yet. When the contract generator
//  will support those methods, the below structure can be removed.
type BondedECDSAKeepEventLog struct {
	contract *abi.BondedECDSAKeep
}

func NewBondedECDSAKeepEventLog(
	contractAddress common.Address,
	backend bind.ContractBackend,
) (*BondedECDSAKeepEventLog, error) {
	contract, err := abi.NewBondedECDSAKeep(contractAddress, backend)
	if err != nil {
		return nil, err
	}

	return &BondedECDSAKeepEventLog{contract}, nil
}

type BondedECDSAKeepSignatureSubmitted struct {
	Digest      [32]byte
	R           [32]byte
	S           [32]byte
	RecoveryID  uint8
	BlockNumber uint64
}

func (bekel *BondedECDSAKeepEventLog) PastSignatureSubmittedEvents(
	startBlock uint64,
	endBlock *uint64,
) ([]*BondedECDSAKeepSignatureSubmitted, error) {
	iterator, err := bekel.contract.FilterSignatureSubmitted(
		&bind.FilterOpts{
			Start: startBlock,
			End:   endBlock,
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	events := make([]*BondedECDSAKeepSignatureSubmitted, 0)

	for {
		if !iterator.Next() {
			break
		}

		event := iterator.Event
		events = append(events, &BondedECDSAKeepSignatureSubmitted{
			Digest:      event.Digest,
			R:           event.R,
			S:           event.S,
			RecoveryID:  event.RecoveryID,
			BlockNumber: event.Raw.BlockNumber,
		})
	}

	return events, nil
}
