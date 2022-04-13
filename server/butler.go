package server

import (
	"fmt"
	"github.com/dimaglushkov/go-chat/server/rpc"
	"golang.org/x/net/context"
	"log"
	"sync"
)

const maxRoomSize = 99

type Butler struct {
	rpc.ButlerServer
	mu    sync.RWMutex
	rooms map[string]int32
}

func NewButler() (butler Butler) {
	butler.rooms = make(map[string]int32)
	return
}

func (b *Butler) CreateRoom(ctx context.Context, roomNameSize *rpc.RoomNameSize) (*rpc.RoomPort, error) {
	var roomSize int
	if roomNameSize.Size <= 0 || roomNameSize.Size > maxRoomSize {
		roomSize = maxRoomSize
	} else {
		roomSize = int(roomNameSize.Size)
	}

	b.mu.RLock()
	_, ok := b.rooms[roomNameSize.Name]
	b.mu.RUnlock()
	if ok {
		return nil, fmt.Errorf("room \"%s\" already exists", roomNameSize.Name)
	}
	cr, err := newRoom(roomSize)
	if err != nil {
		return &rpc.RoomPort{Port: 0, Exists: false}, err
	}
	roomPort := int32(cr.getPort())

	go func() {
		log.Printf("creating room \"%s\" at port %d\n", roomNameSize.Name, roomPort)
		b.mu.Lock()
		b.rooms[roomNameSize.Name] = roomPort
		b.mu.Unlock()

		cr.Open()

		b.mu.Lock()
		delete(b.rooms, roomNameSize.Name)
		b.mu.Unlock()
		log.Printf("room \"%s\" at port %d closed successfully", roomNameSize.Name, roomPort)
	}()
	return &rpc.RoomPort{Port: roomPort, Exists: true}, nil
}

func (b *Butler) FindRoom(ctx context.Context, roomName *rpc.RoomName) (*rpc.RoomPort, error) {
	b.mu.RLock()
	roomPort, ok := b.rooms[roomName.Name]
	b.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("room %s does not exist", roomName.Name)
	}
	return &rpc.RoomPort{Port: roomPort, Exists: true}, nil
}
