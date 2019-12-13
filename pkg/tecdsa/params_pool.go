package tecdsa

import (
	"sync"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss"
)

// TSSPreParamsPool is a pool holding TSS pre parameters. It autogenerates entries
// up to the pool size. When an entry is pulled from the pool it will generate
// new entry.
type TSSPreParamsPool struct {
	pumpFuncMutex *sync.Mutex // lock concurrent executions of pumping function

	paramsMutex *sync.Mutex
	params      []*keygen.LocalPreParams

	new func() (*keygen.LocalPreParams, error)

	poolSize int
}

// InitializeTSSPreParamsPool generates TSS pre-parameters and stores them in a pool.
func (t *TECDSA) InitializeTSSPreParamsPool() {
	t.tssParamsPool = &TSSPreParamsPool{
		pumpFuncMutex: &sync.Mutex{},
		paramsMutex:   &sync.Mutex{},
		params:        []*keygen.LocalPreParams{},
		poolSize:      2,
		new: func() (*keygen.LocalPreParams, error) {
			return tss.GenerateTSSPreParams()
		},
	}

	go t.tssParamsPool.pumpPool()
}

func (t *TSSPreParamsPool) pumpPool() {
	t.pumpFuncMutex.Lock()
	defer t.pumpFuncMutex.Unlock()

	for {
		if len(t.params) >= t.poolSize {
			logger.Debugf("tss pre parameters pool is pumped")
			return
		}

		params, err := t.new()
		if err != nil {
			logger.Warningf("failed to generate tss pre parameters: [%v]", err)
			return
		}

		t.paramsMutex.Lock()
		t.params = append(t.params, params)
		t.paramsMutex.Unlock()

		logger.Debugf("generated new tss pre parameters")
	}
}

// Get returns TSS pre parameters from the pool. It pumps the pool after getting
// and entry. If no entries were found in the pool it return nil.
func (t *TSSPreParamsPool) Get() *keygen.LocalPreParams {
	var params *keygen.LocalPreParams

	t.paramsMutex.Lock()
	defer t.paramsMutex.Unlock()

	if len(t.params) > 0 {
		params = t.params[0]
		t.params = t.params[1:len(t.params)]
	}

	go t.pumpPool()

	return params
}
