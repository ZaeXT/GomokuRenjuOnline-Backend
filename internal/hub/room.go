package hub

import (
	"encoding/json"
	"log"
	"time"
	"GomokuRenjuOnline-Backend/pkg/protocol"
	"GomokuRenjuOnline-Backend/internal/game"
	"sync"
	"sync/atomic"
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
	stop       chan struct{}
	isClose    uint32
	once      sync.Once
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
		stop:       make(chan struct{}),
	}
}

func (r *Room) Run() {
	defer func() {
		log.Printf("Room %s goroutine finished.", r.Name)
	}()
	for {
		if atomic.LoadUint32(&r.isClose) == 1 {
			return
		}
		select {
		case client := <-r.register:
			if atomic.LoadUint32(&r.isClose) == 1 {
				if client != nil {
					sendError(client, "The room has been closed.")
				}
				continue
			}
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
	if atomic.LoadUint32(&r.isClose) == 1 {
		return
	}
	if _, ok := r.clients[client]; ok {
		delete(r.clients, client)
		close(client.send)
		r.game.RemovePlayer(client.ID)
		log.Printf("Client %d unregistered from room %s", client.ID, r.ID)
		if len(r.clients) == 0 {
			log.Printf("Room %s is empty, closing immediately.", r.Name)
			r.close()
			return
		}
		if !r.game.IsGameOver {
			log.Printf("Game in room '%s' ends due to player %d disconnection.", r.Name, client.ID)
			r.game.IsGameOver = true
			winner := 3 - client.ID
			r.game.Winner = &winner
			r.endGameAndScheduleDestruction()
		}
	}
}

func (r *Room) handleMessage(client *Client, msg *protocol.InboundMessage) {
	if atomic.LoadUint32(&r.isClose) == 1 {
		return
	}
	switch msg.Type {
	case "MAKE_MOVE":
		var payload protocol.MakeMovePayload
		err := json.Unmarshal(msg.Payload, &payload)
		if err != nil {
			log.Printf("Failed to unmarshal MAKE_MOVE payload: %v", err)
			//r.sendError(client, err.Error())
			return
		}
		isGameOver, err := r.game.MakeMove(client.ID, payload.X, payload.Y)
		if err != nil {
			r.sendError(client, err.Error())
			return
		}
		if isGameOver {
			r.endGameAndScheduleDestruction()
		} else {
			r.broadcastGameState()
		}
	// case "RESTART_GAME":
	// 	if r.game.IsGameOver {
	// 		r.game.Reset()
	// 		log.Printf("Game reset in room %s by Player %d", r.ID, client.ID)
	// 		r.broadcastGameState()
	// 	}
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

func (r *Room) close() {
	r.once.Do(func() {
		atomic.StoreUint32(&r.isClose, 1)
		close(r.stop)
		r.hub.destroyRoom <- r
		log.Printf("Room %s close() method called.", r.Name)
	})
}

func (r *Room) endGameAndScheduleDestruction() {
	if atomic.LoadUint32(&r.isClose) == 1 {
		return
	}
	r.broadcastGameState()
	log.Printf("Game in room '%s' is over. Scheduling destruction.", r.Name)
	go func() {
		time.Sleep(1 * time.Second)
		r.close()
	}()
}
