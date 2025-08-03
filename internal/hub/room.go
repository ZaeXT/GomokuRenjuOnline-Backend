package hub

import (
	"encoding/json"
	"log"
	"GomokuRenjuOnline-Backend/pkg/protocol"
	"GomokuRenjuOnline-Backend/internal/game"
)

type messageFromClient struct {
	client  *Client
	message *protocol.InboundMessage
}

type Room struct {
	ID          string
	Name        string
	hub         *Hub
	clients     map[*Client]bool
	game        *game.Game
	register    chan *Client
	unregister  chan *Client
	forward     chan messageFromClient
}

func newRoom(id, name string, hub *Hub) *Room {
	StandardRule := &game.StandardRule{}
	return &Room{
		ID:         id,
		Name:       name,
		hub:        hub,
		clients:    make(map[*Client]bool),
		game:       game.NewGame(StandardRule),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		forward:    make(chan messageFromClient),
	}
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.register:
			r.handleRegister(client)
		case client := <-r.unregister:
			r.handleUnregister(client)
		case msg := <-r.forward:
			r.handleMessage(msg.client, msg.message)
		}
	}
}

func (r *Room) handleRegister(client *Client) {
	if len(r.clients) >= 2 {
		sendError(client, "Room is full.")
		return
	}
	var assignedId int
	player1Exists := false
	for c := range r.clients {
		if c.ID == 1 {
			player1Exists = true
			break
		}
	}
	if !player1Exists {
		assignedId = 1
	} else {
		assignedId = 2
	}
	r.clients[client] = true
	client.room = r
	client.ID = assignedId
	r.game.AddPlayer(assignedId)
	log.Printf("Client %d registered to room %s as player %d", client.conn.RemoteAddr(), r.ID, assignedId)
	r.hub.broadcastRoomList()
	r.broadcastGameState()
}

func (r *Room) handleUnregister(client *Client) {
	if _, ok := r.clients[client]; ok {
		r.game.RemovePlayer(client.ID)
		delete(r.clients, client)
		close(client.send)
		r.hub.broadcastRoomList()
		log.Printf("Client %d unregistered from room %s", client.ID, r.ID)
		if !r.game.IsGameOver {
			r.game.IsGameOver = true
			winner := 3 - client.ID
			r.game.Winner = &winner
			r.broadcastGameState()
		}
	}
}

func (r *Room) handleMessage(client *Client, msg *protocol.InboundMessage) {
	switch msg.Type {
	case "MAKE_MOVE":
		var payload protocol.MakeMovePayload
		err := json.Unmarshal(msg.Payload, &payload)
		if err != nil {
			log.Printf("Failed to unmarshal MAKE_MOVE payload: %v", err)
			//r.sendError(client, err.Error())
			return
		}
		err = r.game.MakeMove(client.ID, payload.X, payload.Y)
		if err != nil {
			r.sendError(client, err.Error())
			return
		}
		r.broadcastGameState()
	case "RESTART_GAME":
		if r.game.IsGameOver {
			r.game.Reset()
			log.Printf("Game reset in room %s by Player %d", r.ID, client.ID)
			r.broadcastGameState()
		}
	}
}

func (r *Room) broadcastGameState() {
	log.Printf("Broadcasting game state for room '%s' to %d clients", r.Name, len(r.clients))
	for client := range r.clients {
		payload := r.game.CreateStatePayload(client.ID,r.Name)
		log.Printf(" -> Sending to Player %d: roomName=%s, yourPlayerId=%d, currentPlayer=%d", client.ID, payload.RoomName, payload.YourPlayerID, payload.CurrentPlayer)
		msg := protocol.OutboundMessage{
			Type:    "GAME_STATE_UPDATE",
			Payload: payload,
		}
		select {
		case client.send <- msg:
		default:
			log.Printf("Failed to send to client %d. Closing connection.", client.ID)
			close(client.send)
			delete(r.clients, client)
		}
		
	}
}

func (r *Room) sendError(client *Client, message string) {
	msg := protocol.OutboundMessage{
		Type:    "ERROR",
		Payload: protocol.ErrorPayload{Message: message},
	}
	select {
	case client.send <- msg:
	default:
		log.Printf("Failed to send error to client %d.", client.ID)
	}
}