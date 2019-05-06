package cmd

import (
	"fmt"

	"github.com/keep-network/cli"
	"github.com/keep-network/keep-tecdsa/internal/config"
	"github.com/keep-network/keep-tecdsa/pkg/chain/electrum"
	"github.com/keep-network/keep-tecdsa/pkg/transaction"
)

// PublishCommand contains the definition of the publish command-line subcommand.
var PublishCommand cli.Command

const publishDescription = `The publish command connects to Electrum Server and
sends a transaction to the network.`

func init() {
	PublishCommand = cli.Command{
		Name:        "publish",
		Usage:       "Publish a transaction",
		Description: publishDescription,
		Action:      Publish,
	}
}

// Publish connects to Electrum Server and broadcasts a raw transaction provided
// as a CLI argument.
func Publish(c *cli.Context) error {
	arg := c.Args().First()

	configFile, err := config.ReadConfig(c.GlobalString("config"))
	if err != nil {
		return err
	}

	chain, err := electrum.Connect(&configFile.Electrum)
	if err != nil {
		return err
	}

	result, err := transaction.Publish(chain, arg)
	if err != nil {
		return fmt.Errorf("publish failed [%v]", err)
	}

	fmt.Printf("Published transaction ID: %v\n", result)

	return nil
}
