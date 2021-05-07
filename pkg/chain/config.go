package chain

import "github.com/keep-network/keep-ecdsa/pkg/chain/bitcoin"

// TBTC stores configuration of application extension responsible for
// executing signer actions specific for TBTC application.
type Config struct {
	TBTCSystem string
	ElectrsURL *string
	BTCRefunds bitcoin.Config
}

// ElectrsURLWithDefault dereferences ElectrsURL in the following way: if there
// is a configured value, use it. Otherwise, default to
// https://blockstream.info/api/. This allows us to add bitcoin connection
// functionality to nodes that haven't made config changes yet while also
// letting a user connect to the node of their choice.
func (c Config) ElectrsURLWithDefault() string {
	if c.ElectrsURL == nil {
		return "https://blockstream.info/api/"
	}
	return *c.ElectrsURL
}
