package websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(c *gin.Context) {
	roomCode := c.Param("roomCode")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	room, exists := Rooms[roomCode]

	if !exists {
		room = &Room{
			Clients: make(map[*websocket.Conn]bool),
		}
		Rooms[roomCode] = room
	}

	room.Mutex.Lock()
	room.Clients[conn] = true
	room.Mutex.Unlock()

	defer func() {
		room.Mutex.Lock()
		delete(room.Clients, conn)
		room.Mutex.Unlock()
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			break
		}

		Broadcast(roomCode, message)
	}
}