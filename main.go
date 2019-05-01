package main

import (
	"log"
	"os"
	"path"
	"time"

	"github.com/keep-network/keep-tecdsa/cmd"
	"github.com/urfave/cli"
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

	app.Commands = []cli.Command{
		cmd.SignCommand,
	}

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}
