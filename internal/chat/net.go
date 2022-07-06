package chat

import (
	"bufio"
	"errors"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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

func grpcConnector(addr, port string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(addr+":"+port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second*3))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func tcpConnector(addr, port string) (*net.TCPConn, error) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr+":"+port)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)

	if err != nil {
		return nil, err
	}
	return conn, nil
}

func sendMsg(sender *bufio.Writer, msg string) error {
	if sender == nil {
		return errors.New("chat.msgSender is nil")
	}
	if len(msg) == 0 {
		return errors.New("msg is empty")
	}
	if msg[len(msg)-1] != '\n' {
		msg += "\n"
	}
	_, err := sender.WriteString(msg)
	if err != nil {
		return err
	}
	err = sender.Flush()
	return err
}

func receiveMsg(receiver *bufio.Scanner, printer func(string), done chan<- struct{}) {
	if receiver == nil {
		log.Println("receiver is nil")
		return
	}
	for receiver.Scan() {
		msgText := receiver.Text()
		printer(msgText)
	}
	done <- struct{}{}
}
