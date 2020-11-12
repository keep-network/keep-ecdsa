package local

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"sync"

	chain "github.com/keep-network/keep-ecdsa/pkg/chain"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
)

const (
	defaultUTXOValue            = 1000
	defaultInitialRedemptionFee = 10
)

type localDeposit struct {
	keepAddress string
	pubkey      []byte
	state       chain.DepositState

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

type localChainLogger struct {
	retrieveSignerPubkeyCalls       int
	provideRedemptionSignatureCalls int
	increaseRedemptionFeeCalls      int
}

func (lcl *localChainLogger) logRetrieveSignerPubkeyCall() {
	lcl.retrieveSignerPubkeyCalls++
}

func (lcl *localChainLogger) RetrieveSignerPubkeyCalls() int {
	return lcl.retrieveSignerPubkeyCalls
}

func (lcl *localChainLogger) logProvideRedemptionSignatureCall() {
	lcl.provideRedemptionSignatureCalls++
}

func (lcl *localChainLogger) ProvideRedemptionSignatureCalls() int {
	return lcl.provideRedemptionSignatureCalls
}

func (lcl *localChainLogger) logIncreaseRedemptionFeeCalls() {
	lcl.increaseRedemptionFeeCalls++
}

func (lcl *localChainLogger) IncreaseRedemptionFeeCalls() int {
	return lcl.increaseRedemptionFeeCalls
}

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

func NewTBTCLocalChain(ctx context.Context) *TBTCLocalChain {
	return &TBTCLocalChain{
		localChain:                            Connect(ctx).(*localChain),
		logger:                                &localChainLogger{},
		alwaysFailingTransactions:             make(map[string]bool),
		deposits:                              make(map[string]*localDeposit),
		depositCreatedHandlers:                make(map[int]func(depositAddress string)),
		depositRegisteredPubkeyHandlers:       make(map[int]func(depositAddress string)),
		depositRedemptionRequestedHandlers:    make(map[int]func(depositAddress string)),
		depositGotRedemptionSignatureHandlers: make(map[int]func(depositAddress string)),
		depositRedeemedHandlers:               make(map[int]func(depositAddress string)),
	}
}

func (tlc *TBTCLocalChain) CreateDeposit(
	depositAddress string,
	signers []common.Address,
) {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	keepAddress := generateAddress()
	tlc.OpenKeep(keepAddress, signers)

	tlc.deposits[depositAddress] = &localDeposit{
		keepAddress:               keepAddress.Hex(),
		state:                     chain.AwaitingSignerSetup,
		utxoValue:                 big.NewInt(defaultUTXOValue),
		redemptionRequestedEvents: make([]*chain.DepositRedemptionRequestedEvent, 0),
	}

	for _, handler := range tlc.depositCreatedHandlers {
		go func(handler func(depositAddress string), depositAddress string) {
			handler(depositAddress)
		}(handler, depositAddress)
	}
}

func (tlc *TBTCLocalChain) OnDepositCreated(
	handler func(depositAddress string),
) (subscription.EventSubscription, error) {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	handlerID := generateHandlerID()

	tlc.depositCreatedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		tlc.tbtcLocalChainMutex.Lock()
		defer tlc.tbtcLocalChainMutex.Unlock()

		delete(tlc.depositCreatedHandlers, handlerID)
	}), nil
}

func (tlc *TBTCLocalChain) OnDepositRegisteredPubkey(
	handler func(depositAddress string),
) (subscription.EventSubscription, error) {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	handlerID := generateHandlerID()

	tlc.depositRegisteredPubkeyHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		tlc.tbtcLocalChainMutex.Lock()
		defer tlc.tbtcLocalChainMutex.Unlock()

		delete(tlc.depositRegisteredPubkeyHandlers, handlerID)
	}), nil
}

func (tlc *TBTCLocalChain) RedeemDeposit(depositAddress string) error {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	deposit, ok := tlc.deposits[depositAddress]
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

	err = tlc.requestSignature(
		common.HexToAddress(deposit.keepAddress),
		deposit.redemptionDigest,
	)
	if err != nil {
		return err
	}

	for _, handler := range tlc.depositRedemptionRequestedHandlers {
		go func(handler func(depositAddress string), depositAddress string) {
			handler(depositAddress)
		}(handler, depositAddress)
	}

	currentBlock, err := tlc.BlockCounter().CurrentBlock()
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

func (tlc *TBTCLocalChain) OnDepositRedemptionRequested(
	handler func(depositAddress string),
) (subscription.EventSubscription, error) {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	handlerID := generateHandlerID()

	tlc.depositRedemptionRequestedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		tlc.tbtcLocalChainMutex.Lock()
		defer tlc.tbtcLocalChainMutex.Unlock()

		delete(tlc.depositRedemptionRequestedHandlers, handlerID)
	}), nil
}

func (tlc *TBTCLocalChain) OnDepositGotRedemptionSignature(
	handler func(depositAddress string),
) (subscription.EventSubscription, error) {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	handlerID := generateHandlerID()

	tlc.depositGotRedemptionSignatureHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		tlc.tbtcLocalChainMutex.Lock()
		defer tlc.tbtcLocalChainMutex.Unlock()

		delete(tlc.depositGotRedemptionSignatureHandlers, handlerID)
	}), nil
}

func (tlc *TBTCLocalChain) OnDepositRedeemed(
	handler func(depositAddress string),
) (subscription.EventSubscription, error) {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	handlerID := generateHandlerID()

	tlc.depositRedeemedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		tlc.tbtcLocalChainMutex.Lock()
		defer tlc.tbtcLocalChainMutex.Unlock()

		delete(tlc.depositRedeemedHandlers, handlerID)
	}), nil
}

func (tlc *TBTCLocalChain) PastDepositRedemptionRequestedEvents(
	depositAddress string,
	startBlock uint64,
) ([]*chain.DepositRedemptionRequestedEvent, error) {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	deposit, ok := tlc.deposits[depositAddress]
	if !ok {
		return nil, fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	return deposit.redemptionRequestedEvents, nil
}

func (tlc *TBTCLocalChain) KeepAddress(depositAddress string) (string, error) {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	deposit, ok := tlc.deposits[depositAddress]
	if !ok {
		return "", fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	return deposit.keepAddress, nil
}

func (tlc *TBTCLocalChain) RetrieveSignerPubkey(depositAddress string) error {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	tlc.logger.logRetrieveSignerPubkeyCall()

	deposit, ok := tlc.deposits[depositAddress]
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
	tlc.localChainMutex.Lock()
	defer tlc.localChainMutex.Unlock()

	keep, ok := tlc.keeps[common.HexToAddress(deposit.keepAddress)]
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

	for _, handler := range tlc.depositRegisteredPubkeyHandlers {
		go func(handler func(depositAddress string), depositAddress string) {
			handler(depositAddress)
		}(handler, depositAddress)
	}

	return nil
}

func (tlc *TBTCLocalChain) ProvideRedemptionSignature(
	depositAddress string,
	v uint8,
	r [32]uint8,
	s [32]uint8,
) error {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	tlc.logger.logProvideRedemptionSignatureCall()

	if _, exists := tlc.alwaysFailingTransactions["ProvideRedemptionSignature"]; exists {
		return fmt.Errorf("always failing transaction")
	}

	deposit, ok := tlc.deposits[depositAddress]
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

	for _, handler := range tlc.depositGotRedemptionSignatureHandlers {
		go func(handler func(depositAddress string), depositAddress string) {
			handler(depositAddress)
		}(handler, depositAddress)
	}

	return nil
}

func (tlc *TBTCLocalChain) IncreaseRedemptionFee(
	depositAddress string,
	previousOutputValueBytes [8]uint8,
	newOutputValueBytes [8]uint8,
) error {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	tlc.logger.logIncreaseRedemptionFeeCalls()

	if _, exists := tlc.alwaysFailingTransactions["IncreaseRedemptionFee"]; exists {
		return fmt.Errorf("always failing transaction")
	}

	deposit, ok := tlc.deposits[depositAddress]
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

	err = tlc.requestSignature(
		common.HexToAddress(deposit.keepAddress),
		deposit.redemptionDigest,
	)
	if err != nil {
		return err
	}

	for _, handler := range tlc.depositRedemptionRequestedHandlers {
		go func(handler func(depositAddress string), depositAddress string) {
			handler(depositAddress)
		}(handler, depositAddress)
	}

	currentBlock, err := tlc.BlockCounter().CurrentBlock()
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

func (tlc *TBTCLocalChain) ProvideRedemptionProof(
	depositAddress string,
	txVersion [4]uint8,
	txInputVector []uint8,
	txOutputVector []uint8,
	txLocktime [4]uint8,
	merkleProof []uint8,
	txIndexInBlock *big.Int,
	bitcoinHeaders []uint8,
) error {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	deposit, ok := tlc.deposits[depositAddress]
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

	for _, handler := range tlc.depositRedeemedHandlers {
		go func(handler func(depositAddress string), depositAddress string) {
			handler(depositAddress)
		}(handler, depositAddress)
	}

	return nil
}

func (tlc *TBTCLocalChain) CurrentState(
	depositAddress string,
) (chain.DepositState, error) {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	deposit, ok := tlc.deposits[depositAddress]
	if !ok {
		return 0, fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	return deposit.state, nil
}

func (tlc *TBTCLocalChain) DepositPubkey(
	depositAddress string,
) ([]byte, error) {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	deposit, ok := tlc.deposits[depositAddress]
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

func (tlc *TBTCLocalChain) DepositRedemptionSignature(
	depositAddress string,
) (*Signature, error) {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	deposit, ok := tlc.deposits[depositAddress]
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

func (tlc *TBTCLocalChain) DepositRedemptionProof(
	depositAddress string,
) (*TxProof, error) {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	deposit, ok := tlc.deposits[depositAddress]
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

func (tlc *TBTCLocalChain) DepositRedemptionFee(
	depositAddress string,
) (*big.Int, error) {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	deposit, ok := tlc.deposits[depositAddress]
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

func (tlc *TBTCLocalChain) SetAlwaysFailingTransactions(transactions ...string) {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	for _, tx := range transactions {
		tlc.alwaysFailingTransactions[tx] = true
	}
}

func (tlc *TBTCLocalChain) Logger() *localChainLogger {
	return tlc.logger
}

func fromLittleEndianBytes(bytes [8]byte) *big.Int {
	return new(big.Int).SetUint64(binary.LittleEndian.Uint64(bytes[:]))
}
