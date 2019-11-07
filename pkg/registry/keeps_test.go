package registry

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-tecdsa/internal/testutils/mock"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

var (
	keepAddress1 = common.HexToAddress("0x770a9E2F2Aa1eC2d3Ca916Fc3e6A55058A898632")
	keepAddress2 = common.HexToAddress("0x8B3BccB3A3994681A1C1584DE4b4E8b23ed1Ed6d")

	signer1 = newTestSigner(big.NewInt(10))
	signer2 = newTestSigner(big.NewInt(20))
)

func TestRegisterSigner(t *testing.T) {
	persistenceMock := &mock.PersistenceHandle{}
	gr := NewKeepsRegistry(persistenceMock)

	expectedSignerBytes, _ := signer1.Marshal()
	expectedFile := &mock.TestFileInfo{
		Data:      expectedSignerBytes,
		Directory: keepAddress1.String(),
		Name:      "/signer_0",
	}

	gr.RegisterSigner(keepAddress1, signer1)

	// Verify persisted to storage.
	if len(persistenceMock.PersistedGroups) != 1 {
		t.Errorf(
			"unexpected number of persisted groups\nexpected: [%d]\nactual:   [%d]",
			1,
			len(persistenceMock.PersistedGroups),
		)
	}

	if !reflect.DeepEqual(
		expectedFile,
		persistenceMock.PersistedGroups[0],
	) {
		t.Errorf(
			"unexpected persisted group\nexpected: [%+v]\nactual:   [%+v]",
			expectedFile,
			persistenceMock.PersistedGroups[0],
		)
	}
}

func TestGetGroup(t *testing.T) {
	persistenceMock := &mock.PersistenceHandle{}
	gr := NewKeepsRegistry(persistenceMock)

	gr.RegisterSigner(keepAddress1, signer1)

	var tests = map[string]struct {
		keepAddress    common.Address
		expectedSigner *ecdsa.Signer
		expectedError  error
	}{
		"returns registered keep": {
			keepAddress:    keepAddress1,
			expectedSigner: signer1,
		},
		"returns error for not registered keep": {
			keepAddress:   keepAddress2,
			expectedError: fmt.Errorf("could not find signer: [%s]", keepAddress2.String()),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			signer, err := gr.GetSigner(test.keepAddress)

			if !reflect.DeepEqual(test.expectedSigner, signer) {
				t.Errorf(
					"unexpected group\nexpected: [%+v]\nactual:   [%+v]",
					test.expectedSigner,
					signer,
				)
			}

			if !reflect.DeepEqual(test.expectedError, err) {
				t.Errorf(
					"unexpected error\nexpected: [%v]\nactual:   [%v]",
					test.expectedError,
					err,
				)
			}
		})
	}
}

func TestRegisterNewGroupForTheSameKeep(t *testing.T) {
	persistenceMock := &mock.PersistenceHandle{}
	gr := NewKeepsRegistry(persistenceMock)

	gr.RegisterSigner(keepAddress1, signer1)
	gr.RegisterSigner(keepAddress1, signer2)

	signer, err := gr.GetSigner(keepAddress1)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(signer2, signer) {
		t.Errorf(
			"unexpected group\nexpected: [%+v]\nactual:   [%+v]",
			signer2,
			signer,
		)
	}

}

func TestLoadExistingGroups(t *testing.T) {
	persistenceMock := &mock.PersistenceHandle{}

	gr := NewKeepsRegistry(persistenceMock)

	if len(gr.GetKeepsAddresses()) != 0 {
		t.Fatal("unexpected keeps number at start")
	}

	gr.LoadExistingKeeps()

	signersCount := 0

	if len(gr.GetKeepsAddresses()) != 2 {
		t.Fatalf(
			"unexpected number of keeps\nexpected: [%d]\nactual:   [%d]",
			2,
			signersCount,
		)
	}

	actualSigner1, err := gr.GetSigner(keepAddress1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(signer1, actualSigner1) {
		t.Errorf("\nexpected: [%v]\nactual:   [%v]", signer1, actualSigner1)
	}

	actualSigner2, err := gr.GetSigner(keepAddress2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(signer2, actualSigner2) {
		t.Errorf("\nexpected: [%v]\nactual:   [%v]", signer2, actualSigner2)
	}
}
func newTestSigner(privateKeyD *big.Int) *ecdsa.Signer {
	curve := secp256k1.S256()

	privateKey := new(cecdsa.PrivateKey)
	privateKey.PublicKey.Curve = curve
	privateKey.D = privateKeyD
	privateKey.PublicKey.X, privateKey.PublicKey.Y = curve.ScalarBaseMult(privateKeyD.Bytes())

	return ecdsa.NewSigner(privateKey)
}
