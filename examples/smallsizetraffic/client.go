package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sjy-dv/bridger/client"
	"github.com/sjy-dv/bridger/client/options"
	"github.com/teris-io/shortid"
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
	type req = struct {
		Index1 string
		Index2 string
		Index3 string
		Index4 int
		Index5 int
		Index6 int
		Index7 bool
		Index8 string
	}
	var group sync.WaitGroup
	var wg sync.WaitGroup
	var wg2 sync.WaitGroup

	group.Add(2)
	wg.Add(250)
	wg2.Add(300)
	atomic := &atomic.Int64{}
	go func() {
		defer group.Done()
		for i := 0; i < 250; i++ {
			go func(num int) {
				defer wg.Done()
				index1 := shortid.MustGenerate()
				index2 := shortid.MustGenerate()
				index3 := shortid.MustGenerate()
				authHeader := client.MetadataHeader{}
				expect4 := (num*2)*12 ^ 2
				expect5 := (num * 3) ^ 2/5
				expect6 := (num*4)*3 ^ 2/7 + 102
				headerVal := shortid.MustGenerate()
				authHeader["TestAuthHeader"] = headerVal
				val, err := bridgerClient.Dispatch("/auth_calculate", &req{
					Index1: index1,
					Index2: index2,
					Index3: index3,
					Index4: num * 2,
					Index5: num * 3,
					Index6: num * 4,
					Index7: false,
				}, client.CallOptions{
					MetadataHeader: authHeader,
				})
				if err != nil {
					panic(err)
				}
				response := &req{}
				err = client.Unmarshal(val, response)
				if err != nil {
					panic(err)
				}
				if response.Index1 != index1+"idx1" {
					panic(fmt.Sprintf("unmatched!! %s, %s", response.Index1, index1))
				}
				if response.Index2 != index2+"idx2" {
					panic(fmt.Sprintf("unmatched!! %s, %s", response.Index2, index2))
				}
				if response.Index3 != index3+"idx3" {
					panic(fmt.Sprintf("unmatched!! %s, %s", response.Index3, index3))
				}
				if response.Index4 != expect4 {
					panic(fmt.Sprintf("unmatched!! %v, %v", response.Index4, expect4))
				}
				if response.Index5 != expect5 {
					panic(fmt.Sprintf("unmatched!! %v, %v", response.Index5, expect5))
				}
				if response.Index6 != expect6 {
					panic(fmt.Sprintf("unmatched!! %v, %v", response.Index6, expect6))
				}
				if response.Index7 != true {
					panic(fmt.Sprintf("unmatched!! %v, %v", response.Index7, true))
				}
				if response.Index8 != headerVal+"authok" {
					panic(fmt.Sprintf("unmatched!! %s, %s", response.Index8, headerVal+"authok"))
				}
				atomic.Add(1)
			}(i)
		}
	}()
	go func() {
		defer group.Done()
		for i := 0; i < 300; i++ {
			go func(num int) {
				defer wg2.Done()
				index1 := shortid.MustGenerate()
				index2 := shortid.MustGenerate()
				index3 := shortid.MustGenerate()
				authHeader := client.MetadataHeader{}
				expect4 := (num*2)*12 ^ 2
				expect5 := (num * 3) ^ 2/5
				expect6 := (num*4)*3 ^ 2/7 + 102
				headerVal := shortid.MustGenerate()
				authHeader["TestAuthHeader"] = headerVal
				val, err := bridgerClient.Dispatch("/auth_calculate", &req{
					Index1: index1,
					Index2: index2,
					Index3: index3,
					Index4: num * 2,
					Index5: num * 3,
					Index6: num * 4,
					Index7: false,
				}, client.CallOptions{
					MetadataHeader: authHeader,
				})
				if err != nil {
					panic(err)
				}
				response := &req{}
				err = client.Unmarshal(val, response)
				if err != nil {
					panic(err)
				}
				if response.Index1 != index1+"idx1" {
					panic(fmt.Sprintf("unmatched!! %s, %s", response.Index1, index1))
				}
				if response.Index2 != index2+"idx2" {
					panic(fmt.Sprintf("unmatched!! %s, %s", response.Index2, index2))
				}
				if response.Index3 != index3+"idx3" {
					panic(fmt.Sprintf("unmatched!! %s, %s", response.Index3, index3))
				}
				if response.Index4 != expect4 {
					panic(fmt.Sprintf("unmatched!! %v, %v", response.Index4, expect4))
				}
				if response.Index5 != expect5 {
					panic(fmt.Sprintf("unmatched!! %v, %v", response.Index5, expect5))
				}
				if response.Index6 != expect6 {
					panic(fmt.Sprintf("unmatched!! %v, %v", response.Index6, expect6))
				}
				if response.Index7 != true {
					panic(fmt.Sprintf("unmatched!! %v, %v", response.Index7, true))
				}
				if response.Index8 != headerVal+"authok" {
					panic(fmt.Sprintf("unmatched!! %s, %s", response.Index8, headerVal+"authok"))
				}
				atomic.Add(1)
			}(i)
		}
	}()
	group.Wait()
	wg.Wait()
	wg2.Wait()
	fmt.Println("all celear", atomic.Load())
	// val, err := bridgerClient.Dispatch("/greetings", &req{Msg: "Hello, Dispatcher"})
	// if err != nil {
	// 	panic(err)
	// }
	// response := &req{}
	// err = client.Unmarshal(val, response)
	// if err != nil {
	// 	panic(err)
	// }
	// log.Println("First Message : ", response.Msg)

	// header := client.MetadataHeader{}
	// header["name"] = "gopher"
	// val, err = bridgerClient.Dispatch("/greetings/withname", &req{Msg: "I'm gopher"}, client.CallOptions{
	// 	MetadataHeader: header,
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// response = &req{}
	// err = client.Unmarshal(val, response)
	// if err != nil {
	// 	panic(err)
	// }
	// log.Println("Second Message : ", response.Msg)
}
