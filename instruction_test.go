package main

import "testing"

func TestMakeCrownInstruction(t *testing.T) {
	var row, col byte
	row, col = 5, 6
	i := makeCrownInstruction(row, col)
	if i.t != crownInstruction {
		t.Errorf("expected instruction to be of type crown, is %s", i.t)
		return
	}
	if i.row != row || i.col != col {
		t.Errorf("expected instruction to be crowning %d %d but is crowning %d %d", row, col, i.row, i.col)
	}
}

func TestMakeMoveInstruction(t *testing.T) {
	var srow, scol byte
	var drow, dcol byte

	// invalid, whatever; it's not what we're testing here
	srow, scol = 1, 3
	drow, dcol = 2, 5

	i := makeMoveInstruction(srow, scol, drow, dcol)

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

func TestMakeCaptureInstruction(t *testing.T) {
	type testCase struct {
		row, col byte
		c        color
		k        kind
	}

	cases := [...]testCase{
		{1, 2, kWhite, kKing},
		{3, 1, kBlack, kPawn},
		{2, 2, kWhite, kPawn},
		{5, 7, kBlack, kKing},
	}

	for _, test := range cases {
		row, col, c, k := test.row, test.col, test.c, test.k
		i := makeCaptureInstruction(row, col, c, k)

		if i.t != captureInstruction {
			t.Errorf("expected type capture but got type %s", i.t)
			return
		}

		if row != i.row || col != i.col {
			t.Errorf("expected coord %d %d but got %d %d", row, col, i.row, i.col)
			return
		}

		actualC := color(i.d[0])
		if actualC != c {
			t.Errorf("expected color %s but got %s", c, actualC)
			return
		}

		actualK := kind(i.d[1])
		if actualK != k {
			t.Errorf("expected kind %s but got %s", k, actualK)
			return
		}
	}

}

// TODO test performance and undo of each single type of instruction
// TODO test performance of a sequence, hardcoded example
// TODO test undo of a sequence, hardcoded example
// TODO test performance + undo of random instructions, checking whether startState == endState
