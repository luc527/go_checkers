package main

import (
	"testing"
)

func TestCrownInstruction(t *testing.T) {
	var row, col byte
	row, col = 5, 6
	i := CrownInstruction(row, col)
	if i.t != crownInstruction {
		t.Errorf("expected instruction to be of type crown, is %s", i.t)
		return
	}
	if i.row != row || i.col != col {
		t.Errorf("expected instruction to be crowning %d %d but is crowning %d %d", row, col, i.row, i.col)
	}
}

func TestMoveInstruction(t *testing.T) {
	var srow, scol byte
	var drow, dcol byte

	// invalid, whatever; it's not what we're testing here
	srow, scol = 1, 3
	drow, dcol = 2, 5

	i := MoveInstruction(srow, scol, drow, dcol)

	if i.t != moveInstruction {
		t.Errorf("expected type move but is of type %s", i.t)
		return
	}

	if i.row != srow || i.col != scol {
		t.Errorf("expected source %d %d but is %d %d", srow, scol, i.row, i.col)
		return
	}

	if i.d[0] != drow || i.d[1] != dcol {
		t.Errorf("expected destination %d %d but is %d %d", drow, dcol, i.row, i.col)
		return
	}
}

func TestCaptureInstruction(t *testing.T) {
	type testCase struct {
		row, col byte
		c        Color
		k        Kind
	}

	cases := []testCase{
		{1, 2, WhiteColor, KingKind},
		{3, 1, BlackColor, PawnKind},
		{2, 2, WhiteColor, PawnKind},
		{5, 7, BlackColor, KingKind},
	}

	for _, test := range cases {
		row, col, c, k := test.row, test.col, test.c, test.k
		i := CaptureInstruction(row, col, c, k)

		if i.t != captureInstruction {
			t.Errorf("expected type capture but got type %s", i.t)
			return
		}

		if row != i.row || col != i.col {
			t.Errorf("expected coord %d %d but got %d %d", row, col, i.row, i.col)
			return
		}

		actualC := Color(i.d[0])
		if actualC != c {
			t.Errorf("expected color %s but got %s", c, actualC)
			return
		}

		actualK := Kind(i.d[1])
		if actualK != k {
			t.Errorf("expected kind %s but got %s", k, actualK)
			return
		}
	}

}

func TestMakeCrownInstruction(t *testing.T) {
	b := new(Board)

	var row, col byte
	row, col = 5, 4

	b.Set(row, col, WhiteColor, PawnKind)

	i := CrownInstruction(row, col)
	is := []Instruction{i}

	PerformInstructions(b, is)

	_, newKind := b.Get(row, col)
	if newKind != KingKind {
		t.Errorf("crown instruction failed, %d %d still a pawn", row, col)
	}

	UndoInstructions(b, is)

	_, oldKind := b.Get(row, col)
	if oldKind != PawnKind {
		t.Errorf("undo of crown instruction failed, %d %d still a king", row, col)
	}
}

func TestMakeMoveInstruction(t *testing.T) {
	b := new(Board)

	var frow, fcol byte //from
	var trow, tcol byte //to

	frow, fcol = 3, 7
	trow, tcol = 4, 6
	c, k := BlackColor, KingKind

	b.Set(frow, fcol, c, k)

	i := MoveInstruction(frow, fcol, trow, tcol)
	is := []Instruction{i}

	PerformInstructions(b, is)

	if b.IsOccupied(frow, fcol) {
		t.Errorf("after move, source should be empty")
	}

	if !b.IsOccupied(trow, tcol) {
		t.Errorf("after move, destination should be occupied")
	} else {
		ac, ak := b.Get(trow, tcol)
		if ac != c || ak != k {
			t.Errorf("piece changed after move, was %s %s now is %s %s", c, k, ac, ak)
		}
	}

	UndoInstructions(b, is)

	if b.IsOccupied(trow, tcol) {
		t.Errorf("after undo move, destination should be empty")
	}

	if !b.IsOccupied(frow, fcol) {
		t.Errorf("after undo move, source should be occupied")
	} else {
		ac, ak := b.Get(frow, fcol)
		if ac != c || ak != k {
			t.Errorf("piece changed after undo move, was %s %s now is %s %s", c, k, ac, ak)
		}
	}
}

func TestMakeCaptureInstruction(t *testing.T) {
	b := new(Board)

	var row, col byte
	row, col = 3, 6
	color, kind := WhiteColor, PawnKind

	b.Set(row, col, color, kind)

	t.Log("Before capture:")
	t.Log(b)

	i := CaptureInstruction(row, col, color, kind)
	is := []Instruction{i}

	PerformInstructions(b, is)

	t.Log("After capture:")
	t.Log(b)

	if b.IsOccupied(row, col) {
		t.Errorf("(%d, %d) should be empty after capture, is occupied", row, col)
	}

	UndoInstructions(b, is)

	t.Log("After undoing capture:")
	t.Log(b)

	if !b.IsOccupied(row, col) {
		t.Errorf("(%d, %d) should be occupied after undoing the capture, is empty", row, col)
	} else {
		actualColor, actualKind := b.Get(row, col)
		if actualColor != color || actualKind != kind {
			t.Errorf(
				"expected (%d, %d) to contain a %s %s after undoing the capture, but it contains a %s %s",
				row, col,
				color, kind,
				actualColor, actualKind,
			)
		}
	}
}

func TestInsSequence(t *testing.T) {

	b := new(Board)

	b.Set(3, 5, WhiteColor, PawnKind)
	b.Set(1, 0, BlackColor, KingKind)
	b.Set(2, 2, BlackColor, PawnKind)

	t.Log("Before:")
	t.Log("\n" + b.String())

	before := b.Copy()

	is := []Instruction{
		MoveInstruction(3, 5, 2, 4),
		CrownInstruction(2, 4),
		CaptureInstruction(2, 4, WhiteColor, KingKind),
		MoveInstruction(1, 0, 4, 6),
		MoveInstruction(2, 2, 3, 5),
		CrownInstruction(3, 5),
	}

	PerformInstructions(b, is)

	t.Log("After:")
	t.Log("\n" + b.String())

	assertOccupied(b, t, 3, 5)
	assertContains(b, t, 3, 5, BlackColor, KingKind)
	assertOccupied(b, t, 4, 6)
	assertContains(b, t, 4, 6, BlackColor, KingKind)
	assertEmpty(b, t, 1, 0)
	assertEmpty(b, t, 2, 2)
	assertEmpty(b, t, 2, 4)

	UndoInstructions(b, is)

	t.Log("After undo:")
	t.Log("\n" + b.String())

	for row := byte(0); row < 8; row++ {
		for col := byte(0); col < 8; col++ {
			wantOccupied := before.IsOccupied(row, col)
			gotOccupied := b.IsOccupied(row, col)

			if wantOccupied != gotOccupied {
				t.Errorf("row %d col %d should be occupied(%v) but is occupied(%v)", row, col, wantOccupied, gotOccupied)
			} else if gotOccupied {
				wantColor, wantKind := before.Get(row, col)
				gotColor, gotKind := b.Get(row, col)

				if wantColor != gotColor || wantKind != gotKind {
					t.Errorf("row %d col %d should contain %s %s but contains %s %s", row, col, wantColor, wantKind, gotColor, gotKind)
				}
			}
		}
	}
}

// TODO refactor other tests to use these assertions

func assertOccupied(b *Board, t *testing.T, row, col byte) {
	if !b.IsOccupied(row, col) {
		t.Errorf("row %d col %d should be occupied", row, col)
	}
}

func assertContains(b *Board, t *testing.T, row, col byte, c Color, k Kind) {
	ac, ak := b.Get(row, col)
	if ac != c || ak != k {
		t.Errorf("row %d col %d should contain %s %s but contains %s %s", row, col, c, k, ac, ak)
	}
}

func assertEmpty(b *Board, t *testing.T, row, col byte) {
	if b.IsOccupied(row, col) {
		t.Errorf("row %d col %d should be empty", row, col)
	}
}

// TODO test random instruction sequence
