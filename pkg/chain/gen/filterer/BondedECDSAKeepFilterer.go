package filterer

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-ecdsa/pkg/chain/gen/abi"
)

// FIXME: This is a temporary structure allowing to access past events
//  occurred emitted by the `BondedECDSAKeep` contract. This structure is
//  here because the generated contract wrappers from `gen/contract`
//  don't support `Filter*` methods yet. When the contract generator
//  will support those methods, the below structure can be removed.
type BondedECDSAKeepFilterer struct {
	contract *abi.BondedECDSAKeep
}

func NewBondedECDSAKeepFilterer(
	contractAddress common.Address,
	backend bind.ContractBackend,
) (*BondedECDSAKeepFilterer, error) {
	contract, err := abi.NewBondedECDSAKeep(contractAddress, backend)
	if err != nil {
		return nil, err
	}

	return &BondedECDSAKeepFilterer{contract}, nil
}

type BondedECDSAKeepSignatureSubmitted struct {
	Digest     [32]byte
	R          [32]byte
	S          [32]byte
	RecoveryID uint8
}

func (bekf *BondedECDSAKeepFilterer) FilterSignatureSubmitted(
	startBlock uint64,
	endBlock *uint64,
) ([]*BondedECDSAKeepSignatureSubmitted, error) {
	iterator, err := bekf.contract.FilterSignatureSubmitted(
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
			Digest:     event.Digest,
			R:          event.R,
			S:          event.S,
			RecoveryID: event.RecoveryID,
		})
	}

	return events, nil
}
