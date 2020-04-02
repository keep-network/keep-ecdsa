package tss

import (
	"time"
)

// Config contains configuration for tss protocol execution.
type Config struct {
	// Concurrency level for pre-parameters generation in tss-lib.
	PreParamsGenerationConcurrency int
	// Timeout for pre-parameters generation in tss-lib.
	PreParamsGenerationTimeout duration
}

// We use BurntSushi/toml package to parse configuration file. Unfortunately it
// doesn't support time.Duration out of the box. Here we introduce a workaround
// to be able to parse values provided in more friendly way, e.g. "4m20s".
type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
