package main

import (
	. "test_go_socket"
	"net"
	"fmt"
	"strconv"
	"time"
	"flag"
)

var closeSessionSignal chan int
var hideConsoleMsg bool

func init() {
	closeSessionSignal = make(chan int)
}

func main() {

	// 命令行相关：
	max := flag.Int("max", 2, "模拟最大连接数")
	delay := flag.Int("delay", 10, "延迟启动下一个连接（毫秒）")
	ip := flag.String("ip", "127.0.0.1", "服务器IP")
	hidemsg := flag.Bool("hidemsg", false, "隐藏日志")
	flag.Parse()

	maxClient := *max
	addr := *ip + ":10911"
	hideConsoleMsg = *hidemsg
	// fmt.Println("模拟最大连接数：", maxClient)

	for i:=0; i<maxClient; i++ {
		go socketClient(addr)
		// 等一秒
		time.Sleep(time.Millisecond * time.Duration(*delay))
	}

	for {
		tmp := <-closeSessionSignal
		maxClient -= tmp
		fmt.Println("收到一个closesession信号：", maxClient)

		// 如果客户端全部关闭了，那么进程结束
		if maxClient <= 0 {
			break
		}
	}

}
func socketClient(addr string) {

	conn, err := net.DialTimeout("tcp", addr, time.Second * 30)
	if err != nil {
		fmt.Println("无法连接到服务器：", addr)
		closeSessionSignal <- 1
		return
	}
	fmt.Println("连接服务器成功：", addr)

	clientClosed := make(chan bool)
	session := NewSession(conn)
	count := 0
	freq := time.Second * 1

	go func() {
		session.Recv(func(data []byte) bool {
			if !hideConsoleMsg {
				fmt.Println("接收到服务器反馈：", string(data), "session-id:", session.GetID())
			}
			return true
		})

		// 断开链接后
		clientClosed <- true
	}()

	running := true
	for running {
		select {
			case <-clientClosed:
				running = false
			default:
				session.SendMsg("hello:" + strconv.Itoa(count))
				count++
				time.Sleep(freq)
		}
	}

	fmt.Println("断开链接了。")
	closeSessionSignal <- 1
}
