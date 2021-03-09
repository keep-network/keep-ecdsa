package eth

// Config stores app-specific extensions configuration.
type Config struct {
	TBTC TBTC
}

// TBTC stores configuration of application extension responsible for
// executing signer actions specific for TBTC application.
type TBTC struct {
	TBTCSystem string
	BTCRefunds BTCRefunds
}

// BTCRefunds stores configuration related to recovering BTC from a closed keep.
type BTCRefunds struct {
	BeneficiaryAddress string
	MaxFeePerVByte     int32
}
