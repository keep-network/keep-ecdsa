package registry

import (
	cecdsa "crypto/ecdsa"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

var (
	keepAddress1 = common.HexToAddress("0x770a9E2F2Aa1eC2d3Ca916Fc3e6A55058A898632")
	keepAddress2 = common.HexToAddress("0x8B3BccB3A3994681A1C1584DE4b4E8b23ed1Ed6d")

	signer1 = newTestSigner(big.NewInt(10))
	signer2 = newTestSigner(big.NewInt(20))
)

func TestRegisterSigner(t *testing.T) {
	persistenceMock := &persistenceHandleMock{}
	gr := NewKeepsRegistry(persistenceMock)

	expectedSignerBytes, _ := signer1.Marshal()
	expectedPersistedSigner := &testFileInfo{
		data:      expectedSignerBytes,
		directory: keepAddress1.String(),
		name:      "/signer_0",
	}

	gr.RegisterSigner(keepAddress1, signer1)

	// Verify persisted to storage.
	if len(persistenceMock.persistedGroups) != 1 {
		t.Errorf(
			"unexpected number of persisted groups\nexpected: [%d]\nactual:   [%d]",
			1,
			len(persistenceMock.persistedGroups),
		)
	}

	if !reflect.DeepEqual(
		expectedPersistedSigner,
		persistenceMock.persistedGroups[0],
	) {
		t.Errorf(
			"unexpected persisted group\nexpected: [%+v]\nactual:   [%+v]",
			expectedPersistedSigner,
			persistenceMock.persistedGroups[0],
		)
	}

	// Verify stored in a map.
	signer, ok := gr.myKeeps.Load(keepAddress1.String())
	if !ok {
		t.Errorf("failed to load signer")
	}

	if !reflect.DeepEqual(signer1, signer) {
		t.Errorf(
			"unexpected signer\nexpected: [%+v]\nactual:   [%+v]",
			signer1,
			signer,
		)
	}
}

func TestGetGroup(t *testing.T) {
	persistenceMock := &persistenceHandleMock{}
	gr := NewKeepsRegistry(persistenceMock)

	gr.RegisterSigner(keepAddress1, signer1)

	var tests = map[string]struct {
		keepAddress    common.Address
		expectedSigner *ecdsa.Signer
		expectedError  error
	}{
		"returns group for registered keep": {
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
	persistenceMock := &persistenceHandleMock{}
	gr := NewKeepsRegistry(persistenceMock)

	gr.RegisterSigner(keepAddress1, signer1)
	gr.RegisterSigner(keepAddress1, signer2)

	signer, err := gr.GetSigner(keepAddress1)

	if !reflect.DeepEqual(signer2, signer) {
		t.Errorf(
			"unexpected group\nexpected: [%+v]\nactual:   [%+v]",
			signer2,
			signer,
		)
	}

	if err != nil {
		t.Errorf("unexpected error: [%v]", err)
	}
}

func TestLoadExistingGroups(t *testing.T) {
	persistenceMock := &persistenceHandleMock{}

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
		t.Fatalf("unexpected error: [%v]", err)
	}
	if !reflect.DeepEqual(signer1, actualSigner1) {
		t.Errorf("\nexpected: [%v]\nactual:   [%v]", signer1, actualSigner1)
	}

	actualSigner2, err := gr.GetSigner(keepAddress2)
	if err != nil {
		t.Fatalf("unexpected error: [%v]", err)
	}
	if !reflect.DeepEqual(signer2, actualSigner2) {
		t.Errorf("\nexpected: [%v]\nactual:   [%v]", signer2, actualSigner2)
	}
}

type persistenceHandleMock struct {
	persistedGroups []*testFileInfo
	archivedGroups  []string
}

type testFileInfo struct {
	data      []byte
	directory string
	name      string
}

func (phm *persistenceHandleMock) Save(data []byte, directory string, name string) error {
	phm.persistedGroups = append(
		phm.persistedGroups,
		&testFileInfo{data, directory, name},
	)
	return nil
}

func (phm *persistenceHandleMock) ReadAll() (<-chan persistence.DataDescriptor, <-chan error) {
	signerBytes1, _ := signer1.Marshal()
	signerBytes2, _ := signer2.Marshal()

	outputData := make(chan persistence.DataDescriptor, 2)
	outputErrors := make(chan error)

	outputData <- &testDataDescriptor{"/membership_0", keepAddress1.String(), signerBytes1}
	outputData <- &testDataDescriptor{"/membership_0", keepAddress2.String(), signerBytes2}

	close(outputData)
	close(outputErrors)

	return outputData, outputErrors
}

func (phm *persistenceHandleMock) Archive(directory string) error {
	phm.archivedGroups = append(phm.archivedGroups, directory)

	return nil
}

type testDataDescriptor struct {
	name      string
	directory string
	content   []byte
}

func (tdd *testDataDescriptor) Name() string {
	return tdd.name
}

func (tdd *testDataDescriptor) Directory() string {
	return tdd.directory
}

func (tdd *testDataDescriptor) Content() ([]byte, error) {
	return tdd.content, nil
}

func newTestSigner(privateKeyD *big.Int) *ecdsa.Signer {
	curve := secp256k1.S256()

	privateKey := new(cecdsa.PrivateKey)
	privateKey.PublicKey.Curve = curve
	privateKey.D = privateKeyD
	privateKey.PublicKey.X, privateKey.PublicKey.Y = curve.ScalarBaseMult(privateKeyD.Bytes())

	return ecdsa.NewSigner(privateKey)
}
