package ethereum

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/ethereum"
	"github.com/keep-network/keep-tecdsa/pkg/eth/chain/gen/abi"
	"github.com/keep-network/keep-tecdsa/pkg/eth/event"

	configUtil "github.com/keep-network/keep-tecdsa/internal/config"
)

const configPath = "../configs/config.toml"

func TestConnect(t *testing.T) {
	cfg, err := configUtil.ReadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	config := &cfg.Ethereum

	chainAPI, err := ethereum.Connect(config)
	if err != nil {
		t.Fatal(err)
	}

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)

	var actualEvent *event.GroupRequested
	handle := func(event *event.GroupRequested) {
		actualEvent = event
		waitGroup.Done()
	}

	subscription, err := chainAPI.OnGroupRequested(handle)
	if err != nil {
		t.Fatal(err)
	}

	expectedGroupSize := uint32(10)
	expectedDishonestThreshold := uint32(5)

	testChain, err := Connect(config)
	if err != nil {
		t.Fatal(err)
	}

	err = requestGroup(testChain, expectedGroupSize, expectedDishonestThreshold)
	if err != nil {
		t.Fatal(err)
	}

	waitGroup.Wait()
	subscription.Unsubscribe()

	if actualEvent.GroupSize != expectedGroupSize {
		t.Errorf("invalid group size in event\nexpected: %v\nactual:   %v\n", expectedGroupSize, actualEvent.GroupSize)
	}
	if actualEvent.DishonestThreshold != expectedDishonestThreshold {
		t.Errorf("invalid dishonest threshold in event\nexpected: %v\nactual:   %v\n", expectedDishonestThreshold, actualEvent.DishonestThreshold)
	}
}

type testChain struct {
	client                  *ethclient.Client
	keepTECDSAGroupContract *abi.KeepTECDSAGroup
}

func Connect(config *ethereum.Config) (*testChain, error) {
	client, err := ethclient.Dial(config.URL)
	if err != nil {
		log.Fatal(err)
	}

	keepTECDSAGroupContractAddress, err := config.ContractAddress(ethereum.KeepTECDSAGroupContractName)
	if err != nil {
		return nil, err
	}

	keepTECDSAGroupContract, err := abi.NewKeepTECDSAGroup(
		keepTECDSAGroupContractAddress,
		client,
	)
	if err != nil {
		return nil, err
	}

	return &testChain{
		client:                  client,
		keepTECDSAGroupContract: keepTECDSAGroupContract,
	}, nil
}

func requestGroup(chain *testChain, expectedGroupSize, expectedDishonestThreshold uint32) error {
	privateKey, err := crypto.HexToECDSA("61c820b48fe56218ce8260b36e540d677dd0ebae54d906ea0f0baf64933bd810")
	if err != nil {
		return err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := chain.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	// gasPrice, err := client.SuggestGasPrice(context.Background())
	// if err != nil {
	// 			return err
	// }

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	// auth.GasPrice = gasPrice
	auth.GasPrice = big.NewInt(0)
	auth.From = fromAddress

	tx, err := chain.keepTECDSAGroupContract.RequestGroup(auth, expectedGroupSize, expectedDishonestThreshold)
	if err != nil {
		return err
	}

	log.Printf("requested group transaction: %v\n", tx.Hash().Hex())

	return nil
}
