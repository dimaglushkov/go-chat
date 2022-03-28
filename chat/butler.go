package chat

import (
	"golang.org/x/net/context"
	"log"
	"sync"
)

const maxRoomSize = 99

type Butler struct {
	ButlerServer
	mu    sync.RWMutex
	rooms map[string]int32
}

func NewButler() (butler Butler) {
	butler.rooms = make(map[string]int32)
	return
}

func (b *Butler) CreateRoom(ctx context.Context, roomNameSize *RoomNameSize) (*RoomPort, error) {
	var roomSize int
	if roomNameSize.Size <= 0 || roomNameSize.Size > maxRoomSize {
		roomSize = maxRoomSize
	} else {
		roomSize = int(roomNameSize.Size)
	}

	cr, err := newRoom(roomSize)
	if err != nil {
		return &RoomPort{Port: 0, Exists: false}, err
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
	return &RoomPort{Port: roomPort, Exists: true}, nil
}

func (b *Butler) FindRoom(ctx context.Context, roomName *RoomName) (*RoomPort, error) {
	b.mu.RLock()
	roomPort := b.rooms[roomName.Name]
	b.mu.RUnlock()
	return &RoomPort{Port: roomPort, Exists: true}, nil
}
