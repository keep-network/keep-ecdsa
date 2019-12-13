package tecdsa

import (
	"fmt"
	"sync"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss"
)

type tssPreParamsPool struct {
	poolMutex *sync.Mutex
	pool      []*keygen.LocalPreParams
}

// GenerateTSSPreParams generates TSS pre-parameters and stores them in a pool.
func (t *TECDSA) GenerateTSSPreParams() {
	poolSize := 2

	for i := 0; i < poolSize; i++ {
		err := t.tssParamsPool.generateNew()
		if err != nil {
			logger.Warningf("failed to generate new tss pre-parameters: [%]", err)
		}
	}
}

func (t *tssPreParamsPool) generateNew() error {
	params, err := tss.GenerateTSSPreParams()
	if err != nil {
		return fmt.Errorf("failed to generate tss pre parameters: [%v]", err)
	}

	t.poolMutex.Lock()
	defer t.poolMutex.Unlock()

	t.pool = append(t.pool, params)

	return nil
}

func (t *tssPreParamsPool) get() *keygen.LocalPreParams {
	var params *keygen.LocalPreParams

	if len(t.pool) > 0 {
		t.poolMutex.Lock()
		defer t.poolMutex.Unlock()

		params = t.pool[0]
		t.pool = t.pool[1:len(t.pool)]
	}

	go func() {
		if err := t.generateNew(); err != nil {
			logger.Errorf("failed to generate new tss pre parameters", err)
		}
	}()

	return params
}
