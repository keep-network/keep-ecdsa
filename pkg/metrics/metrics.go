package metrics

import (
	"context"
	"time"

	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-ecdsa/pkg/client"

	"github.com/keep-network/keep-common/pkg/metrics"
)

var logger = log.Logger("keep-metrics")

const (
	// DefaultClientMetricsTick is the default duration of the
	// observation tick for client metrics.
	DefaultClientMetricsTick = 1 * time.Minute
)

// ObserveTSSPreParamsPoolSize triggers an observation process of the
// tss_pre_params_pool_size metric.
func ObserveTSSPreParamsPoolSize(
	ctx context.Context,
	registry *metrics.Registry,
	clientHandle *client.Handle,
	tick time.Duration,
) {
	input := func() float64 {
		return float64(clientHandle.TSSPreParamsPoolSize())
	}

	observe(
		ctx,
		"tss_pre_params_pool_size",
		input,
		registry,
		validateTick(tick, DefaultClientMetricsTick),
	)
}

func observe(
	ctx context.Context,
	name string,
	input metrics.ObserverInput,
	registry *metrics.Registry,
	tick time.Duration,
) {
	observer, err := registry.NewGaugeObserver(name, input)
	if err != nil {
		logger.Warningf("could not create gauge observer [%v]", name)
		return
	}

	observer.Observe(ctx, tick)
}

func validateTick(tick time.Duration, defaultTick time.Duration) time.Duration {
	if tick > 0 {
		return tick
	}

	return defaultTick
}
