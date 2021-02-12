package local

import (
	"math/big"
	"sync"

	chain "github.com/keep-network/keep-ecdsa/pkg/chain"
)

type localDeposit struct {
	keep   *localKeep
	pubkey []byte
	state  chain.DepositState

	utxoValue           *big.Int
	redemptionDigest    [32]byte
	redemptionFee       *big.Int
	redemptionSignature *Signature
	redemptionProof     *TxProof

	redemptionRequestedEvents []*chain.DepositRedemptionRequestedEvent
}

type Signature struct {
	V uint8
	R [32]uint8
	S [32]uint8
}

type TxProof struct{}

type TBTCLocalChain struct {
	*localChain

	tbtcLocalChainMutex sync.Mutex

	logger *localChainLogger

	alwaysFailingTransactions map[string]bool

	deposits                              map[string]*localDeposit
	depositCreatedHandlers                map[int]func(depositAddress string)
	depositRegisteredPubkeyHandlers       map[int]func(depositAddress string)
	depositRedemptionRequestedHandlers    map[int]func(depositAddress string)
	depositGotRedemptionSignatureHandlers map[int]func(depositAddress string)
	depositRedeemedHandlers               map[int]func(depositAddress string)
}
