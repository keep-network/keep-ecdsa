package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-ecdsa/internal/config"
	"github.com/urfave/cli"
)

// EthereumSigningCommand contains the definition of the `signing ethereum`
// command-line subcommand and its own subcommands.
var EthereumSigningCommand = cli.Command{
	Name:  "ethereum",
	Usage: "Ethereum signatures calculation",
	Subcommands: []cli.Command{
		{
			Name:        "sign",
			Usage:       "Sign a message using the operator's key",
			Description: ethereumSignDescription,
			Action:      EthereumSign,
			ArgsUsage:   "[message]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "eth-key-file,k",
					Usage: "Path to the ethereum key file. " +
						"If not provided read the path from a config file.",
				},
				cli.StringFlag{
					Name:  "output-file,o",
					Usage: "Output file for the signature",
				},
			},
		},
		{
			Name:        "verify",
			Usage:       "Verifies a signature",
			Description: ethereumVerifyDescription,
			Action:      EthereumVerify,
			ArgsUsage:   "[ethereum-signature]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input-file,i",
					Usage: "Input file with the signature",
				},
			},
		},
	},
}

const ethereumSignDescription = `Calculates an ethereum signature for a given message.
The message is expected to be provided as a string, it is later hashed with SHA-256
and passed to Ethereum ECDSA signing. Signature is calculated in Ethereum specific
format as a hexadecimal string representation of 65-byte {R, S, V} parameters.

It requires an Ethereum key to be provided in an encrypted file. A path to the key file
can be configured in a config file or specified directly with an 'eth-key-file' flag.

The key file is expected to be encrypted with a password provided as ` + config.PasswordEnvVariable + `
environment variable.
	
The result is outputted in a common Ethereum signature format:
{
	"address": "<address>",
	"msg": "<content>",
	"sig": "<signature>",
	"version": "2"
}

If 'output-file' flag is set the result will be stored in a specified file path.
`

const ethereumVerifyDescription = `Verifies if a signature was calculated for a message 
by an ethereum account identified by an address. 

It expects a signature to be provided in a common Ethereum signature format:
{
	"address": "<address>",
	"msg": "<content>",
	"sig": "<signature>",
	"version": "2"
}

If 'input-file' flag is set the input will be read from a specified file path.
`

// EthereumSignature is a common Ethereum signature format.
type EthereumSignature struct {
	Address   common.Address `json:"address"`
	Message   string         `json:"msg"`
	Signature string         `json:"sig"`
	Version   uint           `json:"version"`
}

const ethSignatureVersion uint = 2

// EthereumSign signs a string using operator's ethereum key.
func EthereumSign(c *cli.Context) error {
	message := c.Args().First()
	if len(message) == 0 {
		return fmt.Errorf("invalid digest")
	}

	var ethKeyFilePath, ethKeyPassword string
	// Check if `eth-key-file` flag was set. If not read the key file path from
	// a config file.
	if ethKeyFilePath = c.String("eth-key-file"); len(ethKeyFilePath) > 0 {
		ethKeyPassword = os.Getenv(config.PasswordEnvVariable)
	} else {
		ethereumConfig, err := config.ReadEthereumConfig(c.GlobalString("config"))
		if err != nil {
			return fmt.Errorf("failed while reading config file: [%v]", err)
		}

		ethKeyFilePath = ethereumConfig.Account.KeyFile
		ethKeyPassword = ethereumConfig.Account.KeyFilePassword
	}

	ethereumKey, err := ethutil.DecryptKeyFile(ethKeyFilePath, ethKeyPassword)
	if err != nil {
		return fmt.Errorf(
			"failed to read key file [%s]: [%v]",
			ethKeyFilePath,
			err,
		)
	}

	digest := sha256.Sum256([]byte(message))

	signature, err := crypto.Sign(digest[:], ethereumKey.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to sign: [%v]", err)
	}

	ethereumSignature := &EthereumSignature{
		Address:   ethereumKey.Address,
		Message:   message,
		Signature: hex.EncodeToString(signature),
		Version:   ethSignatureVersion,
	}

	marshaledSignature, err := json.Marshal(ethereumSignature)
	if err != nil {
		return fmt.Errorf("failed to marshal ethereum signature: [%v]", err)
	}

	return outputData(c, marshaledSignature, 0644) // store to user writeable file
}

// EthereumVerify verifies if a signature was calculated by a signer with the
// given ethereum address.
func EthereumVerify(c *cli.Context) error {
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

	ethereumSignature := &EthereumSignature{}
	err := json.Unmarshal(marshaledSignature, ethereumSignature)
	if err != nil {
		return fmt.Errorf("failed to unmarshal ethereum signature: [%v]", err)
	}

	if ethereumSignature.Version != ethSignatureVersion {
		return fmt.Errorf(
			"unsupported ethereum signature version\n"+
				"\texpected: %d\n"+
				"\tactual:   %d",
			ethSignatureVersion,
			ethereumSignature.Version,
		)
	}

	digest := sha256.Sum256([]byte(ethereumSignature.Message))

	signatureBytes, err := hex.DecodeString(ethereumSignature.Signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: [%v]", err)
	}

	publicKey, err := crypto.SigToPub(digest[:], signatureBytes)
	if err != nil {
		return fmt.Errorf("could not recover public key from signature [%v]", err)
	}

	recoveredAddress := crypto.PubkeyToAddress(*publicKey)

	if !bytes.Equal(recoveredAddress.Bytes(), ethereumSignature.Address.Bytes()) {
		return fmt.Errorf(
			"signature verification failed: invalid signer\n"+
				"\texpected: %s\n"+
				"\tactual:   %s",
			ethereumSignature.Address.Hex(),
			recoveredAddress.Hex(),
		)
	}

	fmt.Printf(
		"signature verified correctly, message [%s] was signed by [%s]\n",
		ethereumSignature.Message,
		recoveredAddress.Hex(),
	)

	return nil
}
