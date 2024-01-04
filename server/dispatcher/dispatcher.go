package dispatcher

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/sjy-dv/bridger/protobuf/bridgerpb"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/grpc/metadata"
)

type DispatchAPI struct {
	//todo
	async sync.RWMutex
}

type DispatchContext struct {
	Paylod   []byte
	Domain   string
	Metadata *metadata.MD
}

type ResponseWriter struct {
	*bridgerpb.PayloadReceiver
}

var DMap = make(map[string]func(ctx DispatchContext) *ResponseWriter)

func (d *DispatchAPI) Register(domain string, handler func(ctx DispatchContext) *ResponseWriter, subName ...string) {
	d.async.Lock()
	fmt.Println(fmt.Sprintf("[bridger] bridge registered route: %v", domain))
	defer d.async.Unlock()
	DMap[domain] = handler
}

func (dtx *DispatchContext) ExtractMD(ctx context.Context) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		dtx.Metadata = &metadata.MD{}
		return
	}
	dtx.Metadata = &md
	return
}

func (dtx *DispatchContext) Marshal(v interface{}) ([]byte, error) {
	b, err := msgpack.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (dtx *DispatchContext) UnMarshal(v interface{}) error {
	err := msgpack.Unmarshal(dtx.Paylod, v)
	if err != nil {
		return err
	}
	return nil
}

func (dtx *DispatchContext) Bind(v interface{}) error {
	return dtx.UnMarshal(v)
}

func (dtx *DispatchContext) GetMetadata(key string) string {
	key = strings.ToLower(key)
	if len(dtx.Metadata.Get(key)) == 0 {
		return ""
	}
	return dtx.Metadata.Get(key)[0]
}

func (dtx *DispatchContext) Error(err error) *ResponseWriter {
	paylod := &bridgerpb.PayloadReceiver{
		Info: &bridgerpb.ErrorInfo{
			Domain: dtx.Domain,
			Reason: err.Error(),
		},
	}
	return &ResponseWriter{paylod}
}

func (dtx *DispatchContext) Reply(v interface{}) *ResponseWriter {
	payload, err := dtx.Marshal(v)
	if err != nil {
		return dtx.Error(err)
	}
	resultValue := &bridgerpb.PayloadReceiver{
		Payload: payload,
	}
	return &ResponseWriter{resultValue}
}
