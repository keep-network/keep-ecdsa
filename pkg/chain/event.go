package chain

// BondedECDSAKeepCreatedEvent is an event emitted on a new keep creation.
type BondedECDSAKeepCreatedEvent struct {
	Keep                 BondedECDSAKeepHandle
	MemberIDs            []ID // keep member ids
	HonestThreshold      uint64
	BlockNumber          uint64
	ThisOperatorIsMember bool
}

// ConflictingPublicKeySubmittedEvent is an event emitted each time when one of
// the members of a keep has submitted a key that does not match the keys submitted
// so far by other members.
type ConflictingPublicKeySubmittedEvent struct {
	SubmittingMember     ID
	ConflictingPublicKey []byte
	BlockNumber          uint64
}

// PublicKeyPublishedEvent is an event emitted once all the members have submitted
// the same public key and it was accepted by keep as its public key.
type PublicKeyPublishedEvent struct {
	PublicKey   []byte
	BlockNumber uint64
}

// SignatureRequestedEvent is an event emitted when a user requests
// a digest to be signed.
type SignatureRequestedEvent struct {
	Digest      [32]byte
	BlockNumber uint64
}

// KeepClosedEvent is an event emitted when a keep has been closed.
type KeepClosedEvent struct {
	BlockNumber uint64
}

// KeepTerminatedEvent is an event emitted when a keep has been terminated.
type KeepTerminatedEvent struct {
	BlockNumber uint64
}

// SignatureSubmittedEvent is an event emitted when a keep submits a signature.
type SignatureSubmittedEvent struct {
	Digest      [32]byte
	R           [32]byte
	S           [32]byte
	RecoveryID  uint8
	BlockNumber uint64
}
