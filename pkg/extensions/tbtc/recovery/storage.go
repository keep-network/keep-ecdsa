package recovery

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	chainName     = "bitcoin"
	directoryName = "derivation_indexes"
)

// DerivationIndexStorage provides access to the derivation index persistence
// API, which makes sure we're not reusing derived wallet addresses.
type DerivationIndexStorage struct {
	path  string
	mutex sync.Mutex
}

// NewDerivationIndexStorage is a factory method that creates a new DerivationIndexStorage at the specified path.
func NewDerivationIndexStorage(path string) (*DerivationIndexStorage, error) {
	err := ensureDirectoryExists(fmt.Sprintf("%s/%s", path, chainName))
	if err != nil {
		return nil, err
	}

	err = ensureDirectoryExists(fmt.Sprintf("%s/%s/%s", path, chainName, directoryName))
	if err != nil {
		return nil, err
	}
	return &DerivationIndexStorage{
		path: path,
	}, nil
}

// getStoragePath stores an extended public key as its 4-letter descriptor
// followed by an underscore and then it's 8-letter suffix. For example:
// xpub6Cg41S21VrxkW1WBTZJn95KNpHozP2Xc6AhG27ZcvZvH8XyNzunEqLdk9dxyXQUoy7ALWQFNn5K1me74aEMtS6pUgNDuCYTTMsJzCAk9sk1 => xpub_zCAk9sk1
// ypub6Xxan668aiJqvh4SVfd7EzqjWvf36gWufTkhWHv3gaxnBh44HpkTi2TTkm1u136qjUxk7F3jGzoyfrGpHvALMgJgbF4WNXpoPu3QYrqogMK => ypub_QYrqogMK
// zpub6rePDVHfRP14VpYiejwepBhzu45UbvqvzE3ZMdDnNykG47mZYyGTjsuq6uzQYRakSrHyix1YTXKohag4GDZLcHcLvhSAs2MQNF8VDaZuQT9 => zpub_VDaZuQT9
// This both obfuscates the whole extended key and makes the folder easier to digest for human reading.
func (dis *DerivationIndexStorage) getStoragePath(extendedPublicKey string) (string, error) {
	trimmedKey := strings.TrimSpace(extendedPublicKey)
	if len(trimmedKey) < 12 {
		return "", fmt.Errorf("insufficient length for public key %s", trimmedKey)
	}
	publicKeyDescriptor := trimmedKey[:4]
	suffix := trimmedKey[len(trimmedKey)-8:]
	return fmt.Sprintf("%s/%s/%s/%s_%s", dis.path, chainName, directoryName, publicKeyDescriptor, suffix), nil
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
func (dis *DerivationIndexStorage) Save(extendedPublicKey string, index uint32, btcAddress string) error {
	dis.mutex.Lock()
	defer dis.mutex.Unlock()
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

// Read returns the most recently used index for the extended public key
func (dis *DerivationIndexStorage) read(extendedPublicKey string) (int, error) {
	dirPath, err := dis.getStoragePath(extendedPublicKey)
	if err != nil {
		return 0, err
	}

	index := -1
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return 0, err
	}

	for _, file := range files {
		fileIndex, err := strconv.Atoi(file.Name())
		if err != nil {
			return 0, err
		}

		if fileIndex > index {
			index = fileIndex
		}
	}
	return index, nil
}

// GetNextIndex returns the next unused index for the extended public key
func (dis *DerivationIndexStorage) GetNextIndex(extendedPublicKey string) (uint32, error) {
	dis.mutex.Lock()
	defer dis.mutex.Unlock()
	index, err := dis.read(extendedPublicKey)
	if err != nil {
		return 0, err
	}
	return uint32(index + 1), nil
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
