# Bridger

![logo](./docs/logo.png)

Bridger is a framework designed to support microservices architecture, enabling developers to easily build and manage microservices. This framework focuses on reducing the complexity of constructing and integrating microservices by allowing services to be registered and used in a manner similar to REST APIs. It provides a simplified approach to microservice development and management, streamlining the process of service deployment and communication within a distributed system.


Bridger primarily operates on the principle of gRPC multiplexing, which typically necessitates only a single connection. However, in certain situations, a single connection may prove insufficient, or might risk overloading the target server. To mitigate this, Bridger consistently offers a pool for establishing additional connections.

This approach ensures that if Server A becomes overloaded and Server A2 is deployed for load balancing, the traffic is continuously directed to Server A alone, due to its singleton nature. This strategy prevents the consistent channeling of traffic to a single server, thus efficiently managing server loads and distributing traffic more effectively.


1. Use `go get` to install the latest version of the Bridger Client and Sever dependencies:

   ```shell
   go get -u github.com/sjy-dv/bridger@latest
   ```

2. Client Example:

```go
import (
	"log"
	"time"

	"github.com/sjy-dv/bridger/client"
	"github.com/sjy-dv/bridger/client/options"
)

func main() {
	/**
	default value
	if you want to singleton instance,
	min&max channel size should be set 1
	*/
	bridgerClient := client.RegisterBridgerClient(&options.Options{
		Addr:           "127.0.0.1:50051",
		MinChannelSize: 1,
		MaxChannelSize: 4,
		Timeout:        time.Duration(time.Second * 5),
	})
	defer bridgerClient.Close()
	type req struct {
		Msg string
	}
	val, err := bridgerClient.Dispatch("/greetings", &req{Msg: "Hello, Dispatcher"})
	if err != nil {
		panic(err)
	}
	response := &req{}
	err = client.Unmarshal(val, response)
	if err != nil {
		panic(err)
	}
	log.Println("First Message : ", response.Msg)
    // you can easily use metadata
	header := client.MetadataHeader{}
	header["name"] = "gopher"
	val, err = bridgerClient.Dispatch("/greetings/withname", &req{Msg: "I'm gopher"}, client.CallOptions{
		MetadataHeader: header,
	})
	if err != nil {
		panic(err)
	}
	response = &req{}
	err = client.Unmarshal(val, response)
	if err != nil {
		panic(err)
	}
	log.Println("Second Message : ", response.Msg)
}
```

3. Server Example:

```go
import (
	"github.com/sjy-dv/bridger/server"
	"github.com/sjy-dv/bridger/server/dispatcher"
	"github.com/sjy-dv/bridger/server/options"
)

func main() {
	bridger := server.New()
    // you can use register function like rest api
	bridger.Register("/greetings", greetings)
	bridger.Register("/greetings/withname",
		greetingsWithHeaderName,
		"is using metadata api")
	bridger.RegisterBridgerServer(&options.Options{
		Port:                         50051,
		ChainUnaryInterceptorLogger:  true,
		ChainStreamInterceptorLogger: true,
	})
}

func greetings(dtx dispatcher.DispatchContext) *dispatcher.ResponseWriter {
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
	req.Msg = req.Msg + "\n" + "Me too.."
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
```

4. Overview Options

| ClientOptions | Explanation |
| ------ | ------ |
| Addr | Enter the address of the server to connect to. (Example: localhost:50051) |
| MinChannelSize | Specify the number of clients to connect. Typically, gRPC provides multiplexing, so one client is sufficient unless there is a special or high amount of traffic. (Default: 1) |
| MaxChannelSize | Specify the maximum number of clients to connect. For smooth communication and handling of traffic requests, Bridger by default operates only 100 concurrent sessions per client. (Default: 4) |
| Timeout | Specify the maximum duration for maintaining communication connections. The default value is the same as the disconnection time of the grpc Client. |
| MaxRecvMsgSize | Set the maximum size for messages that can be received. This should be the same as the maximum message size sent by the server. (Default: 10mb) |
| MaxSendMsgSize | Set the maximum size for messages that can be sent. This should be the same as the maximum message size that the server can receive. (Default: 10mb) |

| ServerOptions | Explanation |
| ------ | ------ |
| Port | This is the port number to run on. There is no default value, and it must be entered. |
| ChainUnaryInterceptorLogger | Activate the ChainUnaryInterceptorLogger. (Default: false) |
| ChainStreamInterceptorLogger | Activate the ChainStreamInterceptorLogger. (Default: false) |
| MaxRecvMsgSize | Set the maximum size for messages that can be received. This should be the same as the maximum message size sent by the client. (Default: 10mb) |
| MaxSendMsgSize | Set the maximum size for messages that can be sent. This should be the same as the maximum message size that the client can receive. (Default: 10mb) |