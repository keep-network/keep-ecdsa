package node

import (
	"sync"
	"testing"
	"time"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/ipfs/go-log"
)

func TestTSSPreParamsPool(t *testing.T) {
	err := log.SetLogLevel("*", "DEBUG")
	if err != nil {
		t.Fatalf("logger initialization failed: [%v]", err)
	}

	poolSize := 5

	// Create new pool.
	tssPool := newTestPool(poolSize)

	// Initial pool pump.
	go tssPool.pumpPool()
	time.Sleep(100 * time.Millisecond)

	// Get entry from pool.
	result := tssPool.get()
	if result == nil {
		t.Errorf("result is nil")
	}
}

func TestTSSPreParamsPoolEmpty(t *testing.T) {
	poolSize := 1

	// Create new pool.
	tssPool := newTestPool(poolSize)

	go func() {
		// Delay pumping so we have a chance to test if get function is waiting
		// for an entry.
		time.Sleep(100 * time.Millisecond)

		tssPool.pumpPool()
	}()

	// Get entry from pool.
	result := tssPool.get()
	if result == nil {
		t.Errorf("result is nil")
	}
}

func TestTSSPreParamsPoolConcurrent(t *testing.T) {
	poolSize := 5

	// Create new pool.
	tssPool := newTestPool(poolSize)

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(2)

	go func() {
		if result := tssPool.get(); result == nil {
			t.Errorf("result is nil")
		}
		waitGroup.Done()
	}()
	go func() {
		if result := tssPool.get(); result == nil {
			t.Errorf("result is nil")
		}
		waitGroup.Done()
	}()

	time.Sleep(100 * time.Millisecond)

	go tssPool.pumpPool()

	waitGroup.Wait()
}

func newTestPool(poolSize int) *tssPreParamsPool {
	return &tssPreParamsPool{
		pool: make(chan *keygen.LocalPreParams, poolSize),
		new: func() (*keygen.LocalPreParams, error) {
			time.Sleep(10 * time.Millisecond)
			return &keygen.LocalPreParams{}, nil
		},
	}
}
