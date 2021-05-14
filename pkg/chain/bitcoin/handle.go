package bitcoin

// Handle serves as an interface abstraction around bitcoin network queries
type Handle interface {
	Broadcast(transaction string) error
	VbyteFee() (int32, error)
	IsAddressUnused(btcAddress string) (bool, error)
}

type OfflineHandle struct{}

// Broadcast logs a transaction as a warning five times in a row so that any
// warning monitors would pick it up and alert on it.
func (oh OfflineHandle) Broadcast(transaction string) error {
	for i := 0; i < 5; i++ {
		logger.Warningf("Please broadcast Bitcoin transaction %s", transaction)
	}
	return nil
}

// VbyteFee returns a default fee of 75 in lieu of a network to pull a current
// fee from.
func (oh OfflineHandle) VbyteFee() (int32, error) {
	return 75, nil
}

// IsAddressUnused always returns true in lieu of a way to check whether or not
// an address is used.
func (oh OfflineHandle) IsAddressUnused(btcAddress string) (bool, error) {
	return true, nil
}
