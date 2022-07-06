package server

import (
	"bufio"
	"errors"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dimaglushkov/go-chat/api/butlerpb"
)

func connectToRoom(rp *butlerpb.RoomPort) (*net.TCPConn, error) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", "localhost:"+strconv.FormatInt(int64(rp.Port), 10))
	return net.DialTCP("tcp", nil, tcpAddr)
}

func sendMsg(sender *bufio.Writer, msg string) error {
	time.Sleep(time.Second / 2)

	if sender == nil {
		return errors.New("sender is nil")
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
	time.Sleep(time.Second / 2)
	return err
}

func TestRoom_Open(t *testing.T) {
	var done = make(chan struct{})
	r, err := NewRoom(10)
	require.NoError(t, err)

	go func() {
		r.Open()
		done <- struct{}{}
	}()

	conn, err := connectToRoom(&butlerpb.RoomPort{Port: int32(r.GetPort())})
	require.NoError(t, err)

	w := bufio.NewWriter(conn)

	err = sendMsg(w, "my_name")
	require.NoError(t, err)

	conn.Close()
	err = sendMsg(w, "test")
	require.Error(t, err)
	<-done
}
