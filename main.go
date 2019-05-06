package main

import (
	"log"
	"os"
	"path"
	"time"

	"github.com/keep-network/keep-tecdsa/cmd"
	"github.com/keep-network/cli"
)

const (
	defaultConfigPath = "./configs/config.toml"
	defaultChain      = "blockcypher"
)

var (
	configPath string
	chain      string
)

func main() {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "CLI for t-ECDSA keep"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name:  "Keep Network",
			Email: "info@keep.network",
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config,c",
			Value:       defaultConfigPath,
			Destination: &configPath,
			Usage:       "full path to the configuration file",
		},
		cli.StringFlag{
			Name:        "chain",
			Value:       defaultChain,
			Destination: &chain,
			Usage:       "way of communication with the block chain",
		},
	}
	app.Commands = []cli.Command{
		cmd.SignCommand,
		cmd.PublishCommand,
	}

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}
