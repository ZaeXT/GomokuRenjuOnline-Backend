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


type MakeMovePayload struct {
	X      int `json:"x"`
	Y      int `json:"y"`
}

type JoinRoomPayload struct {
	RoomID string `json:"roomId"`
}

type Piece struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Player int `json:"player"`
}

type GameStateUpdatePayload struct {
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