package app

import (
	"net"
	"strconv"
)

type chatConn struct {
	net.Conn
	done chan any
}

func (cc chatConn) Ch() chan any {
	return cc.done
}

type ChatConn interface {
	net.Conn
	Ch() chan any
}

func NewChatConn(con net.Conn) ChatConn {
	return &chatConn{con, make(chan any)}
}

type serverAddr struct {
	ipAddr, port string
}

func (addr serverAddr) validate() bool {
	if ip := net.ParseIP(addr.ipAddr); ip == nil {
		return false
	}
	if _, err := strconv.ParseInt(addr.port, 10, 32); err != nil {
		return false
	}
	return true
}
