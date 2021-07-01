package testhelper

import (
	"fmt"

	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss"
)

// PersistenceHandleMock is a mock of persistence handle used in tests.
type PersistenceHandleMock struct {
	PersistedGroups  []*TestFileInfo
	Snapshots        []*TestFileInfo
	ArchivedGroups   []string
	outputDataChan   chan persistence.DataDescriptor
	outputErrorsChan chan error
}

// NewPersistenceHandleMock creates a mocked persistence handle.
func NewPersistenceHandleMock(outputDataChanSize int) *PersistenceHandleMock {
	return &PersistenceHandleMock{
		outputDataChan:   make(chan persistence.DataDescriptor, outputDataChanSize),
		outputErrorsChan: make(chan error),
	}
}

// TestFileInfo holds test data stored in persistence handle.
type TestFileInfo struct {
	Data      []byte
	Directory string
	Name      string
}

// Save stores data in persistence handle.
func (phm *PersistenceHandleMock) Save(data []byte, directory string, name string) error {
	phm.PersistedGroups = append(
		phm.PersistedGroups,
		&TestFileInfo{data, directory, name},
	)

	return nil
}

// Snapshot creates a snapshot of data in persistence handle.
func (phm *PersistenceHandleMock) Snapshot(data []byte, directory string, name string) error {
	phm.Snapshots = append(
		phm.Snapshots,
		&TestFileInfo{data, directory, name},
	)

	return nil
}

// MockSigner registers a mock of a signer for membership and keep.
func (phm *PersistenceHandleMock) MockSigner(membershipIndex int, keepID string, signer *tss.ThresholdSigner) error {
	signerBytes, err := signer.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal signer: %w", err)
	}

	phm.outputDataChan <- &testDataDescriptor{
		fmt.Sprintf("/membership_%d", membershipIndex),
		keepID,
		signerBytes,
	}

	return nil
}

// ReadAll reads all data stored in persistence handle.
func (phm *PersistenceHandleMock) ReadAll() (<-chan persistence.DataDescriptor, <-chan error) {
	close(phm.outputDataChan)
	close(phm.outputErrorsChan)

	return phm.outputDataChan, phm.outputErrorsChan
}

// Archive archives data in persistence handle.
func (phm *PersistenceHandleMock) Archive(directory string) error {
	phm.ArchivedGroups = append(phm.ArchivedGroups, directory)
	phm.PersistedGroups = phm.PersistedGroups[:len(phm.ArchivedGroups)-1]

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
