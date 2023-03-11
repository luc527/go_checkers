package main

type gameState byte

const (
	playingState = gameState(iota)
	whiteWonState
	blackWonState
	drawState
)

// the part of the state that you need to remember in order to undo
type rememberedState struct {
	state               gameState
	plies               []ply
	lastPly             ply
	roundsSinceCapture  int
	roundsSincePawnMove int
}

type game struct {
	captureRule
	bestRule
	rememberedState
	board   *board
	toPlay  color
	history []rememberedState
}

func newGameWithBoardAndPlayer(captureRule captureRule, bestRule bestRule, initialBoard *board, initalPlayer color) *game {
	var g game

	if initialBoard == nil {
		g.board = new(board)
		placeInitialPieces(g.board)
	} else {
		g.board = initialBoard
	}

	g.captureRule = captureRule
	g.bestRule = bestRule

	g.toPlay = initalPlayer

	g.lastPly = nil
	g.roundsSinceCapture = 0
	g.roundsSincePawnMove = 0

	g.boardChanged()

	return &g

}

func newGame(captureRule captureRule, bestRule bestRule) *game {
	return newGameWithBoardAndPlayer(captureRule, bestRule, nil, whiteColor)
}

func (g *game) isOver() bool {
	return g.state != playingState
}

func (g *game) hasWinner() bool {
	return g.state == whiteWonState || g.state == blackWonState
}

func (g *game) winner() color {
	if g.state == whiteWonState {
		return whiteColor
	}
	return blackColor
}

func (g *game) doPly(p ply) {
	performInstructions(g.board, p)
	g.toPlay = g.toPlay.opposite()

	// save current state in the history
	g.history = append(g.history, g.rememberedState)
	g.lastPly = p

	g.boardChanged()

	// Draw detection

	isCapture := false
	isPawnMove := false

	for _, ins := range p {
		if ins.t == captureInstruction {
			isCapture = true
		}
		if ins.t == moveInstruction {
			_, kind := g.board.get(ins.row, ins.col)
			if kind == pawnKind {
				isPawnMove = true
			}
		}
	}

	if isCapture {
		g.roundsSinceCapture = 0
	} else {
		g.roundsSinceCapture++
	}

	if isPawnMove {
		g.roundsSincePawnMove = 0
	} else {
		g.roundsSincePawnMove++
	}

	// TODO roundsInSpecialMove

	if g.roundsSincePawnMove >= 20 && g.roundsSinceCapture >= 20 {
		g.state = drawState
	}
}

func (g *game) undoLastPly() {
	if g.lastPly == nil {
		return
	}

	undoInstructions(g.board, g.lastPly)
	g.toPlay = g.toPlay.opposite()
	g.rememberedState = g.history[len(g.history)-1]

	g.history = g.history[:len(g.history)-1]
}

func (g *game) deepishCopy() *game {
	// plies, lastPly, history all shallow-copied
	// board deep-copied
	return &game{
		captureRule: g.captureRule,
		bestRule:    g.bestRule,
		rememberedState: rememberedState{
			state:   g.state,
			plies:   g.plies,
			lastPly: g.lastPly,
		},
		board:   g.board.copy(),
		toPlay:  g.toPlay,
		history: g.history,
	}
}

func (g *game) equals(o *game) bool {
	if g == nil && o == nil {
		return true
	}
	if g == nil || o == nil {
		return false
	}

	pliesEq := func(a []ply, b []ply) bool {
		if len(a) != len(b) {
			return false
		}
		for i, x := range a {
			y := b[i]
			if !sliceEq(x, y) {
				return false
			}
		}
		return true
	}

	return g.captureRule == o.captureRule &&
		g.bestRule == o.bestRule &&
		g.state == o.state &&
		g.toPlay == o.toPlay &&
		g.board.equals(o.board) &&
		sliceEq(g.lastPly, o.lastPly) &&
		pliesEq(g.plies, o.plies)
}

func (g *game) boardChanged() {
	count := g.board.pieceCount()
	whiteCount := count.whiteKings + count.whitePawns
	blackCount := count.blackKings + count.blackPawns

	if whiteCount == 0 {
		g.state = blackWonState
	} else if blackCount == 0 {
		g.state = whiteWonState
	}

	if g.isOver() {
		return
	}

	g.plies = generatePlies(g.board, g.toPlay, g.captureRule, g.bestRule)
	if len(g.plies) == 0 {
		if g.toPlay == whiteColor {
			g.state = blackWonState
		} else {
			g.state = whiteWonState
		}
	}
}
