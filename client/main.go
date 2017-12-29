package main

import (
	. "test_go_socket"
	"net"
	"fmt"
	"strconv"
	"time"
)

var closeSessionSignal chan int

func init() {
	closeSessionSignal = make(chan int)
}

func main() {

	maxClient := 2
	addr := ":10911"

	for i:=0; i<maxClient; i++ {
		go socketClient(addr)
		// 等一秒
		time.Sleep(time.Second)
	}

	for {
		tmp := <-closeSessionSignal
		maxClient -= tmp

		// 如果客户端全部关闭了，那么进程结束
		if tmp <= 0 {
			break
		}
	}

}
func socketClient(addr string) {

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}

	clientClosed := make(chan bool)
	session := NewSession(conn)
	count := 0
	freq := time.Second * 1

	go func() {
		session.Recv(func(data []byte) {
			fmt.Println("接收到服务器反馈：", string(data))
		})

		// 断开链接后
		closeSessionSignal <- 1
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
}
