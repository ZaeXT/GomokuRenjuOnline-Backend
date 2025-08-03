package game

import (
	"errors"
	"GomokuRenjuOnline-Backend/pkg/protocol"
)

const (
	InitialVisibleSize = 15
)

type Game struct {
	Board  Board
	Rule   Rule
	VisibleSize int
	PieceCount  int
	Players     map[int]bool
	CurrentPlayer int
	IsGameOver   bool
	Winner	  *int
}

func NewGame(rule Rule) *Game {
	return &Game{
		Board: make(Board),
		Rule:  rule,
		VisibleSize: InitialVisibleSize,
		Players: make(map[int]bool),
		CurrentPlayer: 1,
	}
}

func (g *Game) AddPlayer(playerID int) {
	g.Players[playerID] = true
}

func (g *Game) MakeMove(playerID int, x, y int) error {
	if g.IsGameOver {
		return errors.New("game is over")
	}
	if playerID != g.CurrentPlayer {
		return errors.New("not your turn")
	}
	if !g.Rule.IsValidMove(g.Board, x, y, playerID) {
		return errors.New("invalid move")
	}
	movePoint := Point{x, y}
	g.Board[movePoint] = playerID
	g.PieceCount++
	state, winnerID := g.Rule.CheckGameState(g.Board, movePoint)
	if state == Win {
		g.IsGameOver = true
		g.Winner = &winnerID
	} else {
		g.CurrentPlayer = 3 - g.CurrentPlayer
		g.checkExpansion()
	}
	return nil
}

func (g *Game) checkExpansion() {
	currentAreaSize := g.VisibleSize * g.VisibleSize
	threshold := float64(currentAreaSize) * 0.8
	if float64(g.PieceCount) >= threshold {
		newSize := int(float64(g.VisibleSize) * 1.5)
		if newSize%2 == 0 {
			newSize++
		}
		g.VisibleSize = newSize
	}
}

func (g *Game) Reset() {
	g.Board = make(Board)
	g.VisibleSize = InitialVisibleSize
	g.PieceCount = 0
	g.CurrentPlayer = 1
	g.IsGameOver = false
	g.Winner = nil
}

func (g *Game) CreateStatePayload(clientID int, roomName string) protocol.GameStateUpdatePayload {
	pieces := make([]protocol.Piece, 0, len(g.Board))
	for point, playerID := range g.Board {
		pieces = append(pieces, protocol.Piece{
			X:      point.X,
			Y:      point.Y,
			Player: playerID,
		})
	}
	return protocol.GameStateUpdatePayload{
		RoomName:	roomName,
		VisibleSize:  g.VisibleSize,
		Pieces:      pieces,
		CurrentPlayer: g.CurrentPlayer,
		IsGameOver:   g.IsGameOver,
		Winner:      g.Winner,
		YourPlayerID: clientID,
	}
}

func (g *Game) RemovePlayer(playerID int) {
	delete(g.Players, playerID)
}