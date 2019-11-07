// Package mock contains implementations which can be used in tests to mock
// external tools integration used in testing.
package mock

import (
	"fmt"

	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-tecdsa/internal/testdata"
)

type PersistenceHandle struct {
	PersistedGroups []*TestFileInfo
	ArchivedGroups  []string
}

type TestFileInfo struct {
	Data      []byte
	Directory string
	Name      string
}

func (phm *PersistenceHandle) Save(data []byte, directory string, name string) error {
	phm.PersistedGroups = append(
		phm.PersistedGroups,
		&TestFileInfo{data, directory, name},
	)
	return nil
}

func (phm *PersistenceHandle) ReadAll() (<-chan persistence.DataDescriptor, <-chan error) {
	outputData := make(chan persistence.DataDescriptor, 2)
	outputErrors := make(chan error)

	for keepAddress, signer := range testdata.KeepSigners {
		signerBytes, err := signer.Marshal()
		if err != nil {
			outputErrors <- fmt.Errorf("failed to marshal signer: [%v]", err)
		}

		outputData <- &testDataDescriptor{"/membership_0", keepAddress.String(), signerBytes}
	}

	close(outputData)
	close(outputErrors)

	return outputData, outputErrors
}

func (phm *PersistenceHandle) Archive(directory string) error {
	phm.ArchivedGroups = append(phm.ArchivedGroups, directory)

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
