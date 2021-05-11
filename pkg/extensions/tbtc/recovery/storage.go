package recovery

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-ecdsa/pkg/chain/bitcoin"
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
	err := persistence.CheckStoragePermission(path)
	if err != nil {
		return nil, err
	}

	err = persistence.EnsureDirectoryExists(path, chainName)
	if err != nil {
		return nil, err
	}

	err = persistence.EnsureDirectoryExists(fmt.Sprintf("%s/%s", path, chainName), directoryName)
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
// We algo return the directory and truncated public key separately as a
// convenience for other methods like persistence.EnsureDirectoryExists.
func (dis *DerivationIndexStorage) getStoragePath(extendedPublicKey string) (string, string, string, error) {
	trimmedKey := strings.TrimSpace(extendedPublicKey)
	if len(trimmedKey) < 12 {
		return "", "", "", fmt.Errorf("insufficient length for public key %s", trimmedKey)
	}
	publicKeyDescriptor := trimmedKey[:4]
	suffix := trimmedKey[len(trimmedKey)-8:]
	directory := fmt.Sprintf("%s/%s/%s", dis.path, chainName, directoryName)
	truncatedKey := fmt.Sprintf("%s_%s", publicKeyDescriptor, suffix)
	path := fmt.Sprintf("%s/%s", directory, truncatedKey)
	return path, directory, truncatedKey, nil
}

// save marks an index as used for a particular extendedPublicKey
func (dis *DerivationIndexStorage) save(extendedPublicKey string, index uint32) error {
	dirPath, directory, truncatedKey, err := dis.getStoragePath(extendedPublicKey)
	if err != nil {
		return err
	}

	err = persistence.EnsureDirectoryExists(directory, truncatedKey)
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		logger.Errorf("something went wrong trying to clean up old index files: [%v]", err)
	}

	// clean up old indexes
	for _, file := range files {
		fileIndex, err := strconv.Atoi(file.Name())
		if err != nil {
			logger.Errorf("something went wrong trying to clean up old index files: [%v]", err)
			continue
		}

		if uint32(fileIndex) < index {
			err = os.Remove(fmt.Sprintf("%s/%s", dirPath, file.Name()))
			if err != nil {
				logger.Errorf("something went wrong trying to clean up old index files: [%v]", err)
				continue
			}
		}
	}
	filePath := fmt.Sprintf("%s/%d", dirPath, index)

	return persistence.Write(filePath, []byte{})
}

// Read returns the most recently used index for the extended public key
func (dis *DerivationIndexStorage) read(extendedPublicKey string) (int, error) {
	dirPath, _, _, err := dis.getStoragePath(extendedPublicKey)
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
func (dis *DerivationIndexStorage) GetNextIndex(
	extendedPublicKey string,
	handle bitcoin.Handle,
) (uint32, error) {
	dis.mutex.Lock()
	defer dis.mutex.Unlock()
	dirPath, _, _, err := dis.getStoragePath(extendedPublicKey)
	if err != nil {
		return 0, err
	}
	_, err = os.Stat(dirPath)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	lastIndex, err := dis.read(extendedPublicKey)
	if err != nil {
		return 0, err
	}
	nextIndex := uint32(lastIndex + 1)
	numberOfTries := uint32(100)
	for i := uint32(0); i < numberOfTries; i++ {
		index := nextIndex + i
		derivedAddress, err := deriveAddress(strings.TrimSpace(extendedPublicKey), index)
		if err != nil {
			return 0, err
		}
		ok, err := handle.IsAddressUnused(derivedAddress)
		err = dis.save(extendedPublicKey, index)
		if err != nil {
			return 0, err
		}
		if ok {
			return index, nil
		}
	}
	return 0, fmt.Errorf("indexes %d through %d were all used", nextIndex, nextIndex+numberOfTries-1)
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		logger.Errorf("could not close file [%v]: [%v]", file.Name(), err)
	}
}