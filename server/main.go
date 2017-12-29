package main

import (
	"net"
	. "test_go_socket"
)


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
		session := NewSession(conn)
		go process(session)
	}
}
func process(session* Session) {
	session.Recv(func(data []byte) {
		session.Send(data)
	})
}
