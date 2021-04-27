package tbtc

import "github.com/keep-network/keep-ecdsa/pkg/chain/bitcoin"

// Config stores configuration of application extensions responsible for
// executing signer actions specific for TBTC application.
type Config struct {
	TBTCSystem string
	Bitcoin    bitcoin.Config
}
