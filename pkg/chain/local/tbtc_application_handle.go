package local

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
	chain "github.com/keep-network/keep-ecdsa/pkg/chain"
)

const (
	defaultUTXOValue            = 1000
	defaultInitialRedemptionFee = 10
)

var tbtcApplicationID = common.Big1

type tbtcApplicationHandle struct {
	handle *localChain

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

func (lc *localChain) TBTCApplicationHandle() (chain.TBTCHandle, error) {
	return &tbtcApplicationHandle{
		handle: lc,

		logger:                                &localChainLogger{},

		alwaysFailingTransactions:             make(map[string]bool),
		deposits:                              make(map[string]*localDeposit),
		depositCreatedHandlers:                make(map[int]func(depositAddress string)),
		depositRegisteredPubkeyHandlers:       make(map[int]func(depositAddress string)),
		depositRedemptionRequestedHandlers:    make(map[int]func(depositAddress string)),
		depositGotRedemptionSignatureHandlers: make(map[int]func(depositAddress string)),
		depositRedeemedHandlers:               make(map[int]func(depositAddress string)),
	}, nil
}

func (*tbtcApplicationHandle) ID() (chain.KeepApplicationID) {
	return combinedChainID(
		common.BigToAddress(tbtcApplicationID),
	)
}

func (*tbtcApplicationHandle) IsRegisteredForApplication() (bool, error) {
	panic("implement")
}

func (*tbtcApplicationHandle) IsEligibleForApplication() (bool, error) {
	panic("implement")
}

func (*tbtcApplicationHandle) IsStatusUpToDateForApplication() (bool, error) {
	panic("implement")
}

func (*tbtcApplicationHandle) UpdateStatusForApplication() error {
	panic("implement")
}

// RegisterAsMemberCandidate registers client as a candidate to be selected
// to a keep.
func (tah *tbtcApplicationHandle) RegisterAsMemberCandidate() error {
	return nil
}

func (tah *tbtcApplicationHandle) IsOperatorAuthorized(operator chain.OperatorID) (bool, error) {
	tah.handle.localChainMutex.Lock()
	defer tah.handle.localChainMutex.Unlock()

	return tah.handle.authorizations[operator], nil
}

func (tah *tbtcApplicationHandle) CreateDeposit(
	depositAddress string,
	signers []common.Address,
) {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	keepAddress := generateAddress()
	keep, err := tah.handle.createKeepWithMembers(keepAddress, signers)
	if err != nil {
		panic(err)
	}

	tah.deposits[depositAddress] = &localDeposit{
		keep:                      keep,
		state:                     chain.AwaitingSignerSetup,
		utxoValue:                 big.NewInt(defaultUTXOValue),
		redemptionRequestedEvents: make([]*chain.DepositRedemptionRequestedEvent, 0),
	}

	for _, handler := range tah.depositCreatedHandlers {
		go func(handler func(depositAddress string), depositAddress string) {
			handler(depositAddress)
		}(handler, depositAddress)
	}
}

func (tah *tbtcApplicationHandle) OnDepositCreated(
	handler func(depositAddress string),
) subscription.EventSubscription {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	handlerID := generateHandlerID()

	tah.depositCreatedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		tah.tbtcLocalChainMutex.Lock()
		defer tah.tbtcLocalChainMutex.Unlock()

		delete(tah.depositCreatedHandlers, handlerID)
	})
}

func (tah *tbtcApplicationHandle) OnDepositRegisteredPubkey(
	handler func(depositAddress string),
) subscription.EventSubscription {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	handlerID := generateHandlerID()

	tah.depositRegisteredPubkeyHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		tah.tbtcLocalChainMutex.Lock()
		defer tah.tbtcLocalChainMutex.Unlock()

		delete(tah.depositRegisteredPubkeyHandlers, handlerID)
	})
}

func (tah *tbtcApplicationHandle) RedeemDeposit(depositAddress string) error {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	deposit, ok := tah.deposits[depositAddress]
	if !ok {
		return fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	if !bytes.Equal(
		deposit.redemptionDigest[:],
		make([]byte, len(deposit.redemptionDigest)),
	) {
		return fmt.Errorf(
			"redemption of deposit [%v] already requested",
			depositAddress,
		)
	}

	var randomDigest [32]byte
	// #nosec G404 (insecure random number source (rand))
	// Local chain implementation doesn't require secure randomness.
	_, err := rand.Read(randomDigest[:])
	if err != nil {
		return err
	}

	deposit.state = chain.AwaitingWithdrawalSignature
	deposit.redemptionDigest = randomDigest
	deposit.redemptionFee = big.NewInt(defaultInitialRedemptionFee)

	err = deposit.keep.RequestSignature(
		deposit.redemptionDigest,
	)
	if err != nil {
		return err
	}

	for _, handler := range tah.depositRedemptionRequestedHandlers {
		go func(handler func(depositAddress string), depositAddress string) {
			handler(depositAddress)
		}(handler, depositAddress)
	}

	blockCounter, _ := tah.handle.BlockCounter() // localChain BlockCounter never errors
	currentBlock, err := blockCounter.CurrentBlock()
	if err != nil {
		return err
	}

	deposit.redemptionRequestedEvents = append(
		deposit.redemptionRequestedEvents,
		&chain.DepositRedemptionRequestedEvent{
			DepositAddress:       depositAddress,
			Digest:               deposit.redemptionDigest,
			UtxoValue:            deposit.utxoValue,
			RedeemerOutputScript: nil,
			RequestedFee:         deposit.redemptionFee,
			Outpoint:             nil,
			BlockNumber:          currentBlock,
		},
	)

	return nil
}

func (tah *tbtcApplicationHandle) OnDepositRedemptionRequested(
	handler func(depositAddress string),
) subscription.EventSubscription {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	handlerID := generateHandlerID()

	tah.depositRedemptionRequestedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		tah.tbtcLocalChainMutex.Lock()
		defer tah.tbtcLocalChainMutex.Unlock()

		delete(tah.depositRedemptionRequestedHandlers, handlerID)
	})
}

func (tah *tbtcApplicationHandle) OnDepositGotRedemptionSignature(
	handler func(depositAddress string),
) subscription.EventSubscription {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	handlerID := generateHandlerID()

	tah.depositGotRedemptionSignatureHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		tah.tbtcLocalChainMutex.Lock()
		defer tah.tbtcLocalChainMutex.Unlock()

		delete(tah.depositGotRedemptionSignatureHandlers, handlerID)
	})
}

func (tah *tbtcApplicationHandle) OnDepositRedeemed(
	handler func(depositAddress string),
) subscription.EventSubscription {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	handlerID := generateHandlerID()

	tah.depositRedeemedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		tah.tbtcLocalChainMutex.Lock()
		defer tah.tbtcLocalChainMutex.Unlock()

		delete(tah.depositRedeemedHandlers, handlerID)
	})
}

func (tah *tbtcApplicationHandle) PastDepositRedemptionRequestedEvents(
	startBlock uint64,
	depositAddress string,
) ([]*chain.DepositRedemptionRequestedEvent, error) {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	deposit, ok := tah.deposits[depositAddress]
	if !ok {
		return nil, fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	return deposit.redemptionRequestedEvents, nil
}

func (tah *tbtcApplicationHandle) Keep(depositAddress string) (chain.BondedECDSAKeepHandle, error) {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	tah.logger.logKeepAddressCall()

	deposit, ok := tah.deposits[depositAddress]
	if !ok {
		return nil, fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	return deposit.keep, nil
}

func (tah *tbtcApplicationHandle) RetrieveSignerPubkey(depositAddress string) error {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	tah.logger.logRetrieveSignerPubkeyCall()

	deposit, ok := tah.deposits[depositAddress]
	if !ok {
		return fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	if len(deposit.pubkey) > 0 {
		return fmt.Errorf(
			"pubkey for deposit [%v] already retrieved",
			depositAddress,
		)
	}

	// lock upstream mutex to access `keeps` map safely
	tah.handle.localChainMutex.Lock()
	defer tah.handle.localChainMutex.Unlock()

	keep, ok := tah.handle.keeps[common.HexToAddress(deposit.keep.address.String())]
	if !ok {
		return fmt.Errorf(
			"could not find keep for deposit [%v]",
			depositAddress,
		)
	}

	if len(keep.publicKey[:]) == 0 ||
		bytes.Equal(keep.publicKey[:], make([]byte, len(keep.publicKey))) {
		return fmt.Errorf(
			"keep of deposit [%v] doesn't have a public key yet",
			depositAddress,
		)
	}

	deposit.pubkey = keep.publicKey[:]
	deposit.state = chain.AwaitingBtcFundingProof

	for _, handler := range tah.depositRegisteredPubkeyHandlers {
		go func(handler func(depositAddress string), depositAddress string) {
			handler(depositAddress)
		}(handler, depositAddress)
	}

	return nil
}

func (tah *tbtcApplicationHandle) ProvideRedemptionSignature(
	depositAddress string,
	v uint8,
	r [32]uint8,
	s [32]uint8,
) error {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	tah.logger.logProvideRedemptionSignatureCall()

	if _, exists := tah.alwaysFailingTransactions["ProvideRedemptionSignature"]; exists {
		return fmt.Errorf("always failing transaction")
	}

	deposit, ok := tah.deposits[depositAddress]
	if !ok {
		return fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	if deposit.redemptionDigest == [32]byte{} {
		return fmt.Errorf("deposit [%v] is not in redemption", depositAddress)
	}

	if deposit.redemptionSignature != nil {
		return fmt.Errorf(
			"redemption signature for deposit [%v] already provided",
			depositAddress,
		)
	}

	deposit.state = chain.AwaitingWithdrawalProof
	deposit.redemptionSignature = &Signature{
		V: v,
		R: r,
		S: s,
	}

	for _, handler := range tah.depositGotRedemptionSignatureHandlers {
		go func(handler func(depositAddress string), depositAddress string) {
			handler(depositAddress)
		}(handler, depositAddress)
	}

	return nil
}

func (tah *tbtcApplicationHandle) IncreaseRedemptionFee(
	depositAddress string,
	previousOutputValueBytes [8]uint8,
	newOutputValueBytes [8]uint8,
) error {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	tah.logger.logIncreaseRedemptionFeeCall()

	if _, exists := tah.alwaysFailingTransactions["IncreaseRedemptionFee"]; exists {
		return fmt.Errorf("always failing transaction")
	}

	deposit, ok := tah.deposits[depositAddress]
	if !ok {
		return fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	if deposit.redemptionSignature == nil {
		return fmt.Errorf(
			"no redemption signature for deposit [%v]; could not increase fee",
			depositAddress,
		)
	}

	previousOutputValue := fromLittleEndianBytes(previousOutputValueBytes)
	expectedPreviousOutputValue := new(big.Int).Sub(
		deposit.utxoValue,
		deposit.redemptionFee,
	)

	if expectedPreviousOutputValue.Cmp(previousOutputValue) != 0 {
		return fmt.Errorf("wrong previous output value")
	}

	newOutputValue := fromLittleEndianBytes(newOutputValueBytes)

	if new(big.Int).Sub(previousOutputValue, newOutputValue).Cmp(
		big.NewInt(defaultInitialRedemptionFee),
	) != 0 {
		return fmt.Errorf("wrong increase fee step")
	}

	var randomDigest [32]byte
	// #nosec G404 (insecure random number source (rand))
	// Local chain implementation doesn't require secure randomness.
	_, err := rand.Read(randomDigest[:])
	if err != nil {
		return err
	}

	deposit.state = chain.AwaitingWithdrawalSignature
	deposit.redemptionDigest = randomDigest
	deposit.redemptionFee = new(big.Int).Sub(deposit.utxoValue, newOutputValue)
	deposit.redemptionSignature = nil

	err = deposit.keep.RequestSignature(
		deposit.redemptionDigest,
	)
	if err != nil {
		return err
	}

	for _, handler := range tah.depositRedemptionRequestedHandlers {
		go func(handler func(depositAddress string), depositAddress string) {
			handler(depositAddress)
		}(handler, depositAddress)
	}

	blockCounter, _ := tah.handle.BlockCounter() // localChain BlockCounter never errors
	currentBlock, err := blockCounter.CurrentBlock()
	if err != nil {
		return err
	}

	deposit.redemptionRequestedEvents = append(
		deposit.redemptionRequestedEvents,
		&chain.DepositRedemptionRequestedEvent{
			DepositAddress:       depositAddress,
			Digest:               deposit.redemptionDigest,
			UtxoValue:            deposit.utxoValue,
			RedeemerOutputScript: nil,
			RequestedFee:         deposit.redemptionFee,
			Outpoint:             nil,
			BlockNumber:          currentBlock,
		},
	)

	return nil
}

func (tah *tbtcApplicationHandle) ProvideRedemptionProof(
	depositAddress string,
	txVersion [4]uint8,
	txInputVector []uint8,
	txOutputVector []uint8,
	txLocktime [4]uint8,
	merkleProof []uint8,
	txIndexInBlock *big.Int,
	bitcoinHeaders []uint8,
) error {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	deposit, ok := tah.deposits[depositAddress]
	if !ok {
		return fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	if deposit.redemptionProof != nil {
		return fmt.Errorf(
			"redemption proof for deposit [%v] already provided",
			depositAddress,
		)
	}

	deposit.state = chain.Redeemed
	deposit.redemptionProof = &TxProof{}

	for _, handler := range tah.depositRedeemedHandlers {
		go func(handler func(depositAddress string), depositAddress string) {
			handler(depositAddress)
		}(handler, depositAddress)
	}

	return nil
}

func (tah *tbtcApplicationHandle) CurrentState(
	depositAddress string,
) (chain.DepositState, error) {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	deposit, ok := tah.deposits[depositAddress]
	if !ok {
		return 0, fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	return deposit.state, nil
}

func (tah *tbtcApplicationHandle) DepositPubkey(
	depositAddress string,
) ([]byte, error) {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	deposit, ok := tah.deposits[depositAddress]
	if !ok {
		return nil, fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	if len(deposit.pubkey) == 0 {
		return nil, fmt.Errorf(
			"no pubkey for deposit [%v]",
			depositAddress,
		)
	}

	return deposit.pubkey, nil
}

func (tah *tbtcApplicationHandle) DepositRedemptionSignature(
	depositAddress string,
) (*Signature, error) {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	deposit, ok := tah.deposits[depositAddress]
	if !ok {
		return nil, fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	if deposit.redemptionSignature == nil {
		return nil, fmt.Errorf(
			"no redemption signature for deposit [%v]",
			depositAddress,
		)
	}

	return deposit.redemptionSignature, nil
}

func (tah *tbtcApplicationHandle) DepositRedemptionProof(
	depositAddress string,
) (*TxProof, error) {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	deposit, ok := tah.deposits[depositAddress]
	if !ok {
		return nil, fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	if deposit.redemptionProof == nil {
		return nil, fmt.Errorf(
			"no redemption proof for deposit [%v]",
			depositAddress,
		)
	}

	return deposit.redemptionProof, nil
}

func (tah *tbtcApplicationHandle) DepositRedemptionFee(
	depositAddress string,
) (*big.Int, error) {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	deposit, ok := tah.deposits[depositAddress]
	if !ok {
		return nil, fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	if deposit.redemptionFee == nil {
		return nil, fmt.Errorf(
			"no redemption fee for deposit [%v]",
			depositAddress,
		)
	}

	return deposit.redemptionFee, nil
}

func (tah *tbtcApplicationHandle) SetAlwaysFailingTransactions(transactions ...string) {
	tah.tbtcLocalChainMutex.Lock()
	defer tah.tbtcLocalChainMutex.Unlock()

	for _, tx := range transactions {
		tah.alwaysFailingTransactions[tx] = true
	}
}

func (tah *tbtcApplicationHandle) Logger() *localChainLogger {
	return tah.logger
}

func fromLittleEndianBytes(bytes [8]byte) *big.Int {
	return new(big.Int).SetUint64(binary.LittleEndian.Uint64(bytes[:]))
}
