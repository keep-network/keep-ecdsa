package tss

import (
	"testing"

	"github.com/keep-network/keep-core/pkg/operator"
)

func TestMemberID(t *testing.T) {
	_, publicKey, err := operator.GenerateKeyPair()
	if err != nil {
		t.Fatalf("could not generate public key: [%v]", err)
	}

	memberID := MemberIDFromPublicKey(publicKey)

	extractedPublicKey, err := memberID.PublicKey()
	if err != nil {
		t.Fatalf("could not extract public key: [%v]", err)
	}

	if !MemberIDFromPublicKey(extractedPublicKey).Equal(memberID) {
		t.Errorf("member from extracted public key doesn't match the original member")
	}

	memberIDString := memberID.String()

	memberIDFromString, err := MemberIDFromString(memberIDString)
	if err != nil {
		t.Fatalf("could not construct member from string: [%v]", err)
	}

	if !memberIDFromString.Equal(memberID) {
		t.Errorf("member from string doesn't match the original member")
	}
}
