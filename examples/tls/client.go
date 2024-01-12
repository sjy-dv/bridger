package main

import (
	"log"
	"time"

	"github.com/sjy-dv/bridger/client"
	"github.com/sjy-dv/bridger/client/options"
	"google.golang.org/grpc/credentials"
)

/*
If you are connecting to a pure gRPC server using SSL in Ingress,

	simply set Enable to true without any additional configuration.
*/
func main() {
	/*
		not tls
	*/
	// bridgerClient := client.RegisterBridgerClient(&options.Options{
	// 	Addr:           "127.0.0.1:50051",
	// 	MinChannelSize: 1,
	// 	MaxChannelSize: 4,
	// 	Timeout:        time.Duration(time.Second * 5),
	// })
	// defer bridgerClient.Close()
	/*
		using ca.cert
	*/
	// b, _ := os.ReadFile("ca.cert")
	// cp := x509.NewCertPool()
	// if !cp.AppendCertsFromPEM(b) {
	// 	panic("credentials: failed to append certificates")
	// }
	// bridgerClient := client.RegisterBridgerClient(&options.Options{
	// 	Addr:           "127.0.0.1:50051",
	// 	MinChannelSize: 1,
	// 	MaxChannelSize: 4,
	// 	Timeout:        time.Duration(time.Second * 5),
	// 	Credentials: options.Credentials{
	// 		Enable: true,
	// 		TLS: &tls.Config{
	// 			InsecureSkipVerify: false,
	// 			RootCAs:            cp,
	// 		},
	// 	},
	// })
	// defer bridgerClient.Close()
	/*
		using pem
	*/
	creds, err := credentials.NewClientTLSFromFile("service.pem", "")
	if err != nil {
		log.Fatalf("could not process the credentials: %v", err)
	}
	bridgerClient := client.RegisterBridgerClient(&options.Options{
		Addr:           "127.0.0.1:50051",
		MinChannelSize: 1,
		MaxChannelSize: 4,
		Timeout:        time.Duration(time.Second * 5),
		Credentials: options.Credentials{
			Enable: true,
			Cred:   creds,
		},
	})
	defer bridgerClient.Close()
}
