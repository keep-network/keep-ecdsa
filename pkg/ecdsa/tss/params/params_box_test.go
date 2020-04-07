package params

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
)

func TestGetContent(t *testing.T) {
	params := &keygen.LocalPreParams{
		P: big.NewInt(1),
		Q: big.NewInt(2),
	}
	box := NewBox(params)

	content, err := box.Content()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(params, content) {
		t.Fatalf(
			"unexpected content\nexpected: [%v]\nactual:   [%v]",
			params,
			content,
		)
	}
}

func TestGetDestroyedContent(t *testing.T) {
	params := &keygen.LocalPreParams{
		P: big.NewInt(1),
		Q: big.NewInt(2),
	}
	box := NewBox(params)

	box.DestroyContent()
	_, err := box.Content()

	expectedError := fmt.Errorf("box is empty")
	if !reflect.DeepEqual(expectedError, err) {
		t.Fatalf(
			"unexpected error\nexpected: [%v]\nactual:   [%v]",
			expectedError,
			err,
		)
	}
}

func TestIsEmpty(t *testing.T) {
	params := &keygen.LocalPreParams{
		P: big.NewInt(1),
		Q: big.NewInt(2),
	}
	box := NewBox(params)

	if box.IsEmpty() {
		t.Fatal("box should not be empty")
	}

	box.DestroyContent()

	if !box.IsEmpty() {
		t.Fatal("box should be empty")
	}
}
