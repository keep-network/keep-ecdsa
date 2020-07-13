package time

import (
	"time"
)

// We use BurntSushi/toml package to parse configuration file. Unfortunately it
// doesn't support time.Duration out of the box. Here we introduce a workaround
// to be able to parse values provided in more friendly way, e.g. "4m20s".
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

func (d *Duration) ToDuration() time.Duration {
	return time.Duration(d.Duration)
}
