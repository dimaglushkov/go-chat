package server

import (
	"bufio"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dimaglushkov/go-chat/api/butlerpb"
)

func TestButler_CreateRoomValid(t *testing.T) {
	butler := NewButler()
	ctx := context.Background()
	rns := &butlerpb.RoomNameSize{
		Size: 1200,
		Name: "testRoom",
	}

	rp, err := butler.CreateRoom(ctx, rns)
	assert.NoError(t, err)

	conn, err := connectToRoom(rp)
	require.NoError(t, err)
	defer conn.Close()

	w := bufio.NewWriter(conn)
	err = sendMsg(w, "testName")
	require.NoError(t, err)

	err = sendMsg(w, "test message")
	require.NoError(t, err)
}

/*func TestButler_FindRoom(t *testing.T) {
	butler := NewButler()
	ctx := context.Background()
	rns := &butlerpb.RoomNameSize{
		Size: 20,
		Name: "testRoom",
	}

	rp, err := butler.CreateRoom(ctx, rns)
	require.NoError(t, err)

	conn, err := connectToRoom(rp)
	require.NoError(t, err)

	w := bufio.NewWriter(conn)
	_, err = w.Write([]byte("testName\n"))
	require.NoError(t, err)
	err = w.Flush()
	require.NoError(t, err)

	r1, err := butler.FindRoom(ctx, &butlerpb.RoomName{Name: rns.Name})
	require.NoError(t, err)
	require.NotNil(t, r1)

	r2, err := butler.FindRoom(ctx, &butlerpb.RoomName{Name: "_testName_"})
	require.Error(t, err)
	require.Nil(t, r2)

	conn.Close()
	time.Sleep(3 * time.Second)
	r1, err = butler.FindRoom(ctx, &butlerpb.RoomName{Name: rns.Name})
	require.Error(t, err)
	require.Nil(t, r1)

}*/
