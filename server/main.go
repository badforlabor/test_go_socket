package main

import (
	"net"
	"sync"
	"sync/atomic"
	"comm/log"
	"encoding/binary"
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
func (s *Session) Close() error {
	s.conn.Close()
	return nil
}


func main() {
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
		session := &Session{conn:conn, sessionId:atomic.AddInt32(&gSessionId, 1),
							headerBuf: make([]byte, HeaderSizeOf)}
		go recv(session)
	}
}
func recv(session* Session) {
	defer session.Close()
	for {
		_, err := session.conn.Read(session.headerBuf)
		if err != nil {
			log.Errorf("read header error:%s,sessionId:%d", err, session.sessionId)
			return
		}
		len := binary.LittleEndian.Uint16(session.headerBuf)
		if len > MaxPacketSize || len < HeaderSizeOf {
			log.Errorf("协议长度不对：%d", err, session.sessionId)
			return
		}

		bodyBuf := make([]byte, int(len)-HeaderSizeOf)
		if _, err := session.conn.Read(bodyBuf); err != nil {
			log.Errorf("read body error:%s,sessionId:%d", err, session.sessionId)
			return
		}

		// 读取到协议体之后，原样返回
		send(session, bodyBuf)
	}
}
func send(session* Session, bodyBuf []byte) {
	bodySize := len(bodyBuf)

	if bodySize > MaxPacketSize - HeaderSizeOf {
		log.Errorf("包体太大")
		return
	}

	// 组织包体
	tmpHeaderBuf := make([]byte, HeaderSizeOf)
	binary.LittleEndian.PutUint16(tmpHeaderBuf, uint16(bodySize))

	session.sendLock.Lock()
	defer session.sendLock.Unlock()


	if _, err := session.conn.Write(tmpHeaderBuf); err != nil {
		log.Errorf("Write send tmpHeaderBuf err %s", err.Error())
		return
	}
	if _, err := session.conn.Write(bodyBuf); err != nil {
		log.Errorf("Write send msgBuf err %s", err.Error())
		return
	}

}
func sendMsg(session* Session, str string) {
	send(session, []byte(str))
}