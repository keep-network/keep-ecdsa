package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli"
)

// StartLocalCommand contains the definition of the local-signer command-line
// subcommand.
var StartLocalCommand cli.Command

const startLocalDescription = `Starts three Keep tECDSA client in the foreground in 
local signer mode. It requires three config files named 'config_1.toml', 'config_2.toml', 
'config_3.toml' for each client to be provided in 'configs' directory.`

func init() {
	StartLocalCommand = cli.Command{
		Name:        "start-local",
		Usage:       "Starts local signer version of Keep ECDSA client",
		Description: startLocalDescription,
		Action:      StartLocal,
	}
}

func StartLocal(c *cli.Context) error {
	groupSize := 3

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, groupSize)

	for index := 1; index <= groupSize; index++ {
		go func(i int) {
			configPath := fmt.Sprintf("./configs/config_%d.toml", i)

			c.GlobalSet("config", configPath)

			err := Start(c)
			if err != nil {
				errChan <- err
				cancel()
			}

			logger.Infof("client %d started", i)
		}(index)
	}

	select {
	case <-ctx.Done():
		return <-errChan
	}
}
