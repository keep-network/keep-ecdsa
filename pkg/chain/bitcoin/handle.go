package bitcoin

// Handle serves as an interface abstraction around bitcoin network queries
type Handle interface {
	Broadcast(transaction string) error
	VbyteFee() (int32, error)
	IsAddressUnused(btcAddress string) (bool, error)
}
