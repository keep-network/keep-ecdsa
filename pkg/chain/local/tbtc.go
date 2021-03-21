package local

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	"github.com/keep-network/keep-common/pkg/subscription"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
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

// A preset application id for tBTC on the local chain.
var tbtcApplicationID = common.Big1

// Signature represents an ecdsa signature
type Signature struct {
	V uint8
	R [32]uint8
	S [32]uint8
}

// TxProof represents where a transaction proof has been provided or not (nil if not)
type TxProof struct{}

// ChainLogger writes log messages relevant to the local chain
type ChainLogger struct {
	retrieveSignerPubkeyCalls       int
	provideRedemptionSignatureCalls int
	increaseRedemptionFeeCalls      int
	keepAddressCalls                int
}

func (cl *ChainLogger) logRetrieveSignerPubkeyCall() {
	cl.retrieveSignerPubkeyCalls++
}

// RetrieveSignerPubkeyCalls returns the number of times we've tried to retrieve the signer public key
func (cl *ChainLogger) RetrieveSignerPubkeyCalls() int {
	return cl.retrieveSignerPubkeyCalls
}

func (cl *ChainLogger) logProvideRedemptionSignatureCall() {
	cl.provideRedemptionSignatureCalls++
}

// ProvideRedemptionSignatureCalls returns the number of times we've tried to provide the redemption signature
func (cl *ChainLogger) ProvideRedemptionSignatureCalls() int {
	return cl.provideRedemptionSignatureCalls
}

func (cl *ChainLogger) logIncreaseRedemptionFeeCall() {
	cl.increaseRedemptionFeeCalls++
}

// IncreaseRedemptionFeeCalls returns the number of times we've increased the redemption fees
func (cl *ChainLogger) IncreaseRedemptionFeeCalls() int {
	return cl.increaseRedemptionFeeCalls
}

func (cl *ChainLogger) logKeepAddressCall() {
	cl.keepAddressCalls++
}

// KeepAddressCalls returns the number of times we've attempted to retrieve the keep address
func (cl *ChainLogger) KeepAddressCalls() int {
	return cl.keepAddressCalls
}

// TBTCLocalChain represents variables and state relative to the TBTC chain
type TBTCLocalChain struct {
	*localChain

	tbtcLocalChainMutex sync.Mutex

	logger *ChainLogger

	alwaysFailingTransactions map[string]bool

	deposits                              map[string]*localDeposit
	depositCreatedHandlers                map[int]func(depositAddress string)
	depositRegisteredPubkeyHandlers       map[int]func(depositAddress string)
	depositRedemptionRequestedHandlers    map[int]func(depositAddress string)
	depositGotRedemptionSignatureHandlers map[int]func(depositAddress string)
	depositRedeemedHandlers               map[int]func(depositAddress string)
}

func (lc *localChain) TBTCApplicationHandle() (chain.TBTCHandle, error) {
	return NewTBTCLocalChain(context.Background()), nil
}

// NewTBTCLocalChain creates a new TBTCLocalChain
func NewTBTCLocalChain(ctx context.Context) *TBTCLocalChain {
	return &TBTCLocalChain{
		localChain: Connect(ctx).(*localChain),
		logger:     &ChainLogger{},

		alwaysFailingTransactions:             make(map[string]bool),
		deposits:                              make(map[string]*localDeposit),
		depositCreatedHandlers:                make(map[int]func(depositAddress string)),
		depositRegisteredPubkeyHandlers:       make(map[int]func(depositAddress string)),
		depositRedemptionRequestedHandlers:    make(map[int]func(depositAddress string)),
		depositGotRedemptionSignatureHandlers: make(map[int]func(depositAddress string)),
		depositRedeemedHandlers:               make(map[int]func(depositAddress string)),
	}
}

// ID implements the ID method in the chain.TBTCHandle interface.
func (tlc *TBTCLocalChain) ID() chain.ID {
	return localChainID(common.BigToAddress(tbtcApplicationID))
}

// RegisterAsMemberCandidate registers client as a candidate to be selected
// to a keep.
func (tlc *TBTCLocalChain) RegisterAsMemberCandidate() error {
	return nil
}

// IsRegisteredForApplication implements the IsRegisteredForApplication method
// in the chain.TBTCHandle interface.
func (tlc *TBTCLocalChain) IsRegisteredForApplication() (bool, error) {
	panic("implement")
}

// IsEligibleForApplication implements the IsEligibleForApplication method in
// the chain.TBTCHandle interface.
func (tlc *TBTCLocalChain) IsEligibleForApplication() (bool, error) {
	panic("implement")
}

// IsStatusUpToDateForApplication implements the IsStatusUpToDateForApplication
// method in the chain.TBTCHandle interface.
func (lc *localChain) IsStatusUpToDateForApplication() (bool, error) {
	panic("implement")
}

// UpdateStatusForApplication implements the UpdateStatusForApplication method
// in the chain.TBTCHandle interface.
func (tlc *TBTCLocalChain) UpdateStatusForApplication() error {
	panic("implement")
}

// CreateDeposit creates a new deposit by mutating the local TBTC chain
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

// OnDepositCreated installs a callback that is invoked when a
// local-chain notification of a new deposit creation is seen.
func (tlc *TBTCLocalChain) OnDepositCreated(
	handler func(depositAddress string),
) subscription.EventSubscription {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	handlerID := generateHandlerID()

	tlc.depositCreatedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		tlc.tbtcLocalChainMutex.Lock()
		defer tlc.tbtcLocalChainMutex.Unlock()

		delete(tlc.depositCreatedHandlers, handlerID)
	})
}

// OnDepositRegisteredPubkey installs a callback that is invoked when a
// local-chain notification of a deposit registration is seen.
func (tlc *TBTCLocalChain) OnDepositRegisteredPubkey(
	handler func(depositAddress string),
) subscription.EventSubscription {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	handlerID := generateHandlerID()

	tlc.depositRegisteredPubkeyHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		tlc.tbtcLocalChainMutex.Lock()
		defer tlc.tbtcLocalChainMutex.Unlock()

		delete(tlc.depositRegisteredPubkeyHandlers, handlerID)
	})
}

// RedeemDeposit initiates the redemption process which involves trading the
// system back the minted TBTC in exhange for the underlying BTC.
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

	err = tlc.RequestSignature(
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

// OnDepositRedemptionRequested installs a callback that is invoked when a
// redemption is requested.
func (tlc *TBTCLocalChain) OnDepositRedemptionRequested(
	handler func(depositAddress string),
) subscription.EventSubscription {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	handlerID := generateHandlerID()

	tlc.depositRedemptionRequestedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		tlc.tbtcLocalChainMutex.Lock()
		defer tlc.tbtcLocalChainMutex.Unlock()

		delete(tlc.depositRedemptionRequestedHandlers, handlerID)
	})
}

// OnDepositGotRedemptionSignature installs a callback that is invoked when the
// signers sign off on a redemption
func (tlc *TBTCLocalChain) OnDepositGotRedemptionSignature(
	handler func(depositAddress string),
) subscription.EventSubscription {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	handlerID := generateHandlerID()

	tlc.depositGotRedemptionSignatureHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		tlc.tbtcLocalChainMutex.Lock()
		defer tlc.tbtcLocalChainMutex.Unlock()

		delete(tlc.depositGotRedemptionSignatureHandlers, handlerID)
	})
}

// OnDepositRedeemed installs a callback that is invoked when the redemption
// process is successful
func (tlc *TBTCLocalChain) OnDepositRedeemed(
	handler func(depositAddress string),
) subscription.EventSubscription {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	handlerID := generateHandlerID()

	tlc.depositRedeemedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		tlc.tbtcLocalChainMutex.Lock()
		defer tlc.tbtcLocalChainMutex.Unlock()

		delete(tlc.depositRedeemedHandlers, handlerID)
	})
}

// PastDepositRedemptionRequestedEvents the redemption requested events relevant to a particular deposit
func (tlc *TBTCLocalChain) PastDepositRedemptionRequestedEvents(
	startBlock uint64,
	depositAddress string,
) ([]*chain.DepositRedemptionRequestedEvent, error) {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	deposit, ok := tlc.deposits[depositAddress]
	if !ok {
		return nil, fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	return deposit.redemptionRequestedEvents, nil
}

// Keep returns the keep for a particular deposit
func (tlc *TBTCLocalChain) Keep(depositAddress string) (chain.BondedECDSAKeepHandle, error) {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	tlc.logger.logKeepAddressCall()

	deposit, ok := tlc.deposits[depositAddress]
	if !ok {
		return nil, fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	return tlc.GetKeepWithID(
		localChainID(common.HexToAddress(deposit.keepAddress)),
	)
}

// RetrieveSignerPubkey enriches the referenced deposit with the signer public
// key and moves the state to AwaitingBtcFundingProof
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

// ProvideRedemptionSignature enriches the deposit with a redemption signature
// and moves the state to AwaitingWithdrawalProof
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

// IncreaseRedemptionFee sets the remeption fee to `newOutputValueBytes` and
// uses `previousOutputValueBytes` for validation.
func (tlc *TBTCLocalChain) IncreaseRedemptionFee(
	depositAddress string,
	previousOutputValueBytes [8]uint8,
	newOutputValueBytes [8]uint8,
) error {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	tlc.logger.logIncreaseRedemptionFeeCall()

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

	err = tlc.RequestSignature(
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

// ProvideRedemptionProof sets the redemption proof on a deposit and updates the state to Redeemed
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

// CurrentState returns the state of a particular deposit
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

// DepositPubkey returns the public key of a particular deposit
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

// DepositRedemptionSignature returns the redemption signature of a particular deposit
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

// DepositRedemptionProof returns the redemption proof of a particular deposit
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

// DepositRedemptionFee returns the redemption fee of a particular deposit
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

// SetAlwaysFailingTransactions adds the supplied transactions to collection of always failing transactions
func (tlc *TBTCLocalChain) SetAlwaysFailingTransactions(transactions ...string) {
	tlc.tbtcLocalChainMutex.Lock()
	defer tlc.tbtcLocalChainMutex.Unlock()

	for _, tx := range transactions {
		tlc.alwaysFailingTransactions[tx] = true
	}
}

// Logger surfaces the chain's logger
func (tlc *TBTCLocalChain) Logger() *ChainLogger {
	return tlc.logger
}

func fromLittleEndianBytes(bytes [8]byte) *big.Int {
	return new(big.Int).SetUint64(binary.LittleEndian.Uint64(bytes[:]))
}
