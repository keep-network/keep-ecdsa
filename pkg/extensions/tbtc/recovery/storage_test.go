package recovery

import (
	"io/ioutil"
	"os"
	"testing"
)

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
	index, err := dis.GetNextIndex("ypub6Z7s8wJuKsxjd16oe85WH1uSbcbbCXuMFEhPMgcf7jQqNhQbT9jE52XVu1eBe18q2J3LwnDd54ufL2jNvidjfCkbd34aVwLtYdztLUqucwR")
	if err != nil {
		t.Fatal(err)
	}
	expectedIndex := 0
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
				err = dis.Save(
					input.publicKey,
					input.index,
					"<btc-address>",
				)
				if err != nil {
					t.Fatal(err)
				}
			}
			for _, expectation := range testData.expectations {
				index, err := dis.GetNextIndex(expectation.publicKey)
				if err != nil {
					t.Fatal(err)
				}

				if index != expectation.index {
					t.Errorf("incorrect extendedPublicKey index for %s\nexpected: %d\nactual:   %d", expectation.publicKey, expectation.index, index)
				}
			}
		})
	}
}
