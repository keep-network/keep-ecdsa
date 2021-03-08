//+build celo

package cmd

import (
	"github.com/celo-org/celo-blockchain/accounts"
	"github.com/celo-org/celo-blockchain/accounts/keystore"
	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/common/hexutil"
	"github.com/celo-org/celo-blockchain/crypto"
	"github.com/keep-network/keep-common/pkg/chain/celo/celoutil"
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
	decryptKeyFile        = celoutil.DecryptKeyFile
	accountsTextHash      = accounts.TextHash
	hexutilEncode         = hexutil.Encode
	hexutilDecode         = hexutil.Decode
	cryptoSigToPub        = crypto.SigToPub
	cryptoPubkeyToAddress = crypto.PubkeyToAddress
	cryptoSign            = crypto.Sign
)

// ChainSigningCommand contains the definition of the `signing celo`
// command-line subcommand and its own subcommands.
var ChainSigningCommand = cli.Command{
	Name:  "celo",
	Usage: "Celo signatures calculation",
	Subcommands: []cli.Command{
		{
			Name:        "sign",
			Usage:       "Sign a message using the operator's key",
			Description: celoSignDescription,
			Action:      CeloSign,
			ArgsUsage:   "[message]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "key-file,k",
					Usage: "Path to the celo key file. " +
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
			Description: celoVerifyDescription,
			Action:      CeloVerify,
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

const celoSignDescription = `Calculates an celo signature for a given message.
The message is expected to be provided as a string, it is later hashed with
celo's hashing algorithm and passed to Celo ECDSA signing. Signature is
calculated in Celo specific format as a hexadecimal string representation of
65-byte {R, S, V} parameters, where V is 0 or 1.

It requires an Celo key to be provided in an encrypted file. A path to the key 
file can be configured in a config file or specified directly with an 
'celo-key-file' flag.

The key file is expected to be encrypted with a password provided 
as ` + config.PasswordEnvVariable + `environment variable.
	
The result is outputted in a common Celo signature format:
{
	"address": "<address>",
	"msg": "<content>",
	"sig": "<signature>",
	"version": "2"
}

If 'output-file' flag is set the result will be stored in a specified file path.
`

const celoVerifyDescription = `Verifies if a signature was calculated for a 
message by a celo account identified by an address. 

It expects a signature to be provided in a common Celo signature format:
{
	"address": "<address>",
	"msg": "<content>",
	"sig": "<signature>",
	"version": "2"
}

If 'input-file' flag is set the input will be read from a specified file path.
`

// CeloSign signs a string using operator's celo key.
func CeloSign(c *cli.Context) error {
	keyFileConfigExtractor := func(
		configFilePath string,
	) (string, string, error) {
		config, err := config.ReadCeloConfig(configFilePath)
		if err != nil {
			return "", "", err
		}

		return config.Account.KeyFile, config.Account.KeyFilePassword, nil
	}

	return EthlikeSign(c, keyFileConfigExtractor)
}

// CeloVerify verifies if a signature was calculated by a signer with the
// given celo address.
func CeloVerify(c *cli.Context) error {
	return EthlikeVerify(c)
}
