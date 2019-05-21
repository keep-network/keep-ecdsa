package cmd

import (
	"fmt"

	"github.com/keep-network/keep-tecdsa/internal/config"
	"github.com/keep-network/keep-tecdsa/tests/smoketest"
	"github.com/urfave/cli"
)

// SmokeTestCommand contains the definition of the smoke test command-line
// subcommand.
var SmokeTestCommand cli.Command

// TODO: Update
const smokeTestDescription = `The smoke test command executes a smoke test to 
validate registration on ECDSA events emission.`

func init() {
	SmokeTestCommand = cli.Command{
		Name:        "smoke-test",
		Usage:       "Execute a smoke test",
		Description: smokeTestDescription,
		Action:      SmokeTest,
	}
}

// SmokeTest executes a smoke test.
func SmokeTest(c *cli.Context) error {
	configFile, err := config.ReadConfig(c.GlobalString("config"))
	if err != nil {
		return err
	}

	if err := smoketest.Execute(&configFile.Ethereum); err != nil {
		return fmt.Errorf("smoke test failed: [%s]", err)
	}

	return nil
}
