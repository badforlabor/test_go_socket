package main

import (
	"net"
	. "test_go_socket"
	"math/rand"
	"time"
	"sync/atomic"
	"fmt"
	"runtime"
)

// 随机关闭客户端
var randomCloseClientSeed = 0
var gCurrentConnCnt int32 = 0

func main() {
	runtime.GOMAXPROCS(8)

	// 计时器
	go timer()

	service := ":10911"
	ln, err := net.Listen("tcp", service)
	if err != nil {
		return
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				continue
			} else {
				return
			}
		}
		session := NewSession(conn)
		go process(session)
	}
}
func process(session* Session) {
	// 增加一个连接数目
	atomic.AddInt32(&gCurrentConnCnt, 1)

	session.Recv(func(data []byte) bool {
		session.Send(data)

		// 随机断开链接
		if randomCloseClientSeed > 0 && rand.Intn(randomCloseClientSeed) == randomCloseClientSeed - 1 {
			session.SendMsg("close.")
			return false
		}

		return true
	})

	// 减少一个连接数目
	atomic.AddInt32(&gCurrentConnCnt, -1)
}

func timer() {
	t1 := time.NewTimer(time.Second)
	for {
		select {
		case <- t1.C:
			// 每秒打印一下当前连接数目
			fmt.Println("当前有效连接数目为：", gCurrentConnCnt)

			t1.Reset(time.Second)
		}
	}
}