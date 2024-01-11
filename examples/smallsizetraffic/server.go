package main

import (
	"log"
	"sync/atomic"

	"github.com/sjy-dv/bridger/server"
	"github.com/sjy-dv/bridger/server/dispatcher"
	"github.com/sjy-dv/bridger/server/options"
)

var at *atomic.Int64

func main() {
	at = &atomic.Int64{}
	bridger := server.New()

	bridger.Register("/auth_calculate", AuthCalCulate)
	bridger.RegisterBridgerServer(&options.Options{
		Port:                         50051,
		ChainUnaryInterceptorLogger:  false,
		ChainStreamInterceptorLogger: false,
	})
}

func AuthCalCulate(dtx dispatcher.DispatchContext) *dispatcher.ResponseWriter {
	var (
		req = struct {
			Index1 string
			Index2 string
			Index3 string
			Index4 int
			Index5 int
			Index6 int
			Index7 bool
			Index8 string
		}{}
		err error
	)
	err = dtx.Bind(&req)
	if err != nil {
		return dtx.Error(err)
	}
	req.Index1 = req.Index1 + "idx1"
	req.Index2 = req.Index2 + "idx2"
	req.Index3 = req.Index3 + "idx3"
	req.Index4 = req.Index4*12 ^ 2
	req.Index5 = req.Index5 ^ 2/5
	req.Index6 = req.Index6*3 ^ 2/7 + 102
	req.Index7 = true
	req.Index8 = dtx.GetMetadata("TestAuthHeader") + "authok"
	at.Add(1)
	log.Println(">>>>>>>>>>", at.Load())
	return dtx.Reply(&req)
}

func greetingsWithHeaderName(dtx dispatcher.DispatchContext) *dispatcher.ResponseWriter {
	var (
		req = struct {
			Msg string
		}{}
		err error
	)
	err = dtx.Bind(&req)
	if err != nil {
		return dtx.Error(err)
	}
	name := dtx.GetMetadata("name")
	req.Msg = "Hello " + name
	return dtx.Reply(&req)
}
