package game
type GameState int

const (
	InProgress GameState = iota
	Win
	Draw
)

type Rule interface {
	IsValidMove(board Board,x,y int,playerID int) bool
	CheckGameState(board Board,lastMove Point) (GameState,int) // 返回游戏状态与赢家
}

type StandardRule struct{}

func (s *StandardRule) IsValidMove(board Board, x, y int, playerID int) bool {
	if _, exists := board[Point{x, y}]; exists {
		return false
	}
	// 其他规则判断
	return true
}

func (s *StandardRule) CheckGameState(board Board, lastMove Point) (GameState, int) {
	player := board[lastMove]
	if player == 0 {
		return InProgress, 0
	}
	direction := []Point{{1, 0}, {0, 1}, {1, 1}, {1, -1}}
	for _, dir := range direction {
		count := 1
		for i := 1; i <= 4; i++ {
			if board[Point{lastMove.X + dir.X*i, lastMove.Y + dir.Y*i}] == player {
				count++
			} else {
				break
			}
		}
		for i := 1; i <= 4; i++ {
			if board[Point{lastMove.X - dir.X*i, lastMove.Y - dir.Y*i}] == player {
				count++
			} else {
				break
			}
		}
		if count >= 5 {
			return Win, player
		}
	}
	return InProgress, 0
}
