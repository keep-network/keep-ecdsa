package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/ethereum"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/gen/abi"

	configUtil "github.com/keep-network/keep-tecdsa/internal/config"
)

const configPath = "../configs/config.toml"

func TestOnECDSAKeepRequestedPath(t *testing.T) {
	// TODO: Update this test to a smoke test executed with CLI command.
	t.Skip("skipping the test - it should be executed as integration smoke test")

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

	var actualEvent *eth.ECDSAKeepRequestedEvent
	handle := func(event *eth.ECDSAKeepRequestedEvent) {
		actualEvent = event
		waitGroup.Done()
	}

	subscription, err := chainAPI.OnECDSAKeepRequested(handle)
	if err != nil {
		t.Fatal(err)
	}

	expectedGroupSize := big.NewInt(10)
	expectedDishonestThreshold := big.NewInt(5)
	expectedOwnerAddress := common.HexToAddress("0x316F8eaf0b6065a53f0eaB3DD19aC6a07af95b3D")

	expectedKeepMembers := []*eth.MemberID{
		big.NewInt(1),
	}

	testChain, err := Connect(config)
	if err != nil {
		t.Fatal(err)
	}

	err = requestKeep(testChain, expectedGroupSize, expectedDishonestThreshold, expectedOwnerAddress)
	if err != nil {
		t.Fatal(err)
	}

	waitGroup.Wait()
	subscription.Unsubscribe()

	if !reflect.DeepEqual(expectedKeepMembers, actualEvent.MemberIDs) {
		t.Errorf("invalid members number\nexpected: %v\nactual:   %v\n", expectedKeepMembers, actualEvent.MemberIDs)
	}
	if actualEvent.DishonestThreshold.Cmp(expectedDishonestThreshold) != 0 {
		t.Errorf("invalid dishonest threshold in event\nexpected: %v\nactual:   %v\n", expectedDishonestThreshold, actualEvent.DishonestThreshold)
	}
}

type testChain struct {
	client                   *ethclient.Client
	ecdsaKeepFactoryContract *abi.ECDSAKeepFactory
}

func Connect(config *ethereum.Config) (*testChain, error) {
	client, err := ethclient.Dial(config.URL)
	if err != nil {
		log.Fatal(err)
	}

	ecdsaKeepFactoryContractAddress, err := config.ContractAddress(ethereum.ECDSAKeepFactoryContractName)
	if err != nil {
		return nil, err
	}
	ecdsaKeepFactoryContract, err := abi.NewECDSAKeepFactory(
		ecdsaKeepFactoryContractAddress,
		client,
	)
	if err != nil {
		return nil, err
	}

	return &testChain{
		client:                   client,
		ecdsaKeepFactoryContract: ecdsaKeepFactoryContract,
	}, nil
}

func requestKeep(chain *testChain, expectedGroupSize, expectedDishonestThreshold *big.Int, ownerAddress common.Address) error {
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

	tx, err := chain.ecdsaKeepFactoryContract.BuildNewKeep(auth, expectedGroupSize, expectedDishonestThreshold, ownerAddress)
	if err != nil {
		return fmt.Errorf("call to contract failed: [%s]", err)
	}

	log.Printf("requested group transaction: %v\n", tx.Hash().Hex())

	return nil
}
