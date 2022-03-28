package chat

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

func newRoom(roomSize int) (r room, err error) {
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

func (r *room) getPort() int {
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
				log.Print(err)
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
			delete(r.clients, cl)
			close(cl.messages)

			leaveMsg := message{text: cl.name + " left"}
			for cli := range r.clients {
				cli.messages <- leaveMsg
			}

			if len(r.clients) == 0 {
				close(r.close)
				err := r.listener.Close()
				if err != nil {
					log.Println(err)
				}
				return
			}
		}
	}
}

func (r *room) handleConn(conn net.Conn) {
	r.sema <- struct{}{}
	defer func() { <-r.sema }()
	defer conn.Close()

	input := bufio.NewScanner(conn)
	cl := client{}
	cl.addr = conn.RemoteAddr().String()
	cl.messages = make(chan message)
	input.Scan()
	cl.name = input.Text()

	go r.messageWriter(conn, cl)

	r.toEnter <- cl
	for input.Scan() {
		r.messages <- message{sender: cl.name, text: input.Text()}
	}
	r.toLeave <- cl
}

func (r *room) messageWriter(conn net.Conn, cl client) {
	for {
		select {
		case msg := <-cl.messages:
			fmt.Fprintln(conn, msg.String())
		case <-r.close:
			return
		}
	}
}
