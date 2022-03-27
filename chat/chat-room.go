package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

const roomSize = 20

type message struct {
	sender, text string
}

func (msg message) String() string {
	if msg.sender != "" {
		return msg.sender + ": " + msg.text
	}
	return msg.text
}

type client struct {
	name, addr string
}

type chatRoom struct {
	sema     chan interface{}
	messages chan message
	toEnter  chan client
	toLeave  chan client

	listener net.Listener
	clients  map[client]chan message
}

func newChatRoom() (cr chatRoom, err error) {
	cr = chatRoom{}
	cr.listener, err = net.Listen("tcp", ":0")
	if err != nil {
		return
	}
	cr.clients = make(map[client]chan message, roomSize)
	cr.sema = make(chan interface{}, roomSize)
	cr.messages = make(chan message)
	cr.toEnter = make(chan client)
	cr.toLeave = make(chan client)
	return
}

func (cr *chatRoom) roomMonitor() {
	for {
		select {
		case msg := <-cr.messages:
			for cl := range cr.clients {
				if cl.name != msg.sender {
					cr.clients[cl] <- msg
				}
			}
		case cl := <-cr.toEnter:
			cr.clients[cl] = make(chan message)
			enterMsg := message{text: cl.name + " joined"}
			for cli := range cr.clients {
				cr.clients[cli] <- enterMsg
			}

		case cl := <-cr.toLeave:
			close(cr.clients[cl])
			delete(cr.clients, cl)
			leaveMsg := message{text: cl.name + " left"}
			for cli := range cr.clients {
				cr.clients[cli] <- leaveMsg
			}
		}
	}
}

func (cr *chatRoom) handleConn(conn net.Conn) {
	defer conn.Close()

	input := bufio.NewScanner(conn)
	cl := client{}
	cl.addr = conn.RemoteAddr().String()

	input.Scan()
	cl.name = input.Text()

	go cr.messageWriter(conn, cl)

	cr.toEnter <- cl
	for input.Scan() {
		cr.messages <- message{sender: cl.name, text: input.Text()}
	}
	cr.toLeave <- cl
}

func (cr *chatRoom) messageWriter(conn net.Conn, cl client) {
	for msg := range cr.clients[cl] {
		fmt.Fprintln(conn, msg.String())
	}
}

func (cr *chatRoom) start() {
	go cr.roomMonitor()

	for {
		conn, err := cr.listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go cr.handleConn(conn)
	}
}

func run() {
	cr, err := newChatRoom()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("starting chat room at port: ", cr.listener.Addr().(*net.TCPAddr).Port)
	cr.start()
}

func main() {
	run()
}
