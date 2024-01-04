package options

import "time"

type Options struct {
	Addr           string
	MinChannelSize int
	MaxChannelSize int
	Timeout        time.Duration
}
