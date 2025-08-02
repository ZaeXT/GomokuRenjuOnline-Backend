package hub

import (
	"log"
	"sync"
	"github.com/gorilla/websocket"
	"GomokuRenjuOnline-Backend/pkg/protocol"
)

type Hub struct {
	rooms      map[string]*Room
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex // 保护rooms的并发访问
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			roomID := "room1"
			h.mu.Lock()
			room, exists := h.rooms[roomID]
			if !exists {
				log.Printf("Creating new room: %s", roomID)
				room = newRoom(roomID)
				h.rooms[roomID] = room
				go room.Run()
			}
			h.mu.Unlock()
			room.register <- client
		case client := <-h.unregister:
			if client.room != nil {
				client.room.unregister <- client
			}
		}
	}
}

func (h *Hub) CreateAndRegisterClient(conn *websocket.Conn) {
	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan protocol.OutboundMessage, 256),
	}
	h.register <- client
	go client.readPump()
	go client.writePump()
}