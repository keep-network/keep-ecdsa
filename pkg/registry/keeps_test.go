package registry

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-tecdsa/internal/testdata"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss"
)

var (
	keepAddress1 = common.HexToAddress("0x770a9E2F2Aa1eC2d3Ca916Fc3e6A55058A898632")
	keepAddress2 = common.HexToAddress("0x8B3BccB3A3994681A1C1584DE4b4E8b23ed1Ed6d")

	groupMemberIDs = []tss.MemberID{
		tss.MemberID("member-1"),
		tss.MemberID("member-2"),
	}

	signer1 = newTestSigner(groupMemberIDs[0])
	signer2 = newTestSigner(groupMemberIDs[1])
)

func TestRegisterSigner(t *testing.T) {
	persistenceMock := &persistenceHandleMock{}
	gr := NewKeepsRegistry(persistenceMock)

	expectedSignerBytes, _ := signer1.Marshal()
	expectedFile := &testFileInfo{
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
		expectedFile,
		persistenceMock.persistedGroups[0],
	) {
		t.Errorf(
			"unexpected persisted group\nexpected: [%+v]\nactual:   [%+v]",
			expectedFile,
			persistenceMock.persistedGroups[0],
		)
	}
}

func TestGetGroup(t *testing.T) {
	persistenceMock := &persistenceHandleMock{}
	gr := NewKeepsRegistry(persistenceMock)

	gr.RegisterSigner(keepAddress1, signer1)

	var tests = map[string]struct {
		keepAddress    common.Address
		expectedSigner *tss.ThresholdSigner
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
					"unexpected signer\nexpected: [%+v]\nactual:   [%+v]",
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

func newTestSigner(memberID tss.MemberID) *tss.ThresholdSigner {
	testData, _ := testdata.LoadKeygenTestFixtures(1)

	return &tss.ThresholdSigner(
		tss.GroupInfo{
			GroupID:            "test-group-1",
			MemberID:           memberID,
			GroupMemberIDs:     groupMemberIDs,
			DishonestThreshold: 3,
		},
		testData[0],
	)
}
