//+build !celo

package cmd

import (
	chaincmd "github.com/keep-network/keep-ecdsa/pkg/chain/gen/ethereum/cmd"
	"github.com/urfave/cli"
)

// ChainCLICommand contains the definition of the ethereum command-line
// subcommand and its own subcommands.
var ChainCLICommand cli.Command

const ethereumDescription = `The ethereum command allows interacting with Keep's Ethereum
	contracts directly. Each subcommand corresponds to one contract, and has
	subcommands corresponding to each method on that contract, which respectively
	each take parameters based on the contract method's parameters.

    See the subcommand help for additional details.`

func init() {
	ChainCLICommand = cli.Command{
		Name:        "ethereum",
		Usage:       `Provides access to Keep network Ethereum contracts.`,
		Description: ethereumDescription,
		Subcommands: chaincmd.AvailableCommands,
	}
}
