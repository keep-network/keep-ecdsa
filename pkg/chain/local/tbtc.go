package local

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-common/pkg/subscription"
)

type localDeposit struct {
	keepAddress string
	pubkey      []byte
}

type localChainLogger struct {
	retrieveSignerPubkeyCalls int
}

func (lcl *localChainLogger) logRetrieveSignerPubkeyCall() {
	lcl.retrieveSignerPubkeyCalls++
}

func (lcl *localChainLogger) RetrieveSignerPubkeyCalls() int {
	return lcl.retrieveSignerPubkeyCalls
}

type TBTCLocalChain struct {
	*localChain

	mutex sync.Mutex

	logger *localChainLogger

	deposits                        map[string]*localDeposit
	depositCreatedHandlers          map[int]func(depositAddress string)
	depositRegisteredPubkeyHandlers map[int]func(depositAddress string)
}

func NewTBTCLocalChain() *TBTCLocalChain {
	return &TBTCLocalChain{
		localChain:                      Connect().(*localChain),
		logger:                          &localChainLogger{},
		deposits:                        make(map[string]*localDeposit),
		depositCreatedHandlers:          make(map[int]func(depositAddress string)),
		depositRegisteredPubkeyHandlers: make(map[int]func(depositAddress string)),
	}
}

func (tlc *TBTCLocalChain) CreateDeposit(depositAddress string) {
	tlc.mutex.Lock()
	defer tlc.mutex.Unlock()

	keepAddress := generateAddress()
	tlc.OpenKeep(keepAddress, []common.Address{
		generateAddress(),
		generateAddress(),
		generateAddress(),
	})

	tlc.deposits[depositAddress] = &localDeposit{
		keepAddress: keepAddress.Hex(),
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
	tlc.mutex.Lock()
	defer tlc.mutex.Unlock()

	handlerID := generateHandlerID()

	tlc.depositCreatedHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		tlc.mutex.Lock()
		defer tlc.mutex.Unlock()

		delete(tlc.depositCreatedHandlers, handlerID)
	}), nil
}

func (tlc *TBTCLocalChain) OnDepositRegisteredPubkey(
	handler func(depositAddress string),
) (subscription.EventSubscription, error) {
	tlc.mutex.Lock()
	defer tlc.mutex.Unlock()

	handlerID := generateHandlerID()

	tlc.depositRegisteredPubkeyHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		tlc.mutex.Lock()
		defer tlc.mutex.Unlock()

		delete(tlc.depositRegisteredPubkeyHandlers, handlerID)
	}), nil
}

func (tlc *TBTCLocalChain) KeepAddress(depositAddress string) (string, error) {
	tlc.mutex.Lock()
	defer tlc.mutex.Unlock()

	deposit, ok := tlc.deposits[depositAddress]
	if !ok {
		return "", fmt.Errorf("no deposit with address [%v]", depositAddress)
	}

	return deposit.keepAddress, nil
}

func (tlc *TBTCLocalChain) RetrieveSignerPubkey(depositAddress string) error {
	tlc.mutex.Lock()
	defer tlc.mutex.Unlock()

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
	tlc.handlerMutex.Lock()
	defer tlc.handlerMutex.Unlock()

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

	for _, handler := range tlc.depositRegisteredPubkeyHandlers {
		go func(handler func(depositAddress string), depositAddress string) {
			handler(depositAddress)
		}(handler, depositAddress)
	}

	return nil
}

func (tlc *TBTCLocalChain) DepositPubkey(
	depositAddress string,
) ([]byte, error) {
	tlc.mutex.Lock()
	defer tlc.mutex.Unlock()

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

func (tlc *TBTCLocalChain) Logger() *localChainLogger {
	return tlc.logger
}
