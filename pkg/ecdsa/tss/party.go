package tss

import (
	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	tssLib "github.com/binance-chain/tss-lib/tss"
)

func newParty(
	partyID *tssLib.PartyID,
	groupPartiesIDs []*tssLib.PartyID,
	threshold int,
	preParams keygen.LocalPreParams,
	chanOut chan tssLib.Message,
	chanEnd chan keygen.LocalPartySaveData,
) tssLib.Party {
	ctx := tssLib.NewPeerContext(tssLib.SortPartyIDs(groupPartiesIDs))
	params := tssLib.NewParameters(ctx, partyID, len(groupPartiesIDs), threshold)

	return keygen.NewLocalParty(params, chanOut, chanEnd, preParams)
}
