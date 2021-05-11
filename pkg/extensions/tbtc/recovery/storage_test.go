package recovery

import (
	"io/ioutil"
	"os"
	"testing"
)

type mockBitcoinHandle struct {
	broadcast       func(transaction string) error
	vbyteFee        func() (int32, error)
	isAddressUnused func(btcAddress string) (bool, error)
}

func newMockBitcoinHandle() *mockBitcoinHandle {
	return &mockBitcoinHandle{
		broadcast:       func(_ string) error { return nil },
		vbyteFee:        func() (int32, error) { return 75, nil },
		isAddressUnused: func(_ string) (bool, error) { return true, nil },
	}
}
func (mbh mockBitcoinHandle) Broadcast(transaction string) error {
	return mbh.broadcast(transaction)
}
func (mbh mockBitcoinHandle) VbyteFee() (int32, error) {
	return mbh.vbyteFee()
}
func (mbh mockBitcoinHandle) IsAddressUnused(btcAddress string) (bool, error) {
	return mbh.isAddressUnused(btcAddress)
}

func TestDerivationIndexStorage_GetNextIndexOnNewKey(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	dis, err := NewDerivationIndexStorage(dir)
	if err != nil {
		t.Fatal(err)
	}
	index, err := dis.GetNextIndex("ypub6Z7s8wJuKsxjd16oe85WH1uSbcbbCXuMFEhPMgcf7jQqNhQbT9jE52XVu1eBe18q2J3LwnDd54ufL2jNvidjfCkbd34aVwLtYdztLUqucwR", newMockBitcoinHandle())
	if err != nil {
		t.Fatal(err)
	}
	expectedIndex := uint32(0)
	if index != expectedIndex {
		t.Errorf("incorrect extendedPublicKey index\nexpected: %d\nactual:   %d", expectedIndex, index)
	}
}

type keyAndIndex struct {
	publicKey string
	index     int
}

func TestDerivationIndexStorage_SaveThenGetNextIndex(t *testing.T) {
	testData := map[string]struct {
		inputs       []keyAndIndex
		expectations []keyAndIndex
	}{
		"single key, single entry": {
			[]keyAndIndex{{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 5}},
			[]keyAndIndex{{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 6}},
		},
		"multiple keys, single entry": {
			[]keyAndIndex{
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 5},
				{"ypub6ZpieGfpesfH3KqGr4zZPETidCze6RzeNMz7FLnSPgABwyQNZZmpA4tpUYFn53xtHkHXaoGviseJJcFhSn3Kw9sgzsiSnP5xEqp6Z2Yy4ZH", 48},
				{"zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9", 112},
			},
			[]keyAndIndex{
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 6},
				{"ypub6ZpieGfpesfH3KqGr4zZPETidCze6RzeNMz7FLnSPgABwyQNZZmpA4tpUYFn53xtHkHXaoGviseJJcFhSn3Kw9sgzsiSnP5xEqp6Z2Yy4ZH", 49},
				{"zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9", 113},
			},
		},
		"single key, multiple entries": {
			[]keyAndIndex{
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 5},
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 172},
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 39},
			},
			[]keyAndIndex{
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 173},
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
			[]keyAndIndex{
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 5091},
				{"ypub6ZpieGfpesfH3KqGr4zZPETidCze6RzeNMz7FLnSPgABwyQNZZmpA4tpUYFn53xtHkHXaoGviseJJcFhSn3Kw9sgzsiSnP5xEqp6Z2Yy4ZH", 8983},
				{"zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9", 8560},
			},
		},
		"trim whitespaces": {
			[]keyAndIndex{
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 513},
				{"    xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1    ", 5090},
			},
			[]keyAndIndex{
				{"xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1", 5091},
				{"       xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1          ", 5092},
			},
		},
		"write to the same index multiple times": {
			[]keyAndIndex{
				{"zpub6p6mUAk2dpLVVsguhHA27Qgd8e4q394Csha9jfAJCABrFRcnSv6AYPsgmAzgR8feBR4Spu2piv4xMcneocZajEvKtoHp111pizpe6aAEqfp", 777},
				{"zpub6p6mUAk2dpLVVsguhHA27Qgd8e4q394Csha9jfAJCABrFRcnSv6AYPsgmAzgR8feBR4Spu2piv4xMcneocZajEvKtoHp111pizpe6aAEqfp", 777},
				{"zpub6p6mUAk2dpLVVsguhHA27Qgd8e4q394Csha9jfAJCABrFRcnSv6AYPsgmAzgR8feBR4Spu2piv4xMcneocZajEvKtoHp111pizpe6aAEqfp", 777},
				{"zpub6p6mUAk2dpLVVsguhHA27Qgd8e4q394Csha9jfAJCABrFRcnSv6AYPsgmAzgR8feBR4Spu2piv4xMcneocZajEvKtoHp111pizpe6aAEqfp", 777},
			},
			[]keyAndIndex{
				{"zpub6p6mUAk2dpLVVsguhHA27Qgd8e4q394Csha9jfAJCABrFRcnSv6AYPsgmAzgR8feBR4Spu2piv4xMcneocZajEvKtoHp111pizpe6aAEqfp", 778},
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
				index, err := dis.GetNextIndex(expectation.publicKey, newMockBitcoinHandle())
				if err != nil {
					t.Fatal(err)
				}

				if index != uint32(expectation.index) {
					t.Errorf("incorrect extendedPublicKey index for %s\nexpected: %d\nactual:   %d", expectation.publicKey, expectation.index, index)
				}
			}
		})
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

func TestDerviationIndexStorage_BadPermissions(t *testing.T) {
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

func TestDerivationIndexStorage_MultipleAsyncGetNextIndexes(t *testing.T) {
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

	type pair struct {
		index uint32
		err   error
	}
	iterations := 10
	getNextIndexResults := make(chan pair, iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			//only validate multiples of 10 to test that each concurrent hit respects the previous call's validation
			handle := newMockBitcoinHandle()
			validIndexes := make(map[string]bool)
			for i := 0; i < iterations; i++ {
				// the valid indexes should be 840, 850, 860, 870...
				index := uint32(840) + 10*uint32(i)
				derivedAddress, err := deriveAddress(extendedPublicKey, index)
				if err != nil {
					getNextIndexResults <- pair{0, err}
					return
				}
				validIndexes[derivedAddress] = true
			}
			handle.isAddressUnused = func(btcAddress string) (bool, error) {
				return validIndexes[btcAddress], nil
			}
			nextIndex, err := dis.GetNextIndex(extendedPublicKey, handle)
			getNextIndexResults <- pair{nextIndex, err}
		}()
	}
	for i := 0; i < iterations; i++ {
		result := <-getNextIndexResults
		if result.err != nil {
			t.Fatal(err)
		}
		// the valid indexes should be 840, 850, 860, 870...
		expectedIndex := uint32(840) + 10*uint32(i)
		if result.index != expectedIndex {
			t.Errorf("unexpected next index\nexpected: %d\nactual:   %d", expectedIndex, result.index)
		}
	}
}

func TestDerviationIndexStorage_SaveOverwrites(t *testing.T) {
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