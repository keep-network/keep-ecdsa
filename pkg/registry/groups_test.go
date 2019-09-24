package registry

import (
	cecdsa "crypto/ecdsa"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/keep-network/keep-core/pkg/persistence"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
)

var (
	keepAddress1 = common.HexToAddress("0x770a9E2F2Aa1eC2d3Ca916Fc3e6A55058A898632")
	keepAddress2 = common.HexToAddress("0x8B3BccB3A3994681A1C1584DE4b4E8b23ed1Ed6d")

	signer1 = newTestSigner(big.NewInt(10))
	signer2 = newTestSigner(big.NewInt(20))
)

func TestRegisterGroup(t *testing.T) {
	persistenceMock := &persistenceHandleMock{}
	gr := NewGroupRegistry(persistenceMock)

	expectedGroup := &Membership{keepAddress1, signer1}
	expectedGroupBytes, _ := expectedGroup.Marshal()
	expectedPersistedGroup := &storageMock{
		data:      expectedGroupBytes,
		directory: keepAddress1.String(),
		name:      "/membership_0",
	}

	gr.RegisterGroup(keepAddress1, signer1)

	// Verify persisted to storage.
	if len(persistenceMock.persistedGroups) != 1 {
		t.Errorf(
			"unexpected number of persisted groups\nexpected: [%d]\nactual:   [%d]",
			1,
			len(persistenceMock.persistedGroups),
		)
	}

	if !reflect.DeepEqual(
		expectedPersistedGroup,
		persistenceMock.persistedGroups[0],
	) {
		t.Errorf(
			"unexpected persisted group\nexpected: [%+v]\nactual:   [%+v]",
			expectedPersistedGroup,
			persistenceMock.persistedGroups[0],
		)
	}

	// Verify stored in a map.
	group, ok := gr.myGroups.Load(keepAddress1.String())
	if !ok {
		t.Errorf("failed to load group")
	}

	if !reflect.DeepEqual(expectedGroup, group) {
		t.Errorf(
			"unexpected group\nexpected: [%+v]\nactual:   [%+v]",
			expectedGroup,
			group,
		)
	}
}

func TestGetGroup(t *testing.T) {
	persistenceMock := &persistenceHandleMock{}
	gr := NewGroupRegistry(persistenceMock)

	gr.myGroups.Store(keepAddress1.String(), &Membership{keepAddress1, signer1})

	var tests = map[string]struct {
		keepAddress   common.Address
		expectedGroup *Membership
		expectedError error
	}{
		"returns group for registered keep": {
			keepAddress:   keepAddress1,
			expectedGroup: &Membership{keepAddress1, signer1},
		},
		"returns error for not registered keep": {
			keepAddress:   keepAddress2,
			expectedError: fmt.Errorf("failed to find signer for keep: [%s]", keepAddress2.String()),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			group, err := gr.GetGroup(test.keepAddress)

			if !reflect.DeepEqual(test.expectedGroup, group) {
				t.Errorf(
					"unexpected group\nexpected: [%+v]\nactual:   [%+v]",
					test.expectedGroup,
					group,
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
	gr := NewGroupRegistry(persistenceMock)

	expectedGroup := &Membership{keepAddress1, signer2}

	gr.RegisterGroup(keepAddress1, signer1)
	gr.RegisterGroup(keepAddress1, signer2)

	group, err := gr.GetGroup(keepAddress1)

	if !reflect.DeepEqual(expectedGroup, group) {
		t.Errorf(
			"unexpected group\nexpected: [%+v]\nactual:   [%+v]",
			expectedGroup,
			group,
		)
	}

	if err != nil {
		t.Errorf("unexpected error: [%v]", err)
	}
}
func TestLoadExistingGroups(t *testing.T) {
	persistenceMock := &persistenceHandleMock{}

	gr := NewGroupRegistry(persistenceMock)

	expectedMembership1 := &Membership{keepAddress1, signer1}
	expectedMembership2 := &Membership{keepAddress2, signer2}

	gr.ForEachGroup(func(keepAddress common.Address, membership *Membership) bool {
		t.Fatal("unexpected group membership at start")
		return false
	})

	gr.LoadExistingGroups()

	groupsCount := 0

	gr.ForEachGroup(func(keepAddress common.Address, membership *Membership) bool {
		groupsCount++
		return true
	})

	if groupsCount != 2 {
		t.Fatalf(
			"unexpected number of group memberships\nexpected: [%d]\nactual:   [%d]",
			2,
			groupsCount,
		)
	}

	actualMembership1, err := gr.GetGroup(keepAddress1)
	if err != nil {
		t.Fatalf("unexpected error: [%v]", err)
	}
	if !reflect.DeepEqual(expectedMembership1, actualMembership1) {
		t.Errorf("\nexpected: [%v]\nactual:   [%v]", expectedMembership1, actualMembership1)
	}

	actualMembership2, err := gr.GetGroup(keepAddress2)
	if err != nil {
		t.Fatalf("unexpected error: [%v]", err)
	}
	if !reflect.DeepEqual(expectedMembership2, actualMembership2) {
		t.Errorf("\nexpected: [%v]\nactual:   [%v]", expectedMembership2, actualMembership2)
	}
}

type persistenceHandleMock struct {
	persistedGroups []*storageMock
	archivedGroups  []string
}

type storageMock struct {
	data      []byte
	directory string
	name      string
}

func (phm *persistenceHandleMock) Save(data []byte, directory string, name string) error {
	phm.persistedGroups = append(
		phm.persistedGroups,
		&storageMock{data, directory, name},
	)
	return nil
}

func (phm *persistenceHandleMock) ReadAll() (<-chan persistence.DataDescriptor, <-chan error) {
	membershipBytes1, _ := (&Membership{
		KeepAddress: keepAddress1,
		Signer:      signer1,
	}).Marshal()

	membershipBytes2, _ := (&Membership{
		KeepAddress: keepAddress2,
		Signer:      signer2,
	}).Marshal()

	outputData := make(chan persistence.DataDescriptor, 2)
	outputErrors := make(chan error)

	outputData <- &testDataDescriptor{"1", "dir", membershipBytes1}
	outputData <- &testDataDescriptor{"2", "dir", membershipBytes2}

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
