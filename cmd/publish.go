package cmd

import (
	"fmt"

	"github.com/keep-network/cli"
	"github.com/keep-network/keep-tecdsa/internal/config"
	"github.com/keep-network/keep-tecdsa/pkg/chain"
	"github.com/keep-network/keep-tecdsa/pkg/chain/blockcypher"
	"github.com/keep-network/keep-tecdsa/pkg/chain/electrum"
	"github.com/keep-network/keep-tecdsa/pkg/transaction"
)

// PublishCommand contains the definition of the publish command-line subcommand.
var PublishCommand cli.Command

const publishDescription = `The publish command broadcasts a transaction using 
the specific external service.`

func init() {
	PublishCommand = cli.Command{
		Name:        "publish",
		Usage:       "Publish a transaction",
		Description: publishDescription,
		Action:      Publish,
	}
}

// Publish sends a raw transaction provided as a CLI argument.
func Publish(c *cli.Context) error {
	arg := c.Args().First()

	configFile, err := config.ReadConfig(c.GlobalString("config"))
	if err != nil {
		return err
	}

	var chain chain.Interface

	switch chainFlag := c.GlobalString("broadcast-api"); chainFlag {
	case "blockcypher":
		chain, err = blockcypher.Connect(&configFile.BlockCypher)
		if err != nil {
			return err
		}
	case "electrum":
		chain, err = electrum.Connect(&configFile.Electrum)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown transaction publication service")
	}

	result, err := transaction.Publish(chain, arg)
	if err != nil {
		return fmt.Errorf("publish failed [%v]", err)
	}

	fmt.Printf("Published transaction ID: %v\n", result)

	return nil
}
