package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/keep-network/keep-ecdsa/config"
	"github.com/urfave/cli"
)

// EthlikeSignature is a common ETH-like signature format.
type EthlikeSignature struct {
	Address   commonAddress `json:"address"`
	Message   string        `json:"msg"`
	Signature string        `json:"sig"`
	Version   string        `json:"version"`
}

const ethlikeSignatureVersion = "2"

type keyFileConfigExtractor func(configFilePath string) (
	keyFilePath string,
	keyFilePassword string,
	err error,
)

// EthlikeSign signs a string using operator's ETH-like key.
func EthlikeSign(
	c *cli.Context,
	extractKeyFile keyFileConfigExtractor,
) error {
	message := c.Args().First()
	if len(message) == 0 {
		return fmt.Errorf("invalid digest")
	}

	var keyFilePath, keyPassword string
	// Check if `key-file` flag was set. If not read the key file path from
	// a config file.
	if keyFilePath = c.String("key-file"); len(keyFilePath) > 0 {
		keyPassword = os.Getenv(config.PasswordEnvVariable)
	} else {
		var err error
		keyFilePath, keyPassword, err = extractKeyFile(
			c.GlobalString("config"),
		)
		if err != nil {
			return fmt.Errorf(
				"could not extract key file data from config file: [%v]",
				err,
			)
		}
	}

	key, err := decryptKeyFile(keyFilePath, keyPassword)
	if err != nil {
		return fmt.Errorf(
			"failed to read key file [%s]: [%v]",
			keyFilePath,
			err,
		)
	}

	signature, err := sign(key, message)
	if err != nil {
		return fmt.Errorf("signing failed: [%v]", err)
	}

	marshaledSignature, err := json.Marshal(signature)
	if err != nil {
		return fmt.Errorf("failed to marshal signature: [%v]", err)
	}

	// store to user writeable file
	return outputData(c, marshaledSignature, 0644)
}

// EthlikeVerify verifies if a signature was calculated by a signer with the
// given ETH-like address.
func EthlikeVerify(c *cli.Context) error {
	var marshaledSignature []byte
	if inputFilePath := c.String("input-file"); len(inputFilePath) > 0 {
		fileContent, err := ioutil.ReadFile(filepath.Clean(inputFilePath))
		if err != nil {
			return fmt.Errorf("failed to read a file: [%v]", err)
		}
		marshaledSignature = fileContent
	} else {
		signatureArg := c.Args().First()
		if len(signatureArg) == 0 {
			return fmt.Errorf("missing argument")
		}

		marshaledSignature = []byte(signatureArg)
	}

	signature := &EthlikeSignature{}
	err := json.Unmarshal(marshaledSignature, signature)
	if err != nil {
		return fmt.Errorf("failed to unmarshal signature: [%v]", err)
	}

	err = verify(signature)
	if err != nil {
		return fmt.Errorf("signature verification failed: [%v]", err)
	}

	fmt.Printf(
		"signature verified correctly, "+
			"message [%s] was signed by [%s]\n",
		signature.Message,
		signature.Address.Hex(),
	)

	return nil
}

func sign(key *keystoreKey, message string) (*EthlikeSignature, error) {
	digest := accountsTextHash([]byte(message))

	signature, err := cryptoSign(digest[:], key.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: [%v]", err)
	}

	return &EthlikeSignature{
		Address:   key.Address,
		Message:   message,
		Signature: hexutilEncode(signature),
		Version:   ethlikeSignatureVersion,
	}, nil
}

func verify(signature *EthlikeSignature) error {
	if signature.Version != ethlikeSignatureVersion {
		return fmt.Errorf(
			"unsupported signature version\n"+
				"\texpected: %s\n"+
				"\tactual:   %s",
			ethlikeSignatureVersion,
			signature.Version,
		)
	}

	digest := accountsTextHash([]byte(signature.Message))

	signatureBytes, err := hexutilDecode(signature.Signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: [%v]", err)
	}

	publicKey, err := cryptoSigToPub(digest[:], signatureBytes)
	if err != nil {
		return fmt.Errorf(
			"could not recover public key from signature [%v]",
			err,
		)
	}

	recoveredAddress := cryptoPubkeyToAddress(*publicKey)

	if !bytes.Equal(recoveredAddress.Bytes(), signature.Address.Bytes()) {
		return fmt.Errorf(
			"invalid signer\n"+
				"\texpected: %s\n"+
				"\tactual:   %s",
			signature.Address.Hex(),
			recoveredAddress.Hex(),
		)
	}

	return nil
}
