package registry

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"

	"github.com/keep-network/keep-common/pkg/persistence"
	testdata "github.com/keep-network/keep-tecdsa/internal/testdata/tss"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss/gen/pb"
)

var (
	keepAddress1 = common.HexToAddress("0x770a9E2F2Aa1eC2d3Ca916Fc3e6A55058A898632")
	keepAddress2 = common.HexToAddress("0x8B3BccB3A3994681A1C1584DE4b4E8b23ed1Ed6d")
	keepAddress3 = common.HexToAddress("0x0472ec0185ebb8202f3d4ddb0226998889663cf2")

	groupMemberIDs = [][]byte{
		[]byte("member-1"),
		[]byte("member-2"),
		[]byte("member-3"),
	}
)

func TestRegisterSigner(t *testing.T) {
	persistenceMock := &persistenceHandleMock{}
	gr := NewKeepsRegistry(persistenceMock)

	signer1, err := newTestSigner(0)
	if err != nil {
		t.Fatalf("failed to get signer: [%v]", err)
	}

	expectedSignerBytes, err := signer1.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal signer: [%v]", err)
	}

	expectedFile := &testFileInfo{
		data:      expectedSignerBytes,
		directory: keepAddress1.String(),
		name:      fmt.Sprintf("/membership_%s", signer1.MemberID().String()),
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

	signers, err := testSigners()
	if err != nil {
		t.Fatalf("failed to get signer: [%v]", err)
	}

	signer1 := signers[0]
	signer2 := signers[1]
	signer3 := signers[2]

	gr.RegisterSigner(keepAddress1, signer1)
	gr.RegisterSigner(keepAddress2, signer2)
	gr.RegisterSigner(keepAddress2, signer3)

	var tests = map[string]struct {
		keepAddress    common.Address
		expectedSigner []*tss.ThresholdSigner
		expectedError  error
	}{
		"returns registered keep with one signer": {
			keepAddress:    keepAddress1,
			expectedSigner: []*tss.ThresholdSigner{signer1},
		},
		"returns registered keep with multiple signers": {
			keepAddress:    keepAddress2,
			expectedSigner: []*tss.ThresholdSigner{signer2, signer3},
		},
		"returns error for not registered keep": {
			keepAddress:   keepAddress3,
			expectedError: fmt.Errorf("could not find signers for keep: [%s]", keepAddress3.String()),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			signer, err := gr.GetSigners(test.keepAddress)

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

func TestLoadExistingGroups(t *testing.T) {
	persistenceMock := &persistenceHandleMock{}

	signers, err := testSigners()
	if err != nil {
		t.Fatalf("failed to get signer: [%v]", err)
	}

	signer1 := signers[0]
	signer2 := signers[1]
	signer3 := signers[2]

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

	expectedSigners1 := []*tss.ThresholdSigner{signer1}
	actualSigners1, err := gr.GetSigners(keepAddress1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expectedSigners1, actualSigners1) {
		t.Errorf("\nexpected: [%v]\nactual:   [%v]", expectedSigners1, actualSigners1)
	}

	expectedSigners2 := []*tss.ThresholdSigner{signer2, signer3}
	actualSigners2, err := gr.GetSigners(keepAddress2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expectedSigners2, actualSigners2) {
		t.Errorf("\nexpected: [%v]\nactual:   [%v]", expectedSigners2, actualSigners2)
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
	signers, _ := testSigners()
	signer1 := signers[0]
	signer2 := signers[1]
	signer3 := signers[2]

	signerBytes1, _ := signer1.Marshal()
	signerBytes2, _ := signer2.Marshal()
	signerBytes3, _ := signer3.Marshal()

	outputData := make(chan persistence.DataDescriptor, 3)
	outputErrors := make(chan error)

	outputData <- &testDataDescriptor{"/membership_0", keepAddress1.String(), signerBytes1}
	outputData <- &testDataDescriptor{"/membership_0", keepAddress2.String(), signerBytes2}
	outputData <- &testDataDescriptor{"/membership_1", keepAddress2.String(), signerBytes3}

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

func testSigners() ([]*tss.ThresholdSigner, error) {
	signers := make([]*tss.ThresholdSigner, len(groupMemberIDs))

	for i := range groupMemberIDs {
		signer, err := newTestSigner(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get new signer with index [%d]: [%v]", i, err)
		}
		signers[i] = signer
	}
	return signers, nil
}

func newTestSigner(memberIndex int) (*tss.ThresholdSigner, error) {
	testData, err := testdata.LoadKeygenTestFixtures(1)
	if err != nil {
		return nil, fmt.Errorf("failed to load key gen test fixtures: [%v]", err)
	}

	thresholdKey := tss.ThresholdKey(testData[0])
	threshdolKeyBytes, err := thresholdKey.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal threshold key: [%v]", err)
	}

	signer := &tss.ThresholdSigner{}

	pbGroup := &pb.ThresholdSigner_GroupInfo{
		GroupID:            "test-group-1",
		MemberID:           groupMemberIDs[memberIndex],
		GroupMemberIDs:     groupMemberIDs,
		DishonestThreshold: 3,
	}
	pbSigner := &pb.ThresholdSigner{
		GroupInfo:    pbGroup,
		ThresholdKey: threshdolKeyBytes,
	}

	bytes, err := proto.Marshal(pbSigner)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal signer: [%v]", err)
	}

	err = signer.Unmarshal(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal signer: [%v]", err)
	}

	return signer, nil
}
