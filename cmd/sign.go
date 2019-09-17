package cmd

import (
	crand "crypto/rand"
	"fmt"

	"github.com/keep-network/keep-tecdsa/pkg/sign"
	"github.com/urfave/cli"
)

// SignCommand contains the definition of the sign command-line subcommand.
var SignCommand cli.Command

const signDescription = `The sign command creates a local signer and calculates
   a signature over provided argument.`

func init() {
	SignCommand = cli.Command{
		Name:        "sign",
		Usage:       "Calculates a signature",
		Description: signDescription,
		Action:      Sign,
	}
}

// Sign creates a local signer, generates a public/private ECDSA key pair for them
// and calculates a signature over a provided CLI argument.
func Sign(c *cli.Context) error {
	arg := c.Args().First()

	privateKey, err := sign.GenerateKey(crand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate key: [%v]", err)
	}

	signer := sign.NewSigner(privateKey)

	logger.Debugf("generated public key:\nX: %x\nY: %x",
		signer.PublicKey().X,
		signer.PublicKey().Y,
	)

	signature, err := signer.CalculateSignature(crand.Reader, []byte(arg))
	if err != nil {
		return fmt.Errorf("failed to calculate signature: [%v]", err)
	}

	logger.Infof("calculated signature:\nR: %x\nS: %x", signature.R, signature.S)

	return nil
}
