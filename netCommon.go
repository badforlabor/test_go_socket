/*
	自己写的一个简单的网络协议处理库
	协议是以字符串形式存在的，格式是：head+body，其中head大小为2，body的大小最大为16384
*/


package test_go_socket

import (
	"sync"
	"net"
	"sync/atomic"
	"encoding/binary"
	"fmt"
)

const HeaderSizeOf = 2
const MaxPacketSize = 16384

var gSessionId int32 = 0

type Session struct {
	conn           net.Conn
	sessionId      int32
	headerBuf      []byte
	sendLock       sync.Mutex
}
func NewSession(conn net.Conn) *Session {
	return &Session {
			conn:conn,
			sessionId:    atomic.AddInt32(&gSessionId, 1),
			headerBuf:    make([]byte, HeaderSizeOf),
		}
}
func (s *Session) GetID() int32 {
	return s.sessionId
}

func (s *Session) Close() error {
	s.conn.Close()
	return nil
}
func (session *Session) Recv(callback func([]byte)bool) {
	defer session.Close()
	for {
		_, err := session.conn.Read(session.headerBuf)
		if err != nil {
			fmt.Printf("read header error:%s,sessionId:%d\n", err, session.sessionId)
			return
		}
		len := binary.LittleEndian.Uint16(session.headerBuf)
		if len > MaxPacketSize || len < HeaderSizeOf {
			fmt.Printf("协议长度不对：%d\n", err, session.sessionId)
			return
		}

		bodyBuf := make([]byte, int(len))
		if _, err := session.conn.Read(bodyBuf); err != nil {
			fmt.Printf("read body error:%s,sessionId:%d\n", err, session.sessionId)
			return
		}

		// 读取到协议体之后，告知回调函数。如果回调函数返回false，那么就不在Recv了。
		if !callback(bodyBuf) {
			break
		}
	}
}

// 发协议
func (session* Session) SendMsg(str string) {
	session.Send([]byte(str))
}

// 发二进制协议
func (session* Session) Send(bodyBuf []byte) {
	bodySize := len(bodyBuf)

	if bodySize > MaxPacketSize - HeaderSizeOf {
		fmt.Printf("包体太大\n")
		return
	}

	// 组织包体
	tmpHeaderBuf := make([]byte, HeaderSizeOf)
	binary.LittleEndian.PutUint16(tmpHeaderBuf, uint16(bodySize))

	session.sendLock.Lock()
	defer session.sendLock.Unlock()


	if _, err := session.conn.Write(tmpHeaderBuf); err != nil {
		fmt.Printf("Write send tmpHeaderBuf err %s\n", err.Error())
		return
	}
	if _, err := session.conn.Write(bodyBuf); err != nil {
		fmt.Printf("Write send msgBuf err %s\n", err.Error())
		return
	}
	// fmt.Println("send message succ.")
}
