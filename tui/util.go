package tui

import (
	"fmt"
	"github.com/dimaglushkov/go-chat/chat"
	"github.com/rivo/tview"
	"net"
	"strconv"
	"time"
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

type grpcReq struct {
	chat.RoomNameSize
	useSize bool
}

func (r *grpcReq) validate() bool {
	if len(r.Name) == 0 {
		return false
	}
	if r.useSize && r.Size == 0 {
		return false
	}
	return true
}

func center(width, height int, p tview.Primitive) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}

func spinner() {
	for {
		for _, r := range `-\|/` {
			fmt.Printf("\r%c", r)
			time.Sleep(100 * time.Millisecond)
		}
	}
}
