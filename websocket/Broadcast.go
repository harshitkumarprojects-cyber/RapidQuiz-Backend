package websocket

func Broadcast(roomCode string, message []byte) {
	room, exists := Rooms[roomCode]

	if !exists {
		return
	}

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	for client := range room.Clients {
		err := client.WriteMessage(1, message)

		if err != nil {
			client.Close()
			delete(room.Clients, client)
		}
	}
}