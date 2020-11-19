package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/keep-network/keep-common/pkg/logging"
	"github.com/keep-network/keep-core/pkg/net"
	"github.com/keep-network/keep-core/pkg/net/key"
	"github.com/keep-network/keep-core/pkg/net/local"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss"

	eth "github.com/keep-network/keep-ecdsa/pkg/chain"

	"github.com/ethereum/go-ethereum/common"

	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-ecdsa/internal/config"
	"github.com/keep-network/keep-ecdsa/pkg/registry"
	"github.com/urfave/cli"
)

// SigningCommand contains the definition of the `signing` command-line
// subcommand and its own subcommands.
var SigningCommand cli.Command

func init() {
	SigningCommand = cli.Command{
		Name:  "signing",
		Usage: "Provides several tools useful for out-of-band signing",
		Before: func(c *cli.Context) error {
			// disable the regular logger
			_ = logging.Configure("keep*=fatal")
			return nil
		},
		Subcommands: []cli.Command{
			{
				Name:      "decrypt-key-share",
				Usage:     "Decrypts the key share of the operator for the given keep",
				ArgsUsage: "[keep-address]",
				Action:    DecryptKeyShare,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "output-file,o",
						Usage: "Output file for the decrypted key share",
					},
				},
			},
			{
				Name:      "sign-digest",
				Usage:     "Sign a given digest using provided key shares",
				Action:    SignDigest,
				ArgsUsage: "[unprefixed-hex-digest] [key-shares-dir]",
			},
			EthereumSigningCommand,
		},
	}
}

// DecryptKeyShare decrypt key shares for given keep using provided operator config.
func DecryptKeyShare(c *cli.Context) error {
	config, err := config.ReadConfig(c.GlobalString("config"))
	if err != nil {
		return fmt.Errorf("failed while reading config file: [%v]", err)
	}

	keepAddressHex := c.Args().First()
	if !common.IsHexAddress(keepAddressHex) {
		return fmt.Errorf("invalid keep address")
	}

	keepAddress := common.HexToAddress(keepAddressHex)

	handle, err := persistence.NewDiskHandle(config.Storage.DataDir)
	if err != nil {
		return fmt.Errorf(
			"failed while creating a storage disk handler: [%v]",
			err,
		)
	}

	persistence := persistence.NewEncryptedPersistence(
		handle,
		config.Ethereum.Account.KeyFilePassword,
	)

	keepRegistry := registry.NewKeepsRegistry(persistence)

	keepRegistry.LoadExistingKeeps()

	signer, err := keepRegistry.GetSigner(keepAddress)
	if err != nil {
		return fmt.Errorf(
			"no signers for keep [%s]: [%v]",
			keepAddress.String(),
			err,
		)
	}

	signerBytes, err := signer.Marshal()
	if err != nil {
		return fmt.Errorf(
			"failed to marshall signer for keep [%s]: [%v]",
			keepAddress.String(),
			err,
		)
	}

	return outputData(c, signerBytes, 0444) // store to read-only file
}

// SignDigest signs a given digest using key shares from the provided directory.
func SignDigest(c *cli.Context) error {
	digest := c.Args().First()
	if len(digest) == 0 {
		return fmt.Errorf("invalid digest")
	}

	keySharesDir := c.Args().Get(1)
	if len(keySharesDir) == 0 {
		return fmt.Errorf("invalid key shares directory name")
	}

	keySharesFiles, err := ioutil.ReadDir(keySharesDir)
	if err != nil {
		return fmt.Errorf(
			"could not read key shares directory: [%v]",
			err,
		)
	}

	signers := make([]tss.ThresholdSigner, len(keySharesFiles))
	networkProviders := make([]net.Provider, len(keySharesFiles))

	for i, keyShareFile := range keySharesFiles {
		keyShareBytes, err := ioutil.ReadFile(
			fmt.Sprintf("%s/%s", keySharesDir, keyShareFile.Name()),
		)
		if err != nil {
			return fmt.Errorf(
				"could not read key share file [%v]: [%v]",
				keyShareFile.Name(),
				err,
			)
		}

		var signer tss.ThresholdSigner
		err = signer.Unmarshal(keyShareBytes)
		if err != nil {
			return fmt.Errorf(
				"could not unmarshal signer from file [%v]: [%v]",
				keyShareFile.Name(),
				err,
			)
		}

		operatorPublicKey, err := signer.MemberID().PublicKey()
		if err != nil {
			return fmt.Errorf(
				"could not get operator public key: [%v]",
				err,
			)
		}

		networkKey := key.NetworkPublic(*operatorPublicKey)
		networkProvider := local.ConnectWithKey(&networkKey)

		signers[i] = signer
		networkProviders[i] = networkProvider
	}

	digestBytes, err := hex.DecodeString(digest)
	if err != nil {
		return fmt.Errorf("could not decode digest string: [%v]", err)
	}

	ctx, cancelCtx := context.WithTimeout(
		context.Background(),
		1*time.Minute,
	)
	defer cancelCtx()

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(signers))

	type signingOutcome struct {
		signerIndex int
		signature   *ecdsa.Signature
		err         error
	}

	signingOutcomesChannel := make(chan *signingOutcome, len(signers))

	for i := range signers {
		go func(signerIndex int) {
			defer waitGroup.Done()

			signature, err := signers[signerIndex].CalculateSignature(
				ctx,
				digestBytes,
				networkProviders[signerIndex],
			)

			signingOutcomesChannel <- &signingOutcome{
				signerIndex,
				signature,
				err,
			}
		}(i)
	}

	waitGroup.Wait()
	close(signingOutcomesChannel)

	signatures := make(map[string]int)

	for signingOutcome := range signingOutcomesChannel {
		if signingOutcome.err != nil {
			_, _ = fmt.Fprintf(
				os.Stderr,
				"signer with index [%v] returned an error: [%v]",
				signingOutcome.signerIndex,
				signingOutcome.err,
			)
			continue
		}

		signature := fmt.Sprintf(
			"%064s%064s",
			signingOutcome.signature.R.Text(16),
			signingOutcome.signature.S.Text(16),
		)
		signatures[signature]++
	}

	if len(signatures) != 1 {
		return fmt.Errorf(
			"signing failed; a single signature should be produced",
		)
	}

	for signature, signersCount := range signatures {
		if signersCount != len(signers) {
			return fmt.Errorf(
				"signing failed; all signers should support the signature",
			)
		}

		publicKey, err := eth.SerializePublicKey(signers[0].PublicKey())
		if err != nil {
			return err
		}
		fmt.Println(hex.EncodeToString(publicKey[:]), "\t", signature)
	}

	return nil
}

// If `output-file` flag is provided stores the output in a file.
// `fileMode` determines the access permission for the output file. Sample values:
// 	0444 - read-only for all
// 	0644 - readable for all, but writeable only for the user (owner)
func outputData(c *cli.Context, data []byte, fileMode os.FileMode) error {
	if outputFilePath := c.String("output-file"); len(outputFilePath) > 0 {
		err := ioutil.WriteFile(outputFilePath, data, fileMode)
		if err != nil {
			return fmt.Errorf(
				"failed to write output to a file [%s]: [%v]",
				outputFilePath,
				err,
			)
		}

		fmt.Printf("output stored to a file: %s", outputFilePath)
	} else {
		_, err := os.Stdout.Write(data)
		if err != nil {
			return fmt.Errorf(
				"could not write bytes to stdout: [%v]",
				err,
			)
		}
	}

	return nil
}
