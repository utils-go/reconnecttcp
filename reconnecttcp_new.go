package reconnecttcp

import (
	"fmt"
	"github.com/utils-go/concurrentlist"
	"net"
	"sync"
	"time"
)

//可重连的tcp.发送接收可靠

type ReconnectTcpNew struct {
	//写入的内容
	wBuffer chan []byte
	//写入错误的channel
	chWriteErr chan struct{}
	//接收的内容
	rBuffer *concurrentlist.ConcurrentListT[[]byte]
	//读取错误的channel
	chReadErr chan struct{}
	//tcp连接
	con net.Conn
	//ip地址与端口
	ipStr string
	//发生错误事件
	reconnectChan chan struct{}
	//标识是否关闭
	chClose chan struct{}
}

func NewReconnectTcpNew(ipStr string) *ReconnectTcpNew {
	t := &ReconnectTcpNew{
		wBuffer:       make(chan []byte, 100),
		chWriteErr:    make(chan struct{}, 5),
		rBuffer:       concurrentlist.NewList[[]byte](),
		chReadErr:     make(chan struct{}, 5),
		ipStr:         ipStr,
		reconnectChan: make(chan struct{}, 2),
		chClose:       make(chan struct{}),
	}
	go t.initConnect()
	return t
}
func (t *ReconnectTcpNew) handleRead(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-t.chWriteErr:
			return
		default:
			break
		}
		buffer := make([]byte, 10240)
		n, err := t.con.Read(buffer)
		if err != nil {
			t.chReadErr <- struct{}{}
			return
		}
		t.rBuffer.Add(buffer[0:n])
	}
}
func (t *ReconnectTcpNew) handWrite(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case data, ok := <-t.wBuffer:
			if !ok {
				return
			}
			_, err := t.con.Write(data)
			if err != nil {
				t.chWriteErr <- struct{}{} //写入错误，退出
				return
			}
		case <-t.chReadErr:
			return
		}
	}
}
func (t *ReconnectTcpNew) initConnect() {
	for {
		select {
		case <-t.chClose:
			return
		case <-t.reconnectChan:
			//清空reconnectChan
			t.reconnectChan = make(chan struct{}, 2)
			break
		}
		var err error
		t.con, err = net.Dial("tcp", t.ipStr)
		if err != nil {
			time.Sleep(time.Millisecond * 100)
			t.reconnectChan <- struct{}{}
			fmt.Printf("[initConnect] fail %v,ipStr:%s\n", err, t.ipStr)
		}
		fmt.Printf("connect %s success\n", t.ipStr)
		t.wBuffer = make(chan []byte, 100)
		//连接上了，开启读写线程
		var wg sync.WaitGroup
		wg.Add(2)
		go t.handleRead(&wg)
		go t.handWrite(&wg)
		wg.Wait()
	}
}

func (t *ReconnectTcpNew) Write(data []byte) {
	t.wBuffer <- data
}
func (t *ReconnectTcpNew) Read() []byte {
	buffer := t.rBuffer.TakeAll()
	if len(buffer) <= 0 {
		return nil
	}
	r := make([]byte, 10240)
	n := 0
	for _, bytes := range buffer {
		r = append(r, bytes...)
		n += len(bytes)
	}
	return r[0:n]
}
func (t *ReconnectTcpNew) Close() {
	//关闭连接
	if t.con != nil {
		t.con.Close()
	}
	t.chClose <- struct{}{}
}
