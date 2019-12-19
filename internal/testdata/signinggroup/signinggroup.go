package signinggroup

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"

	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
)

const testFileFormat = "signer_%d.json"

type signerTestData struct {
	NetworkKey []byte
	Signer     []byte
}

// StoreSigner stores a signer data to the test data storage.
func StoreSigner(
	index int,
	networkKey *key.NetworkPublic,
	signer *tss.ThresholdSigner,
) error {
	filePath := makeFilePath(testFileFormat, index)

	signerBytes, err := signer.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal signer: [%v]", err)
	}

	keyBytes, err := networkKey.Bytes()
	if err != nil {
		return fmt.Errorf("failed to marshal network key: [%v]", err)
	}

	content, err := json.MarshalIndent(
		&signerTestData{
			NetworkKey: keyBytes,
			Signer:     signerBytes,
		},
		"",
		"  ",
	)
	if err != nil {
		return fmt.Errorf("failed to marshal signer test data: [%v]", err)
	}

	if err := ioutil.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: [%v]", err)
	}

	return nil
}

// LoadSigner loads signer test data of given index.
func LoadSigner(index int) (*key.NetworkPublic, *tss.ThresholdSigner, error) {
	filePath := makeFilePath(testFileFormat, index)

	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file [%s]: [%v]", filePath, err)
	}

	signerData := &signerTestData{}
	err = json.Unmarshal(fileContent, signerData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal signer test data: [%v]", err)
	}

	libp2pKey, err := libp2pcrypto.UnmarshalPublicKey(signerData.NetworkKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal network key: [%v]", err)
	}

	networkKey := key.Libp2pKeyToNetworkKey(libp2pKey)

	thresholdSigner := &tss.ThresholdSigner{}
	err = thresholdSigner.Unmarshal(signerData.Signer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal threshold signer: [%v]", err)
	}

	return networkKey, thresholdSigner, nil
}

func makeFilePath(fileNameFormat string, memberIndex int) string {
	_, callerFileName, _, _ := runtime.Caller(0)
	srcDirName := filepath.Dir(callerFileName)

	fileName := fmt.Sprintf(fileNameFormat, memberIndex)

	return fmt.Sprintf("%s/%s", srcDirName, fileName)
}
