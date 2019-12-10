package tecdsa

import (
	"sync"
	"testing"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
)

func TestGenerate(t *testing.T) {
	tssPool := &tssPreParamsPool{
		poolMutex: &sync.Mutex{},
		pool: []*keygen.LocalPreParams{
			&keygen.LocalPreParams{},
		},
	}

	startLen := len(tssPool.pool)

	result := tssPool.get()
	if result == nil {
		t.Errorf("result is nil")
	}

	if len(tssPool.pool) != startLen-1 {
		t.Errorf("invalid length")
	}
}
