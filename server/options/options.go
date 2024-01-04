package options

type Options struct {
	Port                         int
	ChainUnaryInterceptorLogger  bool
	ChainStreamInterceptorLogger bool
	MaxRecvMsgSize               int
	MaxSendMsgSize               int
}

const (
	b  = 1
	kb = 1024
	mb = 1024 * 1024
	gb = 1024 * 1024 * 1024

	B  = 1
	KB = 1024
	MB = 1024 * 1024
	GB = 1024 * 1024 * 1024
)
