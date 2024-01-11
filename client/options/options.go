package options

import "time"

type Options struct {
	Addr           string
	MinChannelSize int
	MaxChannelSize int
	Timeout        time.Duration
	MaxRecvMsgSize int
	MaxSendMsgSize int
}

const DefaultMsgSize = 104858000 // 10mb
