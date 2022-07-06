package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

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
	messages   chan message
	done       chan struct{}
}

type room struct {
	sema     chan any
	messages chan message
	toEnter  chan client
	toLeave  chan client

	listener net.Listener
	clients  map[client]bool
	close    chan any
}

func NewRoom(roomSize int) (r room, err error) {
	r = room{}
	r.listener, err = net.Listen("tcp", ":0")
	if err != nil {
		return
	}
	r.clients = make(map[client]bool, roomSize)
	r.sema = make(chan any, roomSize)
	r.messages = make(chan message)
	r.toEnter = make(chan client)
	r.toLeave = make(chan client)
	r.close = make(chan any)
	return
}

func (r *room) GetPort() int {
	return r.listener.Addr().(*net.TCPAddr).Port
}

func (r *room) Open() {
	go r.roomMonitor()

	for {
		conn, err := r.listener.Accept()
		if err != nil {
			select {
			case <-r.close:
				return
			default:
				continue
			}
		}
		go r.handleConn(conn)
	}
}

func (r *room) roomMonitor() {
	for {
		select {
		case msg := <-r.messages:
			for cl := range r.clients {
				if cl.name != msg.sender {
					cl.messages <- msg
				}
			}
		case cl := <-r.toEnter:
			r.clients[cl] = true
			enterMsg := message{text: cl.name + " joined"}
			for cli := range r.clients {
				cli.messages <- enterMsg
			}

		case cl := <-r.toLeave:
			close(cl.messages)
			delete(r.clients, cl)

			leaveMsg := message{text: cl.name + " left"}
			for cli := range r.clients {
				cli.messages <- leaveMsg
			}

			if len(r.clients) == 0 {
				log.Printf("room at port %d is empty, closing it", r.GetPort())
				err := r.listener.Close()
				if err != nil {
					log.Println(err)
				}
				close(r.close)
				return
			}
		}
	}
}

func (r *room) handleConn(conn net.Conn) {
	r.sema <- struct{}{}
	defer func() { <-r.sema }()
	defer conn.Close()

	log.Printf("new unnamed connection in room %d\n", r.GetPort())
	input := bufio.NewScanner(conn)
	cl := client{}
	cl.addr = conn.RemoteAddr().String()
	cl.messages = make(chan message)
	cl.done = make(chan struct{})
	input.Scan()
	cl.name = input.Text()

	log.Printf("new unnamed connection in room %d is %s", r.GetPort(), cl.name)
	go r.messageWriter(conn, cl)

	r.toEnter <- cl
	for input.Scan() {
		r.messages <- message{sender: cl.name, text: input.Text()}
	}
	close(cl.done)
	r.toLeave <- cl
}

func (r *room) messageWriter(conn net.Conn, cl client) {
	for {
		select {
		case msg := <-cl.messages:
			fmt.Fprintln(conn, msg.String())
		case <-cl.done:
			return
		}
	}
}
