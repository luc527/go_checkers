package core

import (
	"encoding/json"
	"testing"
)

func TestCrownInstruction(t *testing.T) {
	var row, col byte
	row, col = 5, 6
	i := MakeCrownInstruction(row, col)
	if i.t != CrownInstruction {
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

	i := MakeMoveInstruction(srow, scol, drow, dcol)

	if i.t != MoveInstruction {
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
		i := MakeCaptureInstruction(row, col, c, k)

		if i.t != CaptureInstruction {
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

	i := MakeCrownInstruction(row, col)
	is := []Instruction{i}

	if err := PerformInstructions(b, is); err != nil {
		t.Log(err)
		t.FailNow()
	}

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

	i := MakeMoveInstruction(frow, fcol, trow, tcol)
	is := []Instruction{i}

	if err := PerformInstructions(b, is); err != nil {
		t.Log(err)
		t.FailNow()
	}

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

	i := MakeCaptureInstruction(row, col, color, kind)
	is := []Instruction{i}

	if err := PerformInstructions(b, is); err != nil {
		t.Log(err)
		t.FailNow()
	}

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
		MakeMoveInstruction(3, 5, 2, 4),
		MakeCrownInstruction(2, 4),
		MakeCaptureInstruction(2, 4, WhiteColor, KingKind),
		MakeMoveInstruction(1, 0, 4, 6),
		MakeMoveInstruction(2, 2, 3, 5),
		MakeCrownInstruction(3, 5),
	}

	if err := PerformInstructions(b, is); err != nil {
		t.Log(err)
		t.FailNow()
	}

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

func TestInstructionEquals(t *testing.T) {
	mov1 := MakeMoveInstruction(1, 2, 3, 4)
	mov2 := MakeMoveInstruction(1, 2, 3, 6)

	if mov1.Equals(mov2) {
		t.Fail()
	}
	if !mov1.Equals(mov1) {
		t.Fail()
	}

	cap1 := MakeCaptureInstruction(1, 2, BlackColor, KingKind)
	cap2 := MakeCaptureInstruction(1, 2, WhiteColor, KingKind)
	if cap1.Equals(cap2) {
		t.Fail()
	}
	if !cap1.Equals(cap1) {
		t.Fail()
	}

	cro1 := MakeCrownInstruction(1, 2)
	cro2 := MakeCrownInstruction(7, 6)
	if cro1.Equals(cro2) {
		t.Fail()
	}
	if !cro1.Equals(cro1) {
		t.Fail()
	}

	// Test different types
	ins := []Instruction{mov1, cap1, cro1}
	for i, ins1 := range ins {
		for j, ins2 := range ins {
			if i == j {
				continue
			}
			if ins1.Equals(ins2) {
				t.Fail()
			}
		}
	}
}

func TestInstructionString(t *testing.T) {
	if MoveInstruction.String() != "move" {
		t.Fail()
	}
	if CaptureInstruction.String() != "capture" {
		t.Fail()
	}
	if CrownInstruction.String() != "crown" {
		t.Fail()
	}
	if InstructionType(9).String() != "UNKNOWN" {
		t.Fail()
	}
}

func TestInstructionMarshal(t *testing.T) {
	type test struct {
		i Instruction
		e []byte
	}
	tests := []test{
		{MakeMoveInstruction(1, 2, 6, 5), []byte("\"m1265\"")},
		{MakeMoveInstruction(7, 6, 1, 7), []byte("\"m7617\"")},
		{MakeCaptureInstruction(4, 4, WhiteColor, KingKind), []byte("\"c44wk\"")},
		{MakeCaptureInstruction(3, 1, WhiteColor, PawnKind), []byte("\"c31wp\"")},
		{MakeCaptureInstruction(1, 7, BlackColor, PawnKind), []byte("\"c17bp\"")},
		{MakeCaptureInstruction(2, 2, BlackColor, KingKind), []byte("\"c22bk\"")},
		{MakeCrownInstruction(1, 5), []byte("\"k15\"")},
	}
	for _, test := range tests {
		json, err := test.i.MarshalJSON()
		if err != nil {
			t.Logf("error: %v", err)
			t.Fail()
		}
		if string(json) != string(test.e) {
			t.Logf("wanted %s got %s", string(test.e), string(json))
			t.Fail()
		}
	}
}

func TestInstructionUnmarshalCorrectly(t *testing.T) {
	type test struct {
		s string
		i Instruction
	}
	tests := []test{
		{"\"m1223\"", MakeMoveInstruction(1, 2, 2, 3)},
		{"\"c12wk\"", MakeCaptureInstruction(1, 2, WhiteColor, KingKind)},
		{"\"c24wp\"", MakeCaptureInstruction(2, 4, WhiteColor, PawnKind)},
		{"\"c37bk\"", MakeCaptureInstruction(3, 7, BlackColor, KingKind)},
		{"\"c66bp\"", MakeCaptureInstruction(6, 6, BlackColor, PawnKind)},
		{"\"k45\"", MakeCrownInstruction(4, 5)},
	}
	i := &Instruction{}
	for _, test := range tests {
		err := i.UnmarshalJSON([]byte(test.s))
		if err != nil {
			t.Logf("error: %v", err)
			t.Fail()
		}
		if !i.Equals(test.i) {
			t.Logf("wanted %v got %v", test.i, i)
			t.Fail()
		}
	}
}

func TestInstructionUnmarshalIncorrectly(t *testing.T) {
	tests := []string{
		"\"b1234\"",
		"\"m123\"",
		"\"m1299\"",
		"\"c13w\"",
		"\"c33bb\"",
		"\"c33mp\"",
		"\"c19wk\"",
		"\"k08\"",
		"\"m12345\"",
		"\"c12bkk\"",
		"\"k666\"",
	}
	i := &Instruction{}
	for _, test := range tests {
		if err := i.UnmarshalJSON([]byte(test)); err == nil {
			t.Fail()
		}
	}
}

func TestJsonMarshalUnmarshalInstruction(t *testing.T) {
	is := []Instruction{
		MakeMoveInstruction(1, 3, 2, 5),
		MakeCrownInstruction(7, 5),
		MakeCaptureInstruction(3, 5, WhiteColor, KingKind),
		MakeCaptureInstruction(1, 4, WhiteColor, PawnKind),
		MakeCaptureInstruction(3, 2, BlackColor, KingKind),
		MakeCaptureInstruction(6, 2, BlackColor, PawnKind),
	}
	for _, i := range is {
		bs, err := json.Marshal(i)
		if err != nil {
			t.Logf("error marshalling: %v", err)
			t.Fail()
		}
		// t.Logf("instruction: %v, marshalled: %v\n", i, string(bs))
		var j Instruction
		err = json.Unmarshal(bs, &j)
		if err != nil {
			t.Logf("error unmarshalling: %v", err)
			t.Fail()
		}
		if !i.Equals(j) {
			t.Logf("before: %v, after: %v", i, j)
			t.Fail()
		}
	}
}
