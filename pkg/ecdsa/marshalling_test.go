package ecdsa

import (
	crand "crypto/rand"
	"reflect"
	"testing"

	"github.com/keep-network/keep-tecdsa/pkg/utils/pbutils"
)

func TestSignerRoundtrip(t *testing.T) {
	privateKey, err := GenerateKey(crand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	signer := NewSigner(privateKey)

	unmarshaled := &Signer{}

	err = pbutils.RoundTrip(signer, unmarshaled)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(signer, unmarshaled) {
		t.Fatalf("unexpected content of unmarshaled signer")
	}
}
