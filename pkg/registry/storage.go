package registry

import (
	"fmt"
	"sync"

	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss"
)

type storage interface {
	save(keepID chain.ID, signer *tss.ThresholdSigner) error
	snapshot(keepID chain.ID, signer *tss.ThresholdSigner) error
	readAll(unmarshalIDFunc func(string) (chain.ID, error)) (<-chan *keepSigner, <-chan error)
	archive(keepID chain.ID) error
}

type persistentStorage struct {
	handle persistence.Handle
}

func newStorage(persistence persistence.Handle) storage {
	return &persistentStorage{
		handle: persistence,
	}
}

func (ps *persistentStorage) save(
	keepID chain.ID,
	signer *tss.ThresholdSigner,
) error {
	signerBytes, err := signer.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal signer: [%v]", err)
	}

	return ps.handle.Save(
		signerBytes,
		keepID.String(),
		// Take just the first 20 bytes of member ID so that we don't produce
		// too long file names.
		fmt.Sprintf("/membership_%.40s", signer.MemberID().String()),
	)
}

func (ps *persistentStorage) snapshot(
	keepID chain.ID,
	signer *tss.ThresholdSigner,
) error {
	signerBytes, err := signer.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal signer: [%v]", err)
	}

	return ps.handle.Snapshot(
		signerBytes,
		keepID.String(),
		// Take just the first 20 bytes of member ID so that we don't produce
		// too long file names.
		fmt.Sprintf("/membership_%.40s", signer.MemberID().String()),
	)
}

type keepSigner struct {
	keepID chain.ID
	signer *tss.ThresholdSigner
}

func (ps *persistentStorage) readAll(
	unmarshalIDFunc func(string) (chain.ID, error),
) (<-chan *keepSigner, <-chan error) {
	outputKeepSigner := make(chan *keepSigner)
	outputErrors := make(chan error)

	inputData, inputErrors := ps.handle.ReadAll()

	// We have two goroutines reading from data and errors channels at the same
	// time. The reason for that is because we don't know in what order
	// producers write information to channels.
	// The third goroutine waits for those two goroutines to finish and it
	// closes the output channels. Channels are not closed by two other goroutines
	// because data goroutine writes both to output signers and errors
	// channel and we want to avoid a situation when we close the errors channel
	// and errors goroutine tries to write to it. The same the other way round.
	var wg sync.WaitGroup
	wg.Add(2)

	// Close channels when signers and errors goroutines are done.
	go func() {
		wg.Wait()
		close(outputKeepSigner)
		close(outputErrors)
	}()

	// Errors goroutine - pass thru errors from input channel to output channel
	// unchanged.
	go func() {
		for err := range inputErrors {
			outputErrors <- err
		}
		wg.Done()
	}()

	// Signers goroutine reads data from input channel, tries to unmarshal
	// the data to Signer and write the unmarshalled Signer to the output signers
	// channel. In case of an error, goroutine writes that error to an output
	// errors channel.
	go func() {
		for descriptor := range inputData {
			content, err := descriptor.Content()
			if err != nil {
				outputErrors <- fmt.Errorf(
					"failed to decode content from file [%v] in directory [%v]: [%v]",
					descriptor.Name(),
					descriptor.Directory(),
					err,
				)
				continue
			}

			keepID, err := unmarshalIDFunc(descriptor.Directory())
			if err != nil {
				outputErrors <- fmt.Errorf(
					"directory name [%v] could not be converted to a keep ID: [%v]",
					descriptor.Directory(),
					keepID,
				)
				continue
			}

			signer := &tss.ThresholdSigner{}
			err = signer.Unmarshal(content)
			if err != nil {
				outputErrors <- fmt.Errorf(
					"failed to unmarshal signer from file [%v] in directory [%v]: [%v]",
					descriptor.Name(),
					descriptor.Directory(),
					err,
				)
				continue
			}

			outputKeepSigner <- &keepSigner{
				keepID: keepID,
				signer: signer,
			}
		}

		wg.Done()
	}()

	return outputKeepSigner, outputErrors
}

func (ps *persistentStorage) archive(keepID chain.ID) error {
	return ps.handle.Archive(keepID.String())
}
