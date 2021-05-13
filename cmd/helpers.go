package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/keep-network/keep-common/pkg/persistence"
	"github.com/keep-network/keep-core/pkg/operator"
	"github.com/keep-network/keep-ecdsa/pkg/chain"
)

func nodeHeader(addrStrings []string, port int) {
	header := ` 

▓▓▌ ▓▓ ▐▓▓ ▓▓▓▓▓▓▓▓▓▓▌▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▄
▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▌▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ▓▓▓▓▓▓    ▓▓▓▓▓▓▓▀    ▐▓▓▓▓▓▓    ▐▓▓▓▓▓   ▓▓▓▓▓▓     ▓▓▓▓▓   ▐▓▓▓▓▓▌   ▐▓▓▓▓▓▓
  ▓▓▓▓▓▓▄▄▓▓▓▓▓▓▓▀      ▐▓▓▓▓▓▓▄▄▄▄         ▓▓▓▓▓▓▄▄▄▄         ▐▓▓▓▓▓▌   ▐▓▓▓▓▓▓
  ▓▓▓▓▓▓▓▓▓▓▓▓▓▀        ▐▓▓▓▓▓▓▓▓▓▓         ▓▓▓▓▓▓▓▓▓▓▌        ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ▓▓▓▓▓▓▀▀▓▓▓▓▓▓▄       ▐▓▓▓▓▓▓▀▀▀▀         ▓▓▓▓▓▓▀▀▀▀         ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▀
  ▓▓▓▓▓▓   ▀▓▓▓▓▓▓▄     ▐▓▓▓▓▓▓     ▓▓▓▓▓   ▓▓▓▓▓▓     ▓▓▓▓▓   ▐▓▓▓▓▓▌
▓▓▓▓▓▓▓▓▓▓ █▓▓▓▓▓▓▓▓▓ ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓
▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓ ▐▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓

Trust math, not hardware.
	
`

	prefix := "| "
	suffix := " |"

	maxLineLength := len(strconv.Itoa(port))

	for _, addrString := range addrStrings {
		if addrLength := len(addrString); addrLength > maxLineLength {
			maxLineLength = addrLength
		}
	}

	maxLineLength += len(prefix) + len(suffix) + 6
	dashes := strings.Repeat("-", maxLineLength)

	fmt.Printf(
		"%s%s\n%s\n%s\n%s\n%s%s\n\n",
		header,
		dashes,
		buildLine(maxLineLength, prefix, suffix, "Keep ECDSA Node"),
		buildLine(maxLineLength, prefix, suffix, ""),
		buildLine(maxLineLength, prefix, suffix, fmt.Sprintf("Port: %d", port)),
		buildMultiLine(maxLineLength, prefix, suffix, "IPs : ", addrStrings),
		dashes,
	)
}

func buildLine(lineLength int, prefix, suffix string, internalContent string) string {
	contentLength := len(prefix) + len(suffix) + len(internalContent)
	padding := lineLength - contentLength

	return fmt.Sprintf(
		"%s%s%s%s",
		prefix,
		internalContent,
		strings.Repeat(" ", padding),
		suffix,
	)
}

func buildMultiLine(lineLength int, prefix, suffix, startPrefix string, lines []string) string {
	combinedLines := buildLine(lineLength, prefix+startPrefix, suffix, lines[0]) + "\n"

	startPadding := strings.Repeat(" ", len(startPrefix))
	for _, line := range lines[1:] {
		combinedLines += buildLine(lineLength, prefix+startPadding, suffix, line) + "\n"
	}

	return combinedLines
}

func buildPersistenceHandle(
	chainHandle chain.OfflineHandle,
	keyFilePassword string,
	dataDir string,
) (persistence.Handle, error) {
	// Validate chain name to avoid issues with persistence later.
	validChainName, err := regexp.MatchString(
		"^[a-z][a-z0-9-_]*$",
		chainHandle.Name(),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to verify chain name [%v]: [%v]",
			chainHandle.Name(),
			err,
		)
	}
	if !validChainName {
		return nil, fmt.Errorf(
			"invalid chain name: [%v]; chain name must start with a lowercase "+
				"letter and then consist solely of lowercase letters, numbers, "+
				" -, or _",
			chainHandle.Name(),
		)
	}

	// Below, use the bare data dir for the Ethereum chain for backwards
	// compatibility. A future version may do a one-time migration of the
	// ethereum directory.
	//
	// For other chains, use the chain's self-reported name as a path prefix
	// within the data directory. Since all directories in the Ethereum path are
	// Ethereum addresses, the validation above requiring a starting letter
	// ensures there will be no clashes with existing Ethereum address
	// directories.
	diskPersistencePath := dataDir
	if chainHandle.Name() != "ethereum" {
		diskPersistencePath += "/" + strings.ToLower(chainHandle.Name())
	}
	handle, err := persistence.NewDiskHandle(diskPersistencePath)
	if err != nil {
		return nil, fmt.Errorf(
			"failed while creating a storage disk handler: [%v]",
			err,
		)
	}

	return persistence.NewEncryptedPersistence(
		handle,
		keyFilePassword,
	), nil
}

type operatorKeys struct {
	public  *operator.PublicKey
	private *operator.PrivateKey
}
