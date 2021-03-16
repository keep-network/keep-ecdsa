package bitcoin

// Config stores configuration related to recovering BTC from a closed keep.
type Config struct {
	BeneficiaryAddress string
	MaxFeePerVByte     int32
}
