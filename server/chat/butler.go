package chat

import (
	"golang.org/x/net/context"
	"log"
	"sync"
)

const roomSize = 20

type Butler struct {
	ButlerServer
	mu    sync.RWMutex
	rooms map[string]int32
}

func NewButler() (butler Butler) {
	butler.rooms = make(map[string]int32)
	return
}

func (b *Butler) CreateRoom(ctx context.Context, roomName *RoomName) (*RoomPort, error) {
	cr, err := newRoom(roomSize)
	if err != nil {
		return &RoomPort{Port: 0, Exists: false}, err
	}
	roomPort := int32(cr.getPort())

	go func() {
		log.Printf("creating room \"%s\" at port %d\n", roomName.Name, roomPort)
		b.mu.Lock()
		b.rooms[roomName.Name] = roomPort
		b.mu.Unlock()

		cr.Open()

		b.mu.Lock()
		delete(b.rooms, roomName.Name)
		b.mu.Unlock()
		log.Printf("room \"%s\" at port %d\n closed successfully", roomName.Name, roomPort)
	}()
	return &RoomPort{Port: roomPort, Exists: true}, nil
}

func (b *Butler) FindRoom(ctx context.Context, roomName *RoomName) (*RoomPort, error) {
	b.mu.RLock()
	roomPort := b.rooms[roomName.Name]
	b.mu.RUnlock()
	return &RoomPort{Port: roomPort, Exists: true}, nil
}
