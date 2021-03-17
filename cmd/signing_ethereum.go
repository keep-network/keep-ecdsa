//+build !celo

package cmd

import (
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-ecdsa/config"
	"github.com/urfave/cli"
)

// Type aliases and function variables needed to expose specific chain types
// without forcing the client code to directly import the host chain module.
// They are used by the common code from `signing_ethlike.go` file.
type (
	keystoreKey   = keystore.Key
	commonAddress = common.Address
)

var (
	decryptKeyFile        = ethutil.DecryptKeyFile
	accountsTextHash      = accounts.TextHash
	hexutilEncode         = hexutil.Encode
	hexutilDecode         = hexutil.Decode
	cryptoSigToPub        = crypto.SigToPub
	cryptoPubkeyToAddress = crypto.PubkeyToAddress
	cryptoSign            = crypto.Sign
)

// ChainSigningCommand contains the definition of the `signing ethereum`
// command-line subcommand and its own subcommands.
var ChainSigningCommand = cli.Command{
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
					Name: "key-file,k",
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
			ArgsUsage:   "[signature]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input-file,i",
					Usage: "Input file with the signature",
				},
			},
		},
	},
}

const ethereumSignDescription = `Calculates an ethereum signature for a given 
message. The message is expected to be provided as a string, it is later hashed 
with ethereum's hashing algorithm and passed to Ethereum ECDSA signing. 
Signature is calculated in Ethereum specific format as a hexadecimal string 
representation of 65-byte {R, S, V} parameters, where V is 0 or 1.

It requires an Ethereum key to be provided in an encrypted file. A path to the 
key file can be configured in a config file or specified directly with an 
'eth-key-file' flag.

The key file is expected to be encrypted with a password provided 
as ` + config.PasswordEnvVariable + `environment variable.
	
The result is outputted in a common Ethereum signature format:
{
	"address": "<address>",
	"msg": "<content>",
	"sig": "<signature>",
	"version": "2"
}

If 'output-file' flag is set the result will be stored in a specified file path.
`

const ethereumVerifyDescription = `Verifies if a signature was calculated for 
a message by an ethereum account identified by an address. 

It expects a signature to be provided in a common Ethereum signature format:
{
	"address": "<address>",
	"msg": "<content>",
	"sig": "<signature>",
	"version": "2"
}

If 'input-file' flag is set the input will be read from a specified file path.
`

// EthereumSign signs a string using operator's ethereum key.
func EthereumSign(c *cli.Context) error {
	keyFileConfigExtractor := func(
		configFilePath string,
	) (string, string, error) {
		config, err := config.ReadEthereumConfig(configFilePath)
		if err != nil {
			return "", "", err
		}

		return config.Account.KeyFile, config.Account.KeyFilePassword, nil
	}

	return EthlikeSign(c, keyFileConfigExtractor)
}

// EthereumVerify verifies if a signature was calculated by a signer with the
// given ethereum address.
func EthereumVerify(c *cli.Context) error {
	return EthlikeVerify(c)
}
