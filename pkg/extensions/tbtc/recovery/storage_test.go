package recovery

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestDerivationIndexStorage_Read(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	dis, err := NewDerivationIndexStorage(dir)
	if err != nil {
		t.Fatal(err)
	}
	index, err := dis.GetNextIndex("extendedPublicKey")
	if err != nil {
		t.Fatal(err)
	}
	expectedIndex := 0
	if index != expectedIndex {
		t.Errorf("incorrect extendedPublicKey index\nexpected: %d\nactual:   %d", expectedIndex, index)
	}
}

func TestDerivationIndexStorage_WriteThenRead(t *testing.T) {
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
	expectedIndex := 7
	err = dis.Save(
		extendedPublicKey,
		expectedIndex-1,
		"1MjCqoLqMZ6Ru64TTtP16XnpSdiE8Kpgcx",
	)
	if err != nil {
		t.Fatal(err)
	}
	index, err := dis.GetNextIndex(extendedPublicKey)
	if err != nil {
		t.Fatal(err)
	}
	if index != expectedIndex {
		t.Errorf("incorrect extendedPublicKey index\nexpected: %d\nactual:   %d", expectedIndex, index)
	}
}
