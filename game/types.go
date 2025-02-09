package game

type Player struct {
	Name  string
	Score int
}

type Game struct {
	Players          []Player
	EliminationScore int
}
