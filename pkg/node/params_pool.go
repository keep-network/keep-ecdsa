package node

import (
	"time"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/keep-network/keep-ecdsa/pkg/ecdsa/tss"
)

// tssPreParamsPool is a pool holding TSS pre parameters. It autogenerates entries
// up to the pool size. When an entry is pulled from the pool it will generate
// new entry.
type tssPreParamsPool struct {
	pool chan *keygen.LocalPreParams
	new  func() (*keygen.LocalPreParams, error)
}

// InitializeTSSPreParamsPool generates TSS pre-parameters and stores them in a pool.
func (n *Node) InitializeTSSPreParamsPool() {
	poolSize := n.tssConfig.GetPreParamsTargetPoolSize()

	logger.Infof("TSS pre-parameters target pool size is [%v]", poolSize)

	n.tssParamsPool = &tssPreParamsPool{
		pool: make(chan *keygen.LocalPreParams, poolSize),
		new: func() (*keygen.LocalPreParams, error) {
			return tss.GenerateTSSPreParams(
				n.tssConfig.GetPreParamsGenerationTimeout(),
			)
		},
	}

	go n.tssParamsPool.pumpPool()
}

// TSSPreParamsPoolSize returns the current size of the TSS params pool.
func (n *Node) TSSPreParamsPoolSize() int {
	if n.tssParamsPool == nil {
		return 0
	}

	return len(n.tssParamsPool.pool)
}

func (t *tssPreParamsPool) pumpPool() {
	for {
		logger.Info("generating new tss pre parameters")

		start := time.Now()

		params, err := t.new()
		if err != nil {
			logger.Warningf(
				"failed to generate tss pre parameters after [%s]: [%v]",
				time.Since(start),
				err,
			)
			continue
		}

		logger.Infof(
			"generated new tss pre parameters, took: [%s], current pool size: [%d]",
			time.Since(start),
			len(t.pool)+1,
		)

		t.pool <- params
	}
}

// get returns TSS pre parameters from the pool. It pumps the pool after getting
// and entry. If the pool is empty it will wait for a new entry to be generated.
func (t *tssPreParamsPool) get() *keygen.LocalPreParams {
	return <-t.pool
}
