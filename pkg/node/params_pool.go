package node

import (
	"sync"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/keep-network/keep-tecdsa/pkg/ecdsa/tss"
)

// tssPreParamsPool is a pool holding TSS pre parameters. It autogenerates entries
// up to the pool size. When an entry is pulled from the pool it will generate
// new entry.
type tssPreParamsPool struct {
	pumpFuncMutex *sync.Mutex // lock concurrent executions of pumping function

	paramsMutex *sync.Cond
	params      []*keygen.LocalPreParams

	new func() (*keygen.LocalPreParams, error)

	poolSize int
}

// InitializeTSSPreParamsPool generates TSS pre-parameters and stores them in a pool.
func (n *Node) InitializeTSSPreParamsPool() {
	n.tssParamsPool = &tssPreParamsPool{
		pumpFuncMutex: &sync.Mutex{},
		paramsMutex:   sync.NewCond(&sync.Mutex{}),
		params:        []*keygen.LocalPreParams{},
		poolSize:      2,
		new: func() (*keygen.LocalPreParams, error) {
			return tss.GenerateTSSPreParams()
		},
	}

	go n.tssParamsPool.pumpPool()
}

func (n *tssPreParamsPool) pumpPool() {
	n.pumpFuncMutex.Lock()
	defer n.pumpFuncMutex.Unlock()

	for {
		if len(n.params) >= n.poolSize {
			logger.Debugf("tss pre parameters pool is pumped")
			return
		}

		params, err := n.new()
		if err != nil {
			logger.Warningf("failed to generate tss pre parameters: [%v]", err)
			continue
		}

		n.paramsMutex.L.Lock()
		n.params = append(n.params, params)
		n.paramsMutex.Signal()
		n.paramsMutex.L.Unlock()

		logger.Debugf("generated new tss pre parameters")
	}
}

// get returns TSS pre parameters from the pool. It pumps the pool after getting
// and entry. If the pool is empty it will wait for a new entry to be generated.
func (n *tssPreParamsPool) get() *keygen.LocalPreParams {
	n.paramsMutex.L.Lock()
	defer n.paramsMutex.L.Unlock()

	for len(n.params) == 0 {
		n.paramsMutex.Wait()
	}

	params := n.params[0]
	n.params = n.params[1:len(n.params)]

	go n.pumpPool()

	return params
}
