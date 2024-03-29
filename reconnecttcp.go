package reconnecttcp

import (
	"fmt"
	"github.com/utils-go/concurrentlist"
	"net"
	"sync"
	"time"
)

//可重连的tcp.发送接收可靠

type ReconnectTcp struct {
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
	//标识是否关闭
	chClose chan struct{}
}

func NewReconnectTcp(ipStr string) *ReconnectTcp {
	t := &ReconnectTcp{
		wBuffer:    make(chan []byte, 100),
		chWriteErr: make(chan struct{}, 5),
		rBuffer:    concurrentlist.NewListT[[]byte](),
		chReadErr:  make(chan struct{}, 5),
		ipStr:      ipStr,
		chClose:    make(chan struct{}),
	}
	go t.initConnect()
	return t
}
func (t *ReconnectTcp) handleRead(wg *sync.WaitGroup) {
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
func (t *ReconnectTcp) handWrite(wg *sync.WaitGroup) {
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
func (t *ReconnectTcp) initConnect() {
	for {
		var err error
		t.con, err = net.Dial("tcp", t.ipStr)
		if err != nil {
			time.Sleep(time.Millisecond * 100)
			fmt.Printf("[initConnect] fail %v,ipStr:%s\n", err, t.ipStr)
			continue
		}
		fmt.Printf("connect %s success\n", t.ipStr)
		t.wBuffer = make(chan []byte, 100)
		//连接上了，开启读写线程
		var wg sync.WaitGroup
		wg.Add(2)
		go t.handleRead(&wg)
		go t.handWrite(&wg)
		wg.Wait()
		//这里结束，可能是接收错误，也可能是发送错误，更有可能是人为关闭了ReconnectTcpNew
		select {
		case <-t.chClose:
			return
		default:
			break
		}
	}
}

func (t *ReconnectTcp) Write(data []byte) {
	t.wBuffer <- data
}
func (t *ReconnectTcp) Read() []byte {
	buffer := t.rBuffer.TakeAll()
	if len(buffer) <= 0 {
		return nil
	}
	r := make([]byte, 0, 10240)
	n := 0
	for _, bytes := range buffer {
		r = append(r, bytes...)
		n += len(bytes)
	}
	return r[0:n]
}
func (t *ReconnectTcp) Close() {
	//关闭连接,会触发read 和 write线程停止
	if t.con != nil {
		t.con.Close()
	}
	t.chClose <- struct{}{}
}
