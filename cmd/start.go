package cmd

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/keep-network/keep-tecdsa/internal/config"
	"github.com/keep-network/keep-tecdsa/pkg/chain/eth/ethereum"
	"github.com/keep-network/keep-tecdsa/pkg/tecdsa"
	"github.com/urfave/cli"
)

// StartCommand contains the definition of the start command-line subcommand.
var StartCommand cli.Command

const startDescription = `Starts the Keep tECDSA client in the foreground.`

func init() {
	StartCommand =
		cli.Command{
			Name:        "start",
			Usage:       `Starts the Keep tECDSA client in the foreground`,
			Description: startDescription,
			Action:      Start,
		}
}

// Start starts a client.
func Start(c *cli.Context) error {
	config, err := config.ReadConfig(c.GlobalString("config"))
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	ethereumChain, err := ethereum.Connect(&config.Ethereum)
	if err != nil {
		return fmt.Errorf("error connecting to Ethereum node: [%v]", err)
	}

	ctx := context.Background()

	tecdsa.Initialize(ethereumChain, &chaincfg.TestNet3Params)

	fmt.Printf("Client started.\n")

	select {
	case <-ctx.Done():
		if err != nil {
			return err
		}

		return fmt.Errorf("unexpected context cancellation")
	}
}
