package app

import (
	"bufio"
	"errors"
	"log"
	"net"
	"strconv"
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

func sendMsg(sender *bufio.Writer, msg string) error {
	if sender == nil {
		return errors.New("app.msgSender is nil")
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
