package recovery

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

const (
	fileName = "derivation_indexes"
)

// DerivationIndexStorage provides access to the derivation index persistence
// API, which makes sure we're not reusing derived wallet addresses.
type DerivationIndexStorage struct {
	path string
}

// NewDerivationIndexStorage is a factory method that creates a new DerivationIndexStorage at the specified path.
func NewDerivationIndexStorage(path string) (*DerivationIndexStorage, error) {
	err := ensureDirectoryExists(fmt.Sprintf("%s/%s", path, fileName))
	if err != nil {
		return nil, err
	}
	return &DerivationIndexStorage{
		path: path,
	}, nil
}

func (dis DerivationIndexStorage) getStoragePath(extendedPublicKey string) (string, error) {
	if len(extendedPublicKey) < 12 {
		return "", fmt.Errorf("insufficient length for public key %s", extendedPublicKey)
	}
	publicKeyDescriptor := extendedPublicKey[:4]
	suffix := extendedPublicKey[len(extendedPublicKey)-8:]
	return fmt.Sprintf("%s/%s/%s_%s", dis.path, fileName, publicKeyDescriptor, suffix), nil
}

func ensureDirectoryExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error occurred while creating a dir: [%v]", err)
		}
	}

	return nil
}

// Save marks an index as used for a particular extendedPublicKey
func (dis DerivationIndexStorage) Save(extendedPublicKey string, index int, btcAddress string) error {
	dirPath, err := dis.getStoragePath(extendedPublicKey)
	if err != nil {
		return err
	}

	err = ensureDirectoryExists(dirPath)
	if err != nil {
		return err
	}
	filePath := fmt.Sprintf("%s/%d", dirPath, index)

	return write(filePath, []byte(btcAddress))
}

// Read returns the next unused index for the extended public key
func (dis DerivationIndexStorage) Read(extendedPublicKey string) (int, error) {
	dirPath, err := dis.getStoragePath(extendedPublicKey)
	if err != nil {
		return 0, err
	}

	err = ensureDirectoryExists(dirPath)
	if err != nil {
		return 0, err
	}

	index := -1
	files, err := ioutil.ReadDir(dirPath)
	for _, file := range files {
		fileIndex, err := strconv.Atoi(file.Name())
		if err != nil {
			return 0, err
		}
		if fileIndex > index {
			index = fileIndex
		}
	}
	return index + 1, nil
}

// create and write data to a file
func write(filePath string, data []byte) error {
	writeFile, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer closeFile(writeFile)

	_, err = writeFile.Write(data)
	if err != nil {
		return err
	}

	err = writeFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		logger.Errorf("could not close file [%v]: [%v]", file.Name(), err)
	}
}
