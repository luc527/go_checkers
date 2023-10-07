package minimax

import (
	"fmt"
	"reflect"
	"runtime"

	c "github.com/luc527/go_checkers/core"
)

// the heuristics take a game and not just a board
// because the game caches the piece count
// and some heuristics rely on the piece count
// -- maybe not a very nice abstraction

// TODO now the game doesn't cache the piece count, redo this with the heuristics taking Board

type Heuristic func(g *c.Game, player c.Color) float64

func (h Heuristic) String() string {
	return fmt.Sprintf("%q", runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name())
}

var _ Heuristic = UnweightedCountHeuristic
var _ Heuristic = WeightedCountHeuristic

func UnweightedCountHeuristic(g *c.Game, player c.Color) float64 {
	count := g.Board().PieceCount()
	whites := int(count.WhitePawns + count.WhiteKings)
	blacks := int(count.BlackPawns + count.BlackKings)

	factor := 1
	if player == c.BlackColor {
		factor = -1
	}

	return float64(factor * (whites - blacks))
}

func WeightedCountHeuristic(g *c.Game, player c.Color) float64 {
	const (
		pawnWeight = 1
		kingWeight = 2
	)

	count := g.Board().PieceCount()
	whites := int(count.WhitePawns*pawnWeight + count.WhiteKings*kingWeight)
	blacks := int(count.BlackPawns*pawnWeight + count.BlackKings*kingWeight)

	factor := 1
	if player == c.BlackColor {
		factor = -1
	}

	return float64(factor * (whites - blacks))
}

// TODO distance heuristic

// TODO "clusters" heuristic but simpler to compute, i.e. only look at neighbours

// TODO more general heuristic where you give a weight to each tile
// and return a heuristic that uses that weight map
// (distance heuristic is a specific version)