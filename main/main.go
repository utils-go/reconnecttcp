package main

import (
	"fmt"
	"github.com/utils-go/reconnecttcp"
	"time"
)

func main() {
	addr := "127.0.0.1:5001"
	client := reconnecttcp.NewReconnectTcp(addr)
	go func() {
		for {
			data := client.Read()
			if data != nil {
				fmt.Printf("接收:%s\n", string(data))
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()
	for {
		fmt.Println("请输入要发送的内容：")
		var str string
		fmt.Scanln(&str)
		client.Write([]byte(str))
	}
}
