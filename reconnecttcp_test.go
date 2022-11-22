package reconnecttcp

import (
	"fmt"
	"testing"
	"time"
)

//func TestReConnection(t *testing.T) {
//	addr := "127.0.0.1:5000"
//	client := NewReconnectTcp(addr)
//	go func(r *ReconnectTcp) {
//		for data := range r.Read() {
//			fmt.Printf("client recv:%s\n", data)
//		}
//	}(client)
//
//	go func(r *ReconnectTcp) {
//		for {
//			r.Write([]byte("message from client"))
//			time.Sleep(time.Millisecond * 10)
//		}
//	}(client)
//
//	l, err := net.Listen("tcp", addr)
//	if err != nil {
//		t.Fatalf("listen %s fail:%v\n", addr, err)
//		return
//	}
//
//	go func(listener net.Listener) {
//		for i := 0; i < 10; i++ {
//			conn, err := l.Accept()
//			if err != nil {
//				fmt.Printf("accept fail:%v\n", err)
//				continue
//			}
//
//			conn.Write([]byte("hello" + strconv.Itoa(i)))
//			time.Sleep(time.Millisecond * 100)
//			conn.Close()
//		}
//	}(l)
//
//	time.Sleep(time.Second * 5)
//
//}

func TestReConnectionNew(t *testing.T) {
	addr := "127.0.0.1:5001"
	client := NewReconnectTcpNew(addr)
	go func() {
		data := client.Read()
		if data != nil {
			fmt.Printf("接收:%s\n", string(data))
		}
		time.Sleep(200 * time.Millisecond)
	}()
	for {
		fmt.Println("请输入要发送的内容：")
		var str string
		fmt.Scanln(&str)
		client.Write([]byte(str))
	}

}
