package registry

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/gogo/protobuf/proto"

	"github.com/keep-network/keep-ecdsa/internal/testdata"
	"github.com/keep-network/keep-ecdsa/internal/testhelper"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/chain/local"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss/gen/pb"
)

var (
	keepID1String = "0x770a9E2F2Aa1eC2d3Ca916Fc3e6A55058A898632"
	keepID2String = "0x8B3BccB3A3994681A1C1584DE4b4E8b23ed1Ed6d"
	keepID3String = "0x0472ec0185ebb8202f3d4ddb0226998889663cf2"

	localChain chain.Handle

	keepID1, keepID2, keepID3 chain.ID

	groupMemberIDs = [][]byte{
		[]byte("member-1"),
		[]byte("member-2"),
		[]byte("member-3"),
	}
)

func init() {
	localChain = local.Connect(context.Background())

	var err error
	keepID1, err = localChain.UnmarshalID(keepID1String)
	keepID2, _ = localChain.UnmarshalID(keepID2String)
	keepID3, _ = localChain.UnmarshalID(keepID3String)

	if err != nil {
		fmt.Println("booyansky", err)
	}
}

func buildRegistry() (*testhelper.PersistenceHandleMock, *Keeps) {
	persistenceMock := testhelper.NewPersistenceHandleMock(3)

	signers, _ := testSigners()
	signer1 := signers[0]
	signer2 := signers[1]

	persistenceMock.MockSigner(0, keepID1.String(), signer1)
	persistenceMock.MockSigner(0, keepID2.String(), signer2)

	return persistenceMock, NewKeepsRegistry(persistenceMock, localChain.UnmarshalID)
}

func TestRegisterSigner(t *testing.T) {
	persistenceMock, kr := buildRegistry()

	signer1, err := newTestSigner(0)
	if err != nil {
		t.Fatalf("failed to get signer: [%v]", err)
	}

	expectedSignerBytes, err := signer1.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal signer: [%v]", err)
	}

	expectedFile := &testhelper.TestFileInfo{
		Data:      expectedSignerBytes,
		Directory: keepID1.String(),
		Name:      fmt.Sprintf("/membership_%s", signer1.MemberID().String()),
	}

	err = kr.RegisterSigner(keepID1, signer1)
	if err != nil {
		t.Fatalf("failed to register signer: [%v]", err)
	}

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

func TestRegisterSignerDuplicate(t *testing.T) {
	_, kr := buildRegistry()

	signer1, err := newTestSigner(0)
	if err != nil {
		t.Fatalf("failed to get signer: [%v]", err)
	}

	err = kr.RegisterSigner(keepID1, signer1)

	signer2, err := newTestSigner(1)
	if err != nil {
		t.Fatalf("failed to get signer: [%v]", err)
	}

	err = kr.RegisterSigner(keepID1, signer2)

	expectedError := fmt.Errorf("signer for keep [%s] already registered", keepID1.String())
	if !reflect.DeepEqual(expectedError, err) {
		t.Errorf(
			"unexpected error\nexpected: [%v]\nactual:   [%v]",
			expectedError,
			err,
		)
	}
}

func TestSnapshotSigner(t *testing.T) {
	persistenceMock, kr := buildRegistry()

	signer1, err := newTestSigner(0)
	if err != nil {
		t.Fatalf("failed to get signer: [%v]", err)
	}

	expectedSignerBytes, err := signer1.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal signer: [%v]", err)
	}

	expectedFile := &testhelper.TestFileInfo{
		Data:      expectedSignerBytes,
		Directory: keepID1.String(),
		Name:      fmt.Sprintf("/membership_%s", signer1.MemberID().String()),
	}

	err = kr.SnapshotSigner(keepID1, signer1)
	if err != nil {
		t.Fatalf("failed to snapshot signer: [%v]", err)
	}

	if len(persistenceMock.Snapshots) != 1 {
		t.Errorf(
			"unexpected number of persisted groups\nexpected: [%d]\nactual:   [%d]",
			1,
			len(persistenceMock.Snapshots),
		)
	}

	if !reflect.DeepEqual(
		expectedFile,
		persistenceMock.Snapshots[0],
	) {
		t.Errorf(
			"unexpected persisted group\nexpected: [%+v]\nactual:   [%+v]",
			expectedFile,
			persistenceMock.Snapshots[0],
		)
	}
}

func TestUnregisterSigner(t *testing.T) {
	persistenceMock, kr := buildRegistry()

	signer1, err := newTestSigner(0)
	if err != nil {
		t.Fatalf("failed to get signer: [%v]", err)
	}

	err = kr.RegisterSigner(keepID1, signer1)
	if err != nil {
		t.Fatalf("failed to register signer: [%v]", err)
	}

	kr.UnregisterKeep(keepID1)

	if len(persistenceMock.PersistedGroups) != 0 {
		t.Errorf(
			"unexpected number of persisted groups\nexpected: [%d]\nactual:   [%d]",
			1,
			len(persistenceMock.PersistedGroups),
		)
	}

	if len(persistenceMock.ArchivedGroups) != 1 {
		t.Errorf(
			"unexpected number of archived groups\nexpected: [%d]\nactual:   [%d]",
			1,
			len(persistenceMock.ArchivedGroups),
		)
	}
}

func TestGetSigner(t *testing.T) {
	_, kr := buildRegistry()

	signers, err := testSigners()
	if err != nil {
		t.Fatalf("failed to get signer: [%v]", err)
	}

	signer1 := signers[0]

	err = kr.RegisterSigner(keepID1, signer1)
	if err != nil {
		t.Fatalf("failed to register signer: [%v]", err)
	}

	var tests = map[string]struct {
		keepID         chain.ID
		expectedSigner *tss.ThresholdSigner
		expectedError  error
	}{
		"returns registered keep with one signer": {
			keepID:         keepID1,
			expectedSigner: signer1,
		},
		"returns error for not registered keep": {
			keepID: keepID3,
			expectedError: fmt.Errorf(
				"could not find signer for keep: [%s]",
				keepID3.String(),
			),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			signer, err := kr.GetSigner(test.keepID)

			if test.expectedSigner != signer {
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
	signers, err := testSigners()
	if err != nil {
		t.Fatalf("failed to get signer: [%v]", err)
	}

	signer1 := signers[0]
	signer2 := signers[1]

	_, kr := buildRegistry()

	if len(kr.GetKeepsIDs()) != 0 {
		t.Fatal("unexpected keeps number at start")
	}

	kr.LoadExistingKeeps()

	signersCount := 0

	if len(kr.GetKeepsIDs()) != 2 {
		t.Fatalf(
			"unexpected number of keeps\nexpected: [%d]\nactual:   [%d]",
			2,
			signersCount,
		)
	}

	expectedSigner1 := signer1
	actualSigner1, err := kr.GetSigner(keepID1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expectedSigner1, actualSigner1) {
		t.Errorf(
			"\nexpected: [%v]\nactual:   [%v]",
			expectedSigner1,
			actualSigner1,
		)
	}

	expectedSigner2 := signer2
	actualSigner2, err := kr.GetSigner(keepID2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expectedSigner2, actualSigner2) {
		t.Errorf(
			"\nexpected: [%v]\nactual:   [%v]",
			expectedSigner2,
			actualSigner2,
		)
	}
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
