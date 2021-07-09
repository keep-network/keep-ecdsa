package recovery

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/keep-network/keep-ecdsa/pkg/chain/bitcoin"
)

type mockBitcoinHandle struct {
	broadcast           func(transaction string) error
	vbyteFeeFor25Blocks func() (int32, error)
	isAddressUnused     func(btcAddress string) (bool, error)
}

func newMockBitcoinHandle() *mockBitcoinHandle {
	return &mockBitcoinHandle{
		broadcast:           func(_ string) error { return nil },
		vbyteFeeFor25Blocks: func() (int32, error) { return 75, nil },
		isAddressUnused:     func(_ string) (bool, error) { return true, nil },
	}
}
func (mbh mockBitcoinHandle) Broadcast(transaction string) error {
	return mbh.broadcast(transaction)
}
func (mbh mockBitcoinHandle) VbyteFeeFor25Blocks() (int32, error) {
	return mbh.vbyteFeeFor25Blocks()
}
func (mbh mockBitcoinHandle) IsAddressUnused(btcAddress string) (bool, error) {
	return mbh.isAddressUnused(btcAddress)
}

func TestDerivationIndexStorage_GetNextAddressOnNewKey(t *testing.T) {
	chainParams := &chaincfg.MainNetParams

	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	dis, err := NewDerivationIndexStorage(dir)
	if err != nil {
		t.Fatal(err)
	}
	extendedPublicKey := "ypub6Z7s8wJuKsxjd16oe85WH1uSbcbbCXuMFEhPMgcf7jQqNhQbT9jE52XVu1eBe18q2J3LwnDd54ufL2jNvidjfCkbd34aVwLtYdztLUqucwR"
	for i := uint32(0); i < 10; i++ {
		btcAddress, err := dis.GetNextAddress(extendedPublicKey, newMockBitcoinHandle(), chainParams, false)
		if err != nil {
			t.Fatal(err)
		}

		expectedBtcAddress, err := bitcoin.DeriveAddress(extendedPublicKey, i, chainParams)
		if err != nil {
			t.Fatal(err)
		}
		if btcAddress != expectedBtcAddress {
			t.Errorf("incorrect derived address for call # %d\nexpected: %s\nactual:   %s", i, expectedBtcAddress, btcAddress)
		}
	}
}

type keyAndIndex struct {
	publicKey string
	index     int
}

type keyAndAddress struct {
	publicKey  string
	btcAddress string
	index      int
}

func TestDerivationIndexStorage_SaveThenGetNextAddress(t *testing.T) {
	testData := map[string]struct {
		inputs       []keyAndIndex
		expectations []keyAndAddress
	}{
		"single key, single entry": {
			[]keyAndIndex{{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 5}},
			[]keyAndAddress{{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", "1QETuEAw5UBdYtz6vJw8L9582TdrrE4b3B", 6}},
		},
		"multiple keys, single entry": {
			[]keyAndIndex{
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 5},
				{"ypub6ZpieGfpesfH3KqGr4zZPETidCze6RzeNMz7FLnSPgABwyQNZZmpA4tpUYFn53xtHkHXaoGviseJJcFhSn3Kw9sgzsiSnP5xEqp6Z2Yy4ZH", 48},
				{"zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9", 112},
			},
			[]keyAndAddress{
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", "1QETuEAw5UBdYtz6vJw8L9582TdrrE4b3B", 6},
				{"ypub6ZpieGfpesfH3KqGr4zZPETidCze6RzeNMz7FLnSPgABwyQNZZmpA4tpUYFn53xtHkHXaoGviseJJcFhSn3Kw9sgzsiSnP5xEqp6Z2Yy4ZH", "3BRGrKZzkuuaqVGK5eZkcA5wrzeQULawMH", 49},
				{"zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9", "bc1qcd39cwrsefagqh4y277q0rgm0stdsth4xr6mjr", 113},
			},
		},
		"single key, multiple entries": {
			[]keyAndIndex{
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 5},
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 172},
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 39},
			},
			[]keyAndAddress{
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", "1Je1vYfst9yGF95KkYitQ7QhdLUkNVCzfX", 173},
			},
		},
		"multiple keys, multiple entries": {
			[]keyAndIndex{
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 513},
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 5090},
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 3544},

				{"ypub6ZpieGfpesfH3KqGr4zZPETidCze6RzeNMz7FLnSPgABwyQNZZmpA4tpUYFn53xtHkHXaoGviseJJcFhSn3Kw9sgzsiSnP5xEqp6Z2Yy4ZH", 1692},
				{"ypub6ZpieGfpesfH3KqGr4zZPETidCze6RzeNMz7FLnSPgABwyQNZZmpA4tpUYFn53xtHkHXaoGviseJJcFhSn3Kw9sgzsiSnP5xEqp6Z2Yy4ZH", 223},
				{"ypub6ZpieGfpesfH3KqGr4zZPETidCze6RzeNMz7FLnSPgABwyQNZZmpA4tpUYFn53xtHkHXaoGviseJJcFhSn3Kw9sgzsiSnP5xEqp6Z2Yy4ZH", 8982},

				{"zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9", 6311},
				{"zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9", 6999},
				{"zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9", 8559},
			},
			[]keyAndAddress{
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", "1Cck1ps6NGB9LGjrymNS21KC7fJyU4X3fw", 5091},
				{"ypub6ZpieGfpesfH3KqGr4zZPETidCze6RzeNMz7FLnSPgABwyQNZZmpA4tpUYFn53xtHkHXaoGviseJJcFhSn3Kw9sgzsiSnP5xEqp6Z2Yy4ZH", "34c8quMWCqNsfVFhgTveC3s8kyTWf9m5t8", 8983},
				{"zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9", "bc1qdvy9xuq2368ywuvgfg77sz688x9v0fjg6f0gw8", 8560},
			},
		},
		"trim whitespaces": {
			[]keyAndIndex{
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 513},
				{"    xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1    ", 5090},
			},
			[]keyAndAddress{
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", "1Cck1ps6NGB9LGjrymNS21KC7fJyU4X3fw", 5091},
				{"       xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1          ", "1GTZk2iwgiDvTpgQ2XK6N4L7Nr98AcbjG6", 5092},
			},
		},
		"write to the same index multiple times": {
			[]keyAndIndex{
				{"zpub6p6mUAk2dpLVVsguhHA27Qgd8e4q394Csha9jfAJCABrFRcnSv6AYPsgmAzgR8feBR4Spu2piv4xMcneocZajEvKtoHp111pizpe6aAEqfp", 777},
				{"zpub6p6mUAk2dpLVVsguhHA27Qgd8e4q394Csha9jfAJCABrFRcnSv6AYPsgmAzgR8feBR4Spu2piv4xMcneocZajEvKtoHp111pizpe6aAEqfp", 777},
				{"zpub6p6mUAk2dpLVVsguhHA27Qgd8e4q394Csha9jfAJCABrFRcnSv6AYPsgmAzgR8feBR4Spu2piv4xMcneocZajEvKtoHp111pizpe6aAEqfp", 777},
				{"zpub6p6mUAk2dpLVVsguhHA27Qgd8e4q394Csha9jfAJCABrFRcnSv6AYPsgmAzgR8feBR4Spu2piv4xMcneocZajEvKtoHp111pizpe6aAEqfp", 777},
			},
			[]keyAndAddress{
				{"zpub6p6mUAk2dpLVVsguhHA27Qgd8e4q394Csha9jfAJCABrFRcnSv6AYPsgmAzgR8feBR4Spu2piv4xMcneocZajEvKtoHp111pizpe6aAEqfp", "bc1qj98w6u98t6t0pwew4fvlxcmcevhqkznp2qjx2f", 778},
			},
		},
	}

	for testName, testData := range testData {
		t.Run(testName, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "example")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)
			dis, err := NewDerivationIndexStorage(dir)

			if err != nil {
				t.Fatal(err)
			}
			for _, input := range testData.inputs {
				err = dis.save(
					input.publicKey,
					uint32(input.index),
				)
				if err != nil {
					t.Fatal(err)
				}
			}
			for _, expectation := range testData.expectations {
				address, err := dis.GetNextAddress(expectation.publicKey, newMockBitcoinHandle(), &chaincfg.MainNetParams, false)
				if err != nil {
					t.Fatal(err)
				}

				if address != expectation.btcAddress {
					t.Errorf("incorrect derived address for %s\nexpected: %s\nactual:   %s", expectation.publicKey, expectation.btcAddress, address)
				}

				storedIndex, err := dis.read(expectation.publicKey)
				if err != nil {
					t.Fatalf("failed to read last used index: %s", err)
				}
				if storedIndex != int(expectation.index) {
					t.Errorf(
						"the resolved index does not match\nexpected: %d\nactual:   %d",
						expectation.index,
						storedIndex,
					)
				}
			}
		})
	}
}

func TestDerivationIndexStorage_GetNextAddressDryRun(t *testing.T) {
	publicKey := "xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1"
	usedIndex := 5
	expectedBtcAddress := "1QETuEAw5UBdYtz6vJw8L9582TdrrE4b3B"
	isDryRun := true

	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	dis, err := NewDerivationIndexStorage(dir)

	if err != nil {
		t.Fatal(err)
	}
	err = dis.save(publicKey, uint32(usedIndex))
	if err != nil {
		t.Fatal(err)
	}

	address, err := dis.GetNextAddress(publicKey, newMockBitcoinHandle(), &chaincfg.MainNetParams, isDryRun)
	if err != nil {
		t.Fatal(err)
	}

	if address != expectedBtcAddress {
		t.Errorf("incorrect derived address\nexpected: %s\nactual:   %s", expectedBtcAddress, address)
	}

	storedIndex, err := dis.read(publicKey)
	if err != nil {
		t.Fatalf("failed to read last used index: %s", err)
	}
	if storedIndex != usedIndex {
		t.Errorf(
			"the resolved index does not match\nexpected: %d\nactual:   %d",
			usedIndex,
			storedIndex,
		)
	}
}

func TestDerivationIndexStorage_OverwriteExistingPair(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	dis, err := NewDerivationIndexStorage(dir)
	if err != nil {
		t.Fatal(err)
	}
	extendedPublicKey := "xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1"
	index := uint32(89)
	err = dis.save(extendedPublicKey, index)
	if err != nil {
		t.Fatal(err)
	}
	err = dis.save(extendedPublicKey, index)
	if err != nil {
		t.Errorf("unexpected error trying to overwrite extendedPublicKey [%s] at index [%d]: [%v]", extendedPublicKey, index, err)
	}
}

func TestDerivationIndexStorage_ShortExtendedPublicKeys(t *testing.T) {
	null := "\xff" // represents no error
	testData := map[string]struct {
		input         keyAndIndex
		expectedError string
	}{
		"6-letter key":  {keyAndIndex{"abc123", 8}, "insufficient length for public key"},
		"11-letter key": {keyAndIndex{"1111.1111.1", 12}, "insufficient length for public key"},
		"12-letter key": {keyAndIndex{"1111.1111.11", 16}, null},
		"13-letter key": {keyAndIndex{"1111.1111.111", 20}, null},
	}
	for testName, testData := range testData {
		t.Run(testName, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "example")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)
			dis, err := NewDerivationIndexStorage(dir)
			if err != nil {
				t.Fatal(err)
			}

			err = dis.save(testData.input.publicKey, uint32(testData.input.index))
			if testData.expectedError == null {
				if err != nil {
					t.Errorf("unexpected error: [%v]", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected an error, but found none")
				}
				if !ErrorContains(err, testData.expectedError) {
					t.Errorf("unexpected error\nexpected: %s\nactual:   %v", testData.expectedError, err)
				}
			}
		})
	}
}

func TestDerivationIndexStorage_NewDerivationIndexStorageOnNonexistantPath(t *testing.T) {
	_, err := NewDerivationIndexStorage("banana-fofana")
	if !ErrorContains(err, "no such file or directory") {
		t.Errorf("unexpected error: [%v]", err)
	}
}

func TestDerivationIndexStorage_BadPermissions(t *testing.T) {
	null := "\xff" // represents no error
	testData := map[string]struct {
		mode int
		err  string
	}{
		"execute only":               {0100, "cannot read from the storage directory"},
		"write only":                 {0200, "cannot read from the storage directory"},
		"write and execute":          {0300, "cannot read from the storage directory"},
		"read only":                  {0400, "cannot write to the storage directory"},
		"read and execute":           {0500, "cannot write to the storage directory"},
		"read and write":             {0600, "cannot write to the storage directory"},
		"read and write and execute": {0700, null},
	}
	for testName, testData := range testData {
		t.Run(testName, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "example")
			if err != nil {
				t.Fatal(err)
			}
			err = os.Chmod(dir, os.FileMode(testData.mode))
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)
			_, err = NewDerivationIndexStorage(dir)
			if testData.err == null {
				if err != nil {
					t.Errorf("unexpected error\nexpected: <nil>\nactual:   %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("unexpected error\nexpected: %s\nactual:   <nil>", testData.err)
				} else if !ErrorContains(err, testData.err) {
					t.Errorf("unexpected error\nexpected: %s\nactual:   %v", testData.err, err)
				}
			}
		})
	}
}

func TestDerivationIndexStorage_MultipleAsyncGetNextAddresses(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	dis, err := NewDerivationIndexStorage(dir)
	if err != nil {
		t.Fatal(err)
	}
	extendedPublicKey := "xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1"
	index := uint32(831)
	err = dis.save(extendedPublicKey, index)
	if err != nil {
		t.Fatal(err)
	}

	chainParams := &chaincfg.MainNetParams

	type pair struct {
		address string
		err     error
	}
	iterations := 10
	getNextAddressResults := make(chan pair, iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			//only validate multiples of 10 to test that each concurrent hit respects the previous call's validation
			handle := newMockBitcoinHandle()
			validIndexes := make(map[string]bool)
			for i := 0; i < iterations; i++ {
				// the valid indexes should be 840, 850, 860, 870...
				index := uint32(840) + 10*uint32(i)
				derivedAddress, err := bitcoin.DeriveAddress(extendedPublicKey, index, chainParams)
				if err != nil {
					getNextAddressResults <- pair{"", err}
					return
				}
				validIndexes[derivedAddress] = true
			}
			handle.isAddressUnused = func(btcAddress string) (bool, error) {
				return validIndexes[btcAddress], nil
			}
			address, err := dis.GetNextAddress(extendedPublicKey, handle, chainParams, false)
			getNextAddressResults <- pair{address, err}
		}()
	}
	for i := 0; i < iterations; i++ {
		result := <-getNextAddressResults
		if result.err != nil {
			t.Fatal(err)
		}
		// the valid indexes should be 840, 850, 860, 870...
		expectedIndex := uint32(840) + 10*uint32(i)
		expectedAddress, err := bitcoin.DeriveAddress(extendedPublicKey, expectedIndex, chainParams)
		if err != nil {
			t.Fatal(err)
		}
		if result.address != expectedAddress {
			t.Errorf("unexpected address\nexpected: %s\nactual:   %s", expectedAddress, result.address)
		}
	}
}

func TestDervationIndexStorage_SaveOverwrites(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	dis, err := NewDerivationIndexStorage(dir)
	if err != nil {
		t.Fatal(err)
	}
	extendedPublicKey := "xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1"
	index := uint32(831)
	iterations := 10
	for i := 0; i < iterations; i++ {
		err = dis.save(extendedPublicKey, index)
		if err != nil {
			t.Fatal(err)
		}
	}
}
