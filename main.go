package main

import (
	goLog "log"
	"os"
	"path"
	"time"

	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-tecdsa/cmd"
	"github.com/urfave/cli"
)

var logger = log.Logger("keep-main")

const (
	defaultConfigPath   = "./configs/config.toml"
	defaultBroadcastAPI = "blockcypher"
)

var (
	configPath   string
	broadcastAPI string
)

func main() {
	err := setUpLogging(os.Getenv("LOG_LEVEL"))
	if err != nil {
		goLog.Fatal(err)
	}

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
			Name:        "broadcast-api",
			Value:       defaultBroadcastAPI,
			Destination: &broadcastAPI,
			Usage:       "external service used to communicate with the blockchain",
		},
	}
	app.Commands = []cli.Command{
		cmd.SignCommand,
		cmd.StartCommand,
		cmd.PublishCommand,
		cmd.SmokeTestCommand,
	}

	err = app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}
