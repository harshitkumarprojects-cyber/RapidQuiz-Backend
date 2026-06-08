package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Room struct {
	Clients map[*websocket.Conn]bool
	Mutex   sync.Mutex
}

var Rooms = make(map[string]*Room)