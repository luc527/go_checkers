package core

import (
	"bytes"
	"fmt"
)

type GameResult byte

const (
	PlayingResult = GameResult(iota)
	WhiteWonResult
	BlackWonResult
	DrawResult
)

func (r GameResult) Over() bool {
	return r != PlayingResult
}

func (r GameResult) HasWinner() bool {
	return r == WhiteWonResult || r == BlackWonResult
}

func (r GameResult) Winner() Color {
	if r == WhiteWonResult {
		return WhiteColor
	} else {
		return BlackColor
	}
}

func (r GameResult) String() string {
	switch r {
	case PlayingResult:
		return "playing"
	case WhiteWonResult:
		return "white won"
	case BlackWonResult:
		return "black won"
	case DrawResult:
		return "draw"
	default:
		return "INVALID GameResult"
	}
}

func (r GameResult) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%q", r.String())
	return buf.Bytes(), nil
}

func (r *GameResult) UnmarshalJSON(bs []byte) error {
	if len(bs) < 2 || bs[0] != '"' || bs[len(bs)-1] != '"' {
		return fmt.Errorf("gameResult unmarshal json: not a string")
	}
	s := string(bs[1 : len(bs)-1])
	switch s {
	case "playing":
		*r = PlayingResult
	case "white won":
		*r = WhiteWonResult
	case "black won":
		*r = BlackWonResult
	case "draw":
		*r = DrawResult
	default:
		return fmt.Errorf("gameResult unmarshal json: invalid string: %q", s)
	}
	return nil
}

type gameState struct {
	turnsSinceCapture    int16
	turnsSincePawnMove   int16
	turnsInSpecialEnding int16
	plies                []Ply
}

type UndoInfo struct {
	plyDone   Ply
	prevState gameState
}

type Game struct {
	stagnantTurnsToDraw int16 // stagnant here means no captures and no pawn moves
	board               *Board
	toPlay              Color
	state               gameState
}

func (g *Game) String() string {
	return fmt.Sprintf(
		"{ToPlay: %v, turnsSinceCapture: %v, turnsSincePawnMove: %v, turnsInSpecialEnding: %v, Board:\n%v\n}",
		g.toPlay,
		g.state.turnsSinceCapture,
		g.state.turnsSincePawnMove,
		g.state.turnsInSpecialEnding,
		g.board,
	)
}

func NewCustomGame(stagnantTurnsToDraw int16, initialBoard *Board, initalPlayer Color) *Game {
	var g Game

	if initialBoard == nil {
		g.board = new(Board)
		PlaceInitialPieces(g.board)
	} else {
		g.board = initialBoard
	}

	g.stagnantTurnsToDraw = stagnantTurnsToDraw

	g.toPlay = initalPlayer

	g.state.turnsSinceCapture = 0
	g.state.turnsSincePawnMove = 0
	g.state.turnsInSpecialEnding = 0
	// once we get in a special ending turnsInSpecialEnding becomes 1 and increases each turn

	g.BoardChanged(nil)

	return &g
}

func NewGame() *Game {
	return NewCustomGame(20, nil, WhiteColor)
}

func (g *Game) Board() *Board {
	return g.board
}

func (g *Game) ToPlay() Color {
	return g.toPlay
}

func (g *Game) WhiteToPlay() bool {
	return g.toPlay == WhiteColor
}

func (g *Game) BlackToPlay() bool {
	return g.toPlay == BlackColor
}

func (g *Game) DoPly(p Ply) (*UndoInfo, error) {
	if len(p) == 0 {
		return nil, fmt.Errorf("game: empty ply")
	}
	if err := PerformInstructions(g.board, p); err != nil {
		return nil, err
	}
	prevState := g.state
	g.toPlay = g.toPlay.Opposite()
	g.BoardChanged(p)

	return &UndoInfo{plyDone: p, prevState: prevState}, nil
}

func (g *Game) Result() GameResult {
	count := g.board.PieceCount()
	whiteCount := count.WhiteKings + count.WhitePawns
	blackCount := count.BlackKings + count.BlackPawns

	if whiteCount == 0 {
		return BlackWonResult
	} else if blackCount == 0 {
		return WhiteWonResult
	}

	if g.state.turnsInSpecialEnding == 5 {
		return DrawResult
	}

	if g.state.turnsSincePawnMove >= g.stagnantTurnsToDraw && g.state.turnsSinceCapture >= g.stagnantTurnsToDraw {
		return DrawResult
	}

	if len(g.Plies()) == 0 {
		if g.toPlay == WhiteColor {
			return BlackWonResult
		} else {
			return WhiteWonResult
		}
	}

	return PlayingResult
}

func (g *Game) UndoPly(undo *UndoInfo) {
	UndoInstructions(g.board, undo.plyDone)
	g.toPlay = g.toPlay.Opposite()
	g.state = undo.prevState
}

func (g *Game) Copy() *Game {
	// plies shallow-copied
	// board deep-copied
	return &Game{
		state: gameState{
			turnsSinceCapture:    g.state.turnsSinceCapture,
			turnsSincePawnMove:   g.state.turnsSincePawnMove,
			turnsInSpecialEnding: g.state.turnsInSpecialEnding,
			plies:                g.state.plies,
		},
		stagnantTurnsToDraw: g.stagnantTurnsToDraw,
		board:               g.board.Copy(),
		toPlay:              g.toPlay,
	}
}

func (g *Game) Equals(o *Game) bool {
	if g == nil && o == nil {
		return true
	}
	if g == nil || o == nil {
		return false
	}

	return g.toPlay == o.toPlay &&
		g.state.turnsInSpecialEnding == o.state.turnsInSpecialEnding &&
		g.state.turnsSinceCapture == o.state.turnsSinceCapture &&
		g.state.turnsSincePawnMove == o.state.turnsSincePawnMove &&
		g.board.Equals(o.board)
}

func (g *Game) BoardChanged(ply Ply) {
	count := g.board.PieceCount()

	if inSpecialEnding(count) {
		g.state.turnsInSpecialEnding++
	} else {
		g.state.turnsInSpecialEnding = 0
	}

	if ply != nil {
		isCapture := false
		isPawnMove := false

		for _, ins := range ply {
			if ins.t == CaptureInstruction {
				isCapture = true
			}
			if ins.t == MoveInstruction {
				_, kind := g.board.Get(ins.row, ins.col)
				if kind == PawnKind {
					isPawnMove = true
				}
			}
		}

		if isCapture {
			g.state.turnsSinceCapture = 0
		} else {
			g.state.turnsSinceCapture++
		}

		if isPawnMove {
			g.state.turnsSincePawnMove = 0
		} else {
			g.state.turnsSincePawnMove++
		}
	}

	g.state.plies = nil
}

func (g *Game) generatePlies() []Ply {
	return GeneratePlies(make([]Ply, 0, 10), g.board, g.toPlay)
}

func (g *Game) Plies() []Ply {
	// Generated on demand, then cached
	if g.state.plies == nil {
		g.state.plies = g.generatePlies()
	}
	return g.state.plies
}

func oneColorSpecialEnding(ourKings, ourPawns, theirKings, theirPawns int8) bool {
	// a) 2 damas vs 2 damas
	// b) 2 damas vs 1 dama
	// c) 2 damas vs 1 dama e 1 pedra
	// d) 1 dama  vs 1 dama
	// e) 1 dama  vs 1 dama e 1 pedra
	//    ^ our   vs ^ their
	if ourPawns > 0 {
		return false
	}
	if ourKings == 2 {
		return (theirPawns == 0 && (theirKings == 2 || theirKings == 1)) || // a ou b
			(theirPawns == 1 && theirKings == 1) // c
	}
	if ourKings == 1 {
		return theirKings == 1 && (theirPawns == 0 || theirPawns == 1) // d or e
	}
	return false

	// let's check whether:
	// once we get in a special ending any further capture still leaves us in another special ending

	// a -> b by losing 1 king
	// b -> (win) by losing 1 king
	// b -> d by losing 1 king
	// c -> e by losing 1 king
	// c -> b by losing 1 pawn
	// c -> 2 damas vs 1 pedra, not an special ending!
	// d -> (win) by losing either king
	// e -> d by losing 1 pawn
	// e -> 1 dama vs 1 pedra, again not an special ending!

	// this means we need to check every time
}

func inSpecialEnding(c PieceCount) bool {
	wk, wp := c.WhiteKings, c.WhitePawns
	bk, bp := c.BlackKings, c.BlackPawns
	return oneColorSpecialEnding(wk, wp, bk, bp) || oneColorSpecialEnding(bk, bp, wk, wp)
}
