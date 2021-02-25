package time

import (
	"time"
)

// Duration is a work around to be able to parse values provided in more
// friendly way, e.g. "4m20s". We have to do this because we use
// BurntSushi/toml package to parse configuration files. Unfortunately it
// doesn't support time.Duration out of the box.
type Duration struct {
	time.Duration
}

// UnmarshalText deserializes the text byte array into a native Duration
func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

// ToDuration converts our custom Duration to the equivalent `time.Duration`
func (d *Duration) ToDuration() time.Duration {
	return time.Duration(d.Duration)
}
