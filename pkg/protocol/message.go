package protocol

import "encoding/json"

type InboundMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type OutboundMessage struct {
	Type    string          `json:"type"`
	Payload interface{}     `json:"payload"`
}

type CreateRoomPayload struct {
	Name string `json:"name"`
}

type MakeMovePayload struct {
	X      int `json:"x"`
	Y      int `json:"y"`
}

type JoinRoomPayload struct {
	ID string `json:"id"`
}

type RoomInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	PlayerCount int  `json:"playerCount"`
}

type RoomListUpdatePayload struct {
	Rooms []RoomInfo `json:"rooms"`
}

type Piece struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Player int `json:"player"`
}

type GameStateUpdatePayload struct {
	RoomName      string  `json:"roomName"`
	VisibleSize   int    `json:"visible_size"`
	Pieces        []Piece `json:"pieces"`
	CurrentPlayer int  `json:"currentPlayer"`
	IsGameOver    bool   `json:"isGameOver"`
	Winner        *int    `json:"winner"` // 支持null
	YourPlayerID  int    `json:"yourPlayerId"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}