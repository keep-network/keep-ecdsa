package testdata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/pkg/errors"
)

const (
	testFixtureDirFormat  = "%s/tss"
	testFixtureFileFormat = "keygen_data_%d.json"
)

// LoadKeygenTestFixtures loads key generation test data.
// Code copied from:
//   https://github.com/binance-chain/tss-lib/blob/master/ecdsa/keygen/test_utils.go
// Test data JSON files copied from:
//   https://github.com/binance-chain/tss-lib/tree/master/test/_fixtures
func LoadKeygenTestFixtures(count int) ([]keygen.LocalPartySaveData, error) {
	keys := make([]keygen.LocalPartySaveData, 0, count)
	for j := 0; j < count; j++ {
		// We do `% 5` because only 5 files are stored in test fixtures path.
		// We loop over the files which allows us to support `count > 5`.
		fixtureFilePath := makeTestFixtureFilePath(j % 5)

		bz, err := ioutil.ReadFile(fixtureFilePath)
		if err != nil {
			return nil, errors.Wrapf(err,
				"could not open the test fixture for party %d in the expected location: %s. run keygen tests first.",
				j, fixtureFilePath)
		}
		var key keygen.LocalPartySaveData
		if err = json.Unmarshal(bz, &key); err != nil {
			return nil, errors.Wrapf(err,
				"could not unmarshal fixture data for party %d located at: %s",
				j, fixtureFilePath)
		}
		keys = append(keys, keygen.LocalPartySaveData{
			LocalPreParams: keygen.LocalPreParams{
				PaillierSK: key.PaillierSK,
				NTildei:    key.NTildei,
				H1i:        key.H1i,
				H2i:        key.H2i,
			},
			LocalSecrets: keygen.LocalSecrets{
				Xi:      key.Xi,
				ShareID: key.ShareID,
			},
			Ks:          key.Ks[:count],
			NTildej:     key.NTildej[:count],
			H1j:         key.H1j[:count],
			H2j:         key.H2j[:count],
			BigXj:       key.BigXj[:count],
			PaillierPKs: key.PaillierPKs[:count],
			ECDSAPub:    key.ECDSAPub,
		})
	}
	return keys, nil
}

func makeTestFixtureFilePath(partyIndex int) string {
	_, callerFileName, _, _ := runtime.Caller(0)
	srcDirName := filepath.Dir(callerFileName)
	fixtureDirName := fmt.Sprintf(testFixtureDirFormat, srcDirName)
	return fmt.Sprintf("%s/"+testFixtureFileFormat, fixtureDirName, partyIndex)
}
