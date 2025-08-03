package hub

import (
	"encoding/json"
	"log"
	"sync"
	"github.com/gorilla/websocket"
	"GomokuRenjuOnline-Backend/pkg/protocol"
	"github.com/google/uuid"
)

type Hub struct {
	clients   map[*Client]bool
	rooms      map[string]*Room
	broadcast  chan messageFromClient
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex // 保护rooms的并发访问
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]*Room),
		broadcast:  make(chan messageFromClient),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)
		case client := <-h.unregister:
			h.handleUnregister(client)
		case msg := <-h.broadcast:
			h.handleMessage(msg.client, msg.message)
		}
	}
}

func (h *Hub) handleRegister(client *Client) {
	h.mu.Lock()
	h.clients[client] = true
	h.mu.Unlock()
	log.Printf("Client connected. Total clients: %d", len(h.clients))
	h.sendRoomList(client)
}

func (h *Hub) handleUnregister(client *Client) {
	h.mu.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		if client.room != nil {
			client.room.unregister <- client
		}
	}
	h.mu.Unlock()
	log.Printf("Client disconnected. Total clients: %d", len(h.clients))
}

func (h *Hub) handleMessage(client *Client, msg *protocol.InboundMessage) {
	if client.room != nil {
		client.room.forward <- messageFromClient{client: client,message: msg}
		return
	}
	switch msg.Type {
	case "CREATE_ROOM":
		var payload protocol.CreateRoomPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Printf("Error unmarshalling CREATE_ROOM: %v", err)
			return
		}
		h.handleCreateRoom(client, payload.Name)
	case "JOIN_ROOM":
		var payload protocol.JoinRoomPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Printf("Error unmarshalling JOIN_ROOM: %v", err)
			return
		}
		h.handleJoinRoom(client, payload.ID)
	}
}

func (h *Hub) handleCreateRoom(client *Client, name string) {
	h.mu.Lock()
	roomID := uuid.New().String()
	room := newRoom(roomID, name, h)
	h.rooms[roomID] = room
	go room.Run()
	h.mu.Unlock()
	log.Printf("Player %d created a new room: %s (%s)", client.ID, name, roomID)
	room.register <- client
	h.broadcastRoomList()
}

func (h *Hub) handleJoinRoom(client *Client, roomID string) {
	h.mu.RLock()
	room, exists := h.rooms[roomID]
	h.mu.RUnlock()
	if !exists {
		sendError(client, "Room not found.")
		return
	}
	room.register <- client
	h.broadcastRoomList()
}

func (h *Hub) sendRoomList(client *Client) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	log.Printf("Broadcasting room list update to all clients.")
	infos := make([]protocol.RoomInfo, 0, len(h.rooms))
	for _, room := range h.rooms {
		if len(room.clients) < 2 && !room.game.IsGameOver {
			infos = append(infos, protocol.RoomInfo{
				ID:         room.ID,
				Name:       room.Name,
				PlayerCount: len(room.clients),
			})
		}
	}
	msg := protocol.OutboundMessage{
		Type:    "ROOM_LIST_UPDATE",
		Payload: protocol.RoomListUpdatePayload{Rooms: infos},
	}
	for client := range h.clients {
		if client.room == nil {
			client.send <- msg
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

func sendError(client *Client, message string) {
	msg := protocol.OutboundMessage{
		Type:    "ERROR",
		Payload: protocol.ErrorPayload{Message: message},
	}
	client.send <- msg

}

func (h *Hub) broadcastRoomList() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	infos := make([]protocol.RoomInfo, 0, len(h.rooms))
	for _, room := range h.rooms {
		if len(room.clients) < 2 && !room.game.IsGameOver {
			infos = append(infos, protocol.RoomInfo{
				ID:         room.ID,
				Name:       room.Name,
				PlayerCount: len(room.clients),
			})
		}
	}
	msg := protocol.OutboundMessage{
		Type:    "ROOM_LIST_UPDATE",
		Payload: protocol.RoomListUpdatePayload{Rooms: infos},
	}
	for client := range h.clients {
		client.send <- msg
	}
}
