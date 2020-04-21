package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/ipfs/go-log/v2"
	"github.com/keep-network/keep-common/pkg/logging"
	"github.com/keep-network/keep-ecdsa/cmd"
	"github.com/urfave/cli"
)

var logger = log.Logger("keep-main")

const (
	defaultConfigPath = "./configs/config.toml"
)

var (
	configPath string
)

func main() {
	err := logging.Configure(os.Getenv("LOG_LEVEL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to configure logging: [%v]\n", err)
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
	}
	app.Commands = []cli.Command{
		cmd.StartCommand,
	}

	err = app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}
