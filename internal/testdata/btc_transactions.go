// Package testdata contains transactions data used in tests.
package testdata

// Transaction holds transaction specific data.
type Transaction struct {
	Hash                 string // hash of the transaction
	PreviousTxHash       string // hash of a previous transaction
	PreviousOutScript    string // script sig of a previous transaction output
	PreviousOutAmount    int64  // amount of a previous transaction output
	UnsignedRaw          string // serialized transaction before signing
	WitnessSignatureHash string // witness signature hash for signing
	SignedRaw            string // serialized transaction after signing
}

var (
	// InitialTx is a prerequisite transaction used in tests to initialize a
	// chain. Next transaction is build based on this one.
	InitialTx = Transaction{
		Hash:      "5e13ca34cf527e7b443afc0d6958a67bf7950a11f6ec3e05f8e3f3e802fbdf99",
		SignedRaw: "0100000001f24d19b6980927dbe47c30fd13b1cc12e56a11cc019efed67a1b4d3937b74bab010000006a47304402201711a033c1b829719716c81419294214a7fce0f0f1f9f51b6821ca3a5beebbdd022059b7bdd0bf1fe08aa4b4654360732d2a1f97c602b2e198a41e7bc53d81376c9a0121028896955d043b5a43957b21901f2cce9f0bfb484531b03ad6cd3153e45e73ee2effffffff022823000000000000160014d849b1e1cede2ac7d7188cf8700e97d6975c91c4b2f9fd00000000001976a914d849b1e1cede2ac7d7188cf8700e97d6975c91c488ac00000000",
	}

	// ValidTx is a transaction build on an initial transaction.
	ValidTx = Transaction{
		Hash:                 "ec367c260ead9e3c91583175f35382e22b66df6d59fd0aac175bb36519b664f7",
		PreviousTxHash:       InitialTx.Hash,
		PreviousOutScript:    "0014d849b1e1cede2ac7d7188cf8700e97d6975c91c4",
		PreviousOutAmount:    int64(9000),
		UnsignedRaw:          "010000000199dffb02e8f3e3f8053eecf6110a95f77ba658690dfc3a447b7e52cf34ca135e0000000000ffffffff02581b000000000000160014d849b1e1cede2ac7d7188cf8700e97d6975c91c4e8030000000000001976a914d849b1e1cede2ac7d7188cf8700e97d6975c91c488ac00000000",
		WitnessSignatureHash: "cc493a708e6ec962f2be8dc0a24c35966ee46f563de8bf219b9c5313a3b24e58",
		SignedRaw:            "0100000000010199dffb02e8f3e3f8053eecf6110a95f77ba658690dfc3a447b7e52cf34ca135e0000000000ffffffff02581b000000000000160014d849b1e1cede2ac7d7188cf8700e97d6975c91c4e8030000000000001976a914d849b1e1cede2ac7d7188cf8700e97d6975c91c488ac02483045022100ecadce07f5c9d84b4fa1b2728806135acd81ad9398c9673eeb4e161d42364b92022076849daa2108ed2a135d16eb9e103c5819db014ea2bad5c92f4aeecf47bf9ac80121028896955d043b5a43957b21901f2cce9f0bfb484531b03ad6cd3153e45e73ee2e00000000",
	}
)
