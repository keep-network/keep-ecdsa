package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

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
			Usage:       "Sign a message using operator's key",
			Description: ethereumSignDescription,
			Action:      EthereumSign,
			ArgsUsage:   "[message]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "eth-key-file,k",
					Usage: "Path to the ethereum key file. " +
						"If not provided read the path from a config file.",
				},
			},
		},
		{
			Name:        "verify",
			Usage:       "Verifies a signature",
			Description: ethereumVerifyDescription,
			Action:      EthereumVerify,
			ArgsUsage:   "[message] [address] [signature]",
		},
	},
}

const ethereumSignDescription = "Calculates a signature for a given message. " +
	"The message is expected to be provided as a string, it is later hashed with " +
	"SHA-256 and passed to ECDSA signing. The calculated signature is returned " +
	"in Ethereum specific format as a hexadecimal string representation of 65-byte " +
	"{R, S, V} parameters.\n" +
	"It reads an Ethereum key from an encrypted file. A path to the key file can be " +
	"configured in a config file or specified directly with a `eth-key-file` flag. " +
	"The key file is expected to be encrypted with a password provided as " +
	config.PasswordEnvVariable + " environment variable."

const ethereumVerifyDescription = "Verifies if a signature was calculated for a " +
	"message by an ethereum account identified by an address. It expects a message " +
	"to be provided as an original message string, that is later hashed with SHA-256. " +
	"Signature should be provided in Ethereum specific format as a hexadecimal string " +
	"representation of 65-byte {R, S, V} parameters."

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

	signatureHex := hex.EncodeToString(signature)

	fmt.Printf(
		"signature calculated by [%s] for message [%s]: %s\n",
		ethereumKey.Address.Hex(), message, signatureHex,
	)

	return nil
}

// EthereumVerify verifies if a signature was calculated by a signer with the
// given ethereum address.
func EthereumVerify(c *cli.Context) error {
	message := c.Args().First()
	if len(message) == 0 {
		return fmt.Errorf("invalid message")
	}

	address := c.Args().Get(1)
	if len(address) == 0 {
		return fmt.Errorf("invalid address")
	}

	signature := c.Args().Get(2)
	if len(signature) == 0 {
		return fmt.Errorf("invalid signature")
	}

	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("could not decode signature string: [%v]", err)
	}

	digest := sha256.Sum256([]byte(message))

	publicKey, err := crypto.SigToPub(digest[:], signatureBytes)
	if err != nil {
		return fmt.Errorf("could not recover public key from signature [%v]", err)
	}

	expectedAddress := common.HexToAddress(address)
	recoveredAddress := crypto.PubkeyToAddress(*publicKey)

	if !bytes.Equal(recoveredAddress.Bytes(), expectedAddress.Bytes()) {
		return fmt.Errorf(
			"invalid signer\n"+
				"\texpected signer: %s\n"+
				"\tactual signer:   %s",
			expectedAddress.Hex(),
			recoveredAddress.Hex(),
		)
	}

	fmt.Printf(
		"signature verified correctly, message [%s] was signed by [%s]\n",
		message,
		recoveredAddress.Hex(),
	)

	return nil
}
