package local

type localChainLogger struct {
	retrieveSignerPubkeyCalls       int
	provideRedemptionSignatureCalls int
	increaseRedemptionFeeCalls      int
	keepAddressCalls                int
}

func (lcl *localChainLogger) logRetrieveSignerPubkeyCall() {
	lcl.retrieveSignerPubkeyCalls++
}

func (lcl *localChainLogger) RetrieveSignerPubkeyCalls() int {
	return lcl.retrieveSignerPubkeyCalls
}

func (lcl *localChainLogger) logProvideRedemptionSignatureCall() {
	lcl.provideRedemptionSignatureCalls++
}

func (lcl *localChainLogger) ProvideRedemptionSignatureCalls() int {
	return lcl.provideRedemptionSignatureCalls
}

func (lcl *localChainLogger) logIncreaseRedemptionFeeCall() {
	lcl.increaseRedemptionFeeCalls++
}

func (lcl *localChainLogger) IncreaseRedemptionFeeCalls() int {
	return lcl.increaseRedemptionFeeCalls
}

func (lcl *localChainLogger) logKeepAddressCall() {
	lcl.keepAddressCalls++
}

func (lcl *localChainLogger) KeepAddressCalls() int {
	return lcl.keepAddressCalls
}
