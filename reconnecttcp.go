package reconnecttcp

import (
	"fmt"
	"net"
	"time"
)

//可重连的tcp.发送接收可靠

type ReconnectTcp struct {
	//写入的内容
	wBuffer chan []byte
	//接收的内容
	rBuffer chan []byte
	//tcp连接
	con net.Conn
	//ip地址与端口
	ipStr string
	//发生错误事件
	errEvent chan interface{}
	//标识是否关闭
	isClose bool
}

func NewReconnectTcp(ipStr string) *ReconnectTcp {
	t := &ReconnectTcp{
		wBuffer:  make(chan []byte, 100),
		rBuffer:  make(chan []byte, 100),
		ipStr:    ipStr,
		errEvent: make(chan interface{}, 1),
		isClose:  false,
	}
	t.initConnect()
	//断线重连和写入
	go func(t *ReconnectTcp) {
		var ok bool
		var err error
		var data []byte
		for {
			if t.isClose {
				break
			}

			select {
			case _, ok = <-t.errEvent:
				//不ok，说明channel关闭了
				if !ok {
					break
				}
				if t.con != nil {
					t.con.Close()
				}
				t.initConnect()
			case data, ok = <-t.wBuffer:
				//不ok，说明channel关闭了
				if !ok {
					break
				}

				if t.con == nil {
					//fmt.Println("【write】t.con为空")
					time.Sleep(time.Millisecond * 100)
					//为空，则忽略
					continue
				}
				_, err = t.con.Write(data)
				if err != nil {
					//写入失败，则重连
					t.errEvent <- struct{}{}
				}
			}
		}
	}(t)
	//读取
	go func(t *ReconnectTcp) {
		var n int
		var err error
		for {
			if t.isClose {
				break
			}
			if t.con == nil {
				//fmt.Println("【read】t.con为空")
				time.Sleep(time.Millisecond * 100)
				continue
			}
			buffer := make([]byte, 10240)
			n, err = t.con.Read(buffer)
			if err != nil {
				//错误时，休眠100ms
				time.Sleep(time.Millisecond * 100)
				continue
			}
			t.rBuffer <- buffer[:n]
		}
	}(t)
	return t
}
func (t *ReconnectTcp) initConnect() {
	var err error
	t.con, err = net.Dial("tcp", t.ipStr)
	if err != nil {
		time.Sleep(time.Millisecond * 100)
		t.errEvent <- struct{}{}
		fmt.Printf("[initConnect] fail %v,ipStr:%s\n", err, t.ipStr)
	}
}

func (t *ReconnectTcp) Write(data []byte) {
	t.wBuffer <- data
}
func (t *ReconnectTcp) Read() chan []byte {
	return t.rBuffer
}
func (t *ReconnectTcp) Close() {
	t.isClose = true
	//关闭连接
	if t.con != nil {
		t.con.Close()
	}
	//关闭读取buffer，使外部退出循环
	close(t.rBuffer)
}
