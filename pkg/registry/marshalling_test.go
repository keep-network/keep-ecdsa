package registry

import (
	crand "crypto/rand"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa"
	"github.com/keep-network/keep-tecdsa/pkg/utils/pbutils"
)

func TestMembershipRoundtrip(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(crand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer := ecdsa.NewSigner(privateKey)
	keepAddress := common.HexToAddress("0x6312d9689665DAB22E21b11B6fDf86547E566288")

	membership := &Membership{
		KeepAddress: keepAddress,
		Signer:      signer,
	}

	unmarshaled := &Membership{}

	err = pbutils.RoundTrip(membership, unmarshaled)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(membership, unmarshaled) {
		t.Fatalf("unexpected content of unmarshaled membership")
	}
}
